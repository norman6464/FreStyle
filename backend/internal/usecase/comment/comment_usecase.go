// Package comment はページ全体へのコメント（FRESTYLE-432 段 2）の usecase を持つ。
//
// 錨付け（block_id / anchor_from / anchor_to / quote）の書き込み経路は段 3 の範囲で、
// このパッケージの usecase はすべて「本文」だけを受け取る。認可（CanComment / CanView）は
// ここでは判定しない — handler が kb.CheckPagePermissionUseCase を先に通す
// （FreStyle のクリーンアーキテクチャ規約: usecase は handler を知らない。
// comment パッケージは他の usecase サブパッケージ（kb 等）を import しない）。
package comment

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CreateCommentThreadUseCase はページに新しいスレッドを立て、最初の発言を同時に作る。
// スレッドだけを本文無しで作ることはできない（「最初の発言」が無いスレッドは
// 一覧に出しても中身が空になり、UI 上「誰が何のスレッドを立てたか」を示せない）。
type CreateCommentThreadUseCase struct {
	repo      repository.CommentRepository
	txManager repository.TxManager
}

func NewCreateCommentThreadUseCase(repo repository.CommentRepository, txManager repository.TxManager) *CreateCommentThreadUseCase {
	return &CreateCommentThreadUseCase{repo: repo, txManager: txManager}
}

type CreateCommentThreadInput struct {
	WorkspaceID  string
	PageID       string
	AuthorUserID uint64
	Body         string
}

// CreateCommentThreadOutput はスレッドと、その最初の発言（コメント）の組。
type CreateCommentThreadOutput struct {
	Thread  domain.CommentThread
	Comment domain.Comment
}

func (u *CreateCommentThreadUseCase) Execute(ctx context.Context, in CreateCommentThreadInput) (*CreateCommentThreadOutput, error) {
	// Body の検証は先に行う。不正な本文のためだけにトランザクションを開いて
	// スレッドだけ作ってしまう（＝発言の無いスレッドが残る）事態を避ける。
	if err := domain.ValidateCommentBody(in.Body); err != nil {
		return nil, err
	}
	var out CreateCommentThreadOutput
	// スレッド作成 → 最初の発言作成は 1 つのトランザクションに入れる。
	// 片方だけ成功すると「発言の無いスレッド」または「存在しないスレッドを指す発言」という
	// 中間状態が残ってしまうため。
	err := u.txManager.DoInTx(ctx, func(ctx context.Context) error {
		thread, err := u.repo.CreateCommentThread(ctx, in.WorkspaceID, in.PageID, in.AuthorUserID)
		if err != nil {
			return err
		}
		c, err := u.repo.CreateComment(ctx, thread.ID, in.AuthorUserID, in.Body)
		if err != nil {
			return err
		}
		out = CreateCommentThreadOutput{Thread: *thread, Comment: *c}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AddCommentUseCase は既存のスレッドへ返信を 1 件足す。
type AddCommentUseCase struct {
	repo repository.CommentRepository
}

func NewAddCommentUseCase(repo repository.CommentRepository) *AddCommentUseCase {
	return &AddCommentUseCase{repo: repo}
}

type AddCommentInput struct {
	WorkspaceID  string
	PageID       string
	ThreadID     string
	AuthorUserID uint64
	Body         string
}

func (u *AddCommentUseCase) Execute(ctx context.Context, in AddCommentInput) (*domain.Comment, error) {
	if err := domain.ValidateCommentBody(in.Body); err != nil {
		return nil, err
	}
	// スレッドの実在確認は workspace_id / page_id まで絞って行う。他ページ・他テナントの
	// thread_id を渡された場合、GetCommentThread が repository.ErrCommentThreadNotFound を
	// 返すのでそのまま伝播させる（別ページの thread_id へ返信を生やせてしまう穴を塞ぐ）。
	if _, err := u.repo.GetCommentThread(ctx, in.WorkspaceID, in.PageID, in.ThreadID); err != nil {
		return nil, err
	}
	return u.repo.CreateComment(ctx, in.ThreadID, in.AuthorUserID, in.Body)
}

// CommentThreadWithComments は 1 スレッドと、それに紐づく発言（最初の発言 + 返信）の組。
type CommentThreadWithComments struct {
	Thread   domain.CommentThread
	Comments []domain.Comment
}

// ListCommentThreadsUseCase はページのスレッド一覧を、それぞれの発言付きで返す。
type ListCommentThreadsUseCase struct {
	repo repository.CommentRepository
}

func NewListCommentThreadsUseCase(repo repository.CommentRepository) *ListCommentThreadsUseCase {
	return &ListCommentThreadsUseCase{repo: repo}
}

type ListCommentThreadsInput struct {
	WorkspaceID string
	PageID      string
}

func (u *ListCommentThreadsUseCase) Execute(ctx context.Context, in ListCommentThreadsInput) ([]CommentThreadWithComments, error) {
	threads, err := u.repo.ListCommentThreadsByPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if len(threads) == 0 {
		return []CommentThreadWithComments{}, nil
	}
	// 発言はスレッドごとに引かず、全スレッド分を 1 回でまとめて取る（N+1 回避）。
	ids := make([]string, 0, len(threads))
	for _, t := range threads {
		ids = append(ids, t.ID)
	}
	comments, err := u.repo.ListCommentsByThreads(ctx, ids)
	if err != nil {
		return nil, err
	}
	byThread := make(map[string][]domain.Comment, len(threads))
	for _, c := range comments {
		byThread[c.ThreadID] = append(byThread[c.ThreadID], c)
	}
	out := make([]CommentThreadWithComments, 0, len(threads))
	for _, t := range threads {
		out = append(out, CommentThreadWithComments{Thread: t, Comments: byThread[t.ID]})
	}
	return out, nil
}

// ResolveCommentThreadUseCase はスレッドを解決済みにする。
type ResolveCommentThreadUseCase struct {
	repo repository.CommentRepository
}

func NewResolveCommentThreadUseCase(repo repository.CommentRepository) *ResolveCommentThreadUseCase {
	return &ResolveCommentThreadUseCase{repo: repo}
}

type ResolveCommentThreadInput struct {
	WorkspaceID      string
	PageID           string
	ThreadID         string
	ResolvedByUserID uint64
}

func (u *ResolveCommentThreadUseCase) Execute(ctx context.Context, in ResolveCommentThreadInput) (*domain.CommentThread, error) {
	return u.repo.ResolveCommentThread(ctx, in.WorkspaceID, in.PageID, in.ThreadID, in.ResolvedByUserID)
}

// ReopenCommentThreadUseCase は解決済みのスレッドを未解決へ戻す。
type ReopenCommentThreadUseCase struct {
	repo repository.CommentRepository
}

func NewReopenCommentThreadUseCase(repo repository.CommentRepository) *ReopenCommentThreadUseCase {
	return &ReopenCommentThreadUseCase{repo: repo}
}

type ReopenCommentThreadInput struct {
	WorkspaceID string
	PageID      string
	ThreadID    string
}

func (u *ReopenCommentThreadUseCase) Execute(ctx context.Context, in ReopenCommentThreadInput) (*domain.CommentThread, error) {
	return u.repo.ReopenCommentThread(ctx, in.WorkspaceID, in.PageID, in.ThreadID)
}
