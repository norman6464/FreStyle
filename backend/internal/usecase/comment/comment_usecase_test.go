package comment_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/comment"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	cWS       = "0198a000-0000-7000-8000-000000000001"
	cPage     = "0198a000-0000-7000-8000-000000000002"
	cThread   = "0198a000-0000-7000-8000-000000000003"
	cAuthor   = uint64(42)
	validBody = `[{"type":"text","text":"hello"}]`
)

// --- CreateCommentThreadUseCase ---

func Test_スレッド作成_成功時にスレッドとコメントが返る(t *testing.T) {
	repo := &mockCommentRepo{}
	thread := &domain.CommentThread{ID: cThread, WorkspaceID: cWS, PageID: cPage, CreatedByUserID: cAuthor}
	c := &domain.Comment{ID: "comment-1", ThreadID: cThread, AuthorUserID: cAuthor, Body: validBody}
	repo.On("CreateCommentThread", mock.Anything, cWS, cPage, cAuthor).Return(thread, nil)
	repo.On("CreateComment", mock.Anything, cThread, cAuthor, validBody).Return(c, nil)
	uc := comment.NewCreateCommentThreadUseCase(repo, &fakeTxManager{})

	out, err := uc.Execute(context.Background(), comment.CreateCommentThreadInput{
		WorkspaceID: cWS, PageID: cPage, AuthorUserID: cAuthor, Body: validBody,
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, *thread, out.Thread)
	assert.Equal(t, *c, out.Comment)
	repo.AssertExpectations(t)
}

// DoInTx が 1 回だけ呼ばれ、スレッド作成 → コメント作成の両方がその中（tx の中）で
// 行われることを固定する。片方だけ tx の外で呼ばれる回帰（発言の無いスレッドが残る）を
// ここで捕まえる。page_usecase_external_test.go の fakeTxManager・inTx パターンを踏襲。
func Test_スレッド作成_DoInTxが1回だけ呼ばれ両方がtxの中で行われる(t *testing.T) {
	repo := &mockCommentRepo{}
	thread := &domain.CommentThread{ID: cThread, WorkspaceID: cWS, PageID: cPage, CreatedByUserID: cAuthor}
	c := &domain.Comment{ID: "comment-1", ThreadID: cThread, AuthorUserID: cAuthor, Body: validBody}
	repo.On("CreateCommentThread", mock.MatchedBy(inTx), cWS, cPage, cAuthor).Return(thread, nil)
	repo.On("CreateComment", mock.MatchedBy(inTx), cThread, cAuthor, validBody).Return(c, nil)
	tx := &fakeTxManager{}
	uc := comment.NewCreateCommentThreadUseCase(repo, tx)

	_, err := uc.Execute(context.Background(), comment.CreateCommentThreadInput{
		WorkspaceID: cWS, PageID: cPage, AuthorUserID: cAuthor, Body: validBody,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, tx.calls, "DoInTx は 1 回だけ（スレッド作成とコメント作成を 1 つの単位にまとめる）")
	repo.AssertExpectations(t)
}

// invalidCommentBodies は境界値・異常系の不正な本文の一覧。スレッド作成・返信の
// 両方で同じ表を使う（CodeRabbit 指摘: 個別に [] や `not json` だけを見ていて
// {}・null・空入力等の境界値が無かった）。
var invalidCommentBodies = []struct {
	name string
	body string
}{
	{name: "空配列（本文の無いコメント）", body: `[]`},
	{name: "不正なJSON", body: `not json`},
	{name: "空文字", body: ``},
	{name: "配列でないJSON（object）", body: `{"type":"text"}`},
	{name: "null要素を含む配列", body: `[null]`},
	{name: "typeの無いobject要素", body: `[{}]`},
}

func Test_スレッド作成_本文が不正ならrepoを一切呼ばずに拒否(t *testing.T) {
	for _, tc := range invalidCommentBodies {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCommentRepo{}
			uc := comment.NewCreateCommentThreadUseCase(repo, &fakeTxManager{})

			_, err := uc.Execute(context.Background(), comment.CreateCommentThreadInput{
				WorkspaceID: cWS, PageID: cPage, AuthorUserID: cAuthor, Body: tc.body,
			})

			require.ErrorIs(t, err, domain.ErrInvalidCommentBody)
			repo.AssertNotCalled(t, "CreateCommentThread", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			repo.AssertNotCalled(t, "CreateComment", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

// --- AddCommentUseCase ---

func Test_返信_成功(t *testing.T) {
	repo := &mockCommentRepo{}
	repo.On("GetCommentThread", mock.Anything, cWS, cPage, cThread).
		Return(&domain.CommentThread{ID: cThread, WorkspaceID: cWS, PageID: cPage}, nil)
	c := &domain.Comment{ID: "comment-2", ThreadID: cThread, AuthorUserID: cAuthor, Body: validBody}
	repo.On("CreateComment", mock.Anything, cThread, cAuthor, validBody).Return(c, nil)
	uc := comment.NewAddCommentUseCase(repo)

	out, err := uc.Execute(context.Background(), comment.AddCommentInput{
		WorkspaceID: cWS, PageID: cPage, ThreadID: cThread, AuthorUserID: cAuthor, Body: validBody,
	})

	require.NoError(t, err)
	assert.Equal(t, *c, *out)
	repo.AssertExpectations(t)
}

func Test_返信_スレッドが見つからなければ伝播しCreateCommentを呼ばない(t *testing.T) {
	repo := &mockCommentRepo{}
	repo.On("GetCommentThread", mock.Anything, cWS, cPage, cThread).
		Return(nil, repository.ErrCommentThreadNotFound)
	uc := comment.NewAddCommentUseCase(repo)

	_, err := uc.Execute(context.Background(), comment.AddCommentInput{
		WorkspaceID: cWS, PageID: cPage, ThreadID: cThread, AuthorUserID: cAuthor, Body: validBody,
	})

	require.ErrorIs(t, err, repository.ErrCommentThreadNotFound)
	repo.AssertNotCalled(t, "CreateComment", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_返信_本文が不正ならスレッド確認すら行わずに拒否(t *testing.T) {
	for _, tc := range invalidCommentBodies {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockCommentRepo{}
			uc := comment.NewAddCommentUseCase(repo)

			_, err := uc.Execute(context.Background(), comment.AddCommentInput{
				WorkspaceID: cWS, PageID: cPage, ThreadID: cThread, AuthorUserID: cAuthor, Body: tc.body,
			})

			require.ErrorIs(t, err, domain.ErrInvalidCommentBody)
			repo.AssertNotCalled(t, "GetCommentThread", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			repo.AssertNotCalled(t, "CreateComment", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

// --- ListCommentThreadsUseCase ---

func Test_スレッド一覧_0件なら空スライス(t *testing.T) {
	repo := &mockCommentRepo{}
	repo.On("ListCommentThreadsByPage", mock.Anything, cWS, cPage).Return([]domain.CommentThread{}, nil)
	uc := comment.NewListCommentThreadsUseCase(repo)

	out, err := uc.Execute(context.Background(), comment.ListCommentThreadsInput{WorkspaceID: cWS, PageID: cPage})

	require.NoError(t, err)
	assert.Empty(t, out)
	repo.AssertNotCalled(t, "ListCommentsByThreads", mock.Anything, mock.Anything)
}

func Test_スレッド一覧_複数スレッド複数コメントを正しくグルーピングする(t *testing.T) {
	repo := &mockCommentRepo{}
	now := time.Now()
	threadA := domain.CommentThread{ID: "thread-a", WorkspaceID: cWS, PageID: cPage, CreatedAt: now}
	threadB := domain.CommentThread{ID: "thread-b", WorkspaceID: cWS, PageID: cPage, CreatedAt: now.Add(time.Second)}
	repo.On("ListCommentThreadsByPage", mock.Anything, cWS, cPage).
		Return([]domain.CommentThread{threadA, threadB}, nil)
	commentA1 := domain.Comment{ID: "a1", ThreadID: "thread-a"}
	commentA2 := domain.Comment{ID: "a2", ThreadID: "thread-a"}
	commentB1 := domain.Comment{ID: "b1", ThreadID: "thread-b"}
	repo.On("ListCommentsByThreads", mock.Anything, []string{"thread-a", "thread-b"}).
		Return([]domain.Comment{commentA1, commentA2, commentB1}, nil)
	uc := comment.NewListCommentThreadsUseCase(repo)

	out, err := uc.Execute(context.Background(), comment.ListCommentThreadsInput{WorkspaceID: cWS, PageID: cPage})

	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, threadA, out[0].Thread)
	assert.Equal(t, []domain.Comment{commentA1, commentA2}, out[0].Comments)
	assert.Equal(t, threadB, out[1].Thread)
	assert.Equal(t, []domain.Comment{commentB1}, out[1].Comments)
	repo.AssertExpectations(t)
}

// --- ResolveCommentThreadUseCase / ReopenCommentThreadUseCase ---

func Test_スレッド解決(t *testing.T) {
	repo := &mockCommentRepo{}
	resolvedByUserID := cAuthor
	want := &domain.CommentThread{ID: cThread, WorkspaceID: cWS, PageID: cPage, ResolvedByUserID: &resolvedByUserID}
	repo.On("ResolveCommentThread", mock.Anything, cWS, cPage, cThread, cAuthor).Return(want, nil)
	uc := comment.NewResolveCommentThreadUseCase(repo)

	out, err := uc.Execute(context.Background(), comment.ResolveCommentThreadInput{
		WorkspaceID: cWS, PageID: cPage, ThreadID: cThread, ResolvedByUserID: cAuthor,
	})

	require.NoError(t, err)
	assert.Equal(t, want, out)
	repo.AssertExpectations(t)
}

func Test_スレッド解決_見つからなければ伝播(t *testing.T) {
	repo := &mockCommentRepo{}
	repo.On("ResolveCommentThread", mock.Anything, cWS, cPage, cThread, cAuthor).
		Return(nil, repository.ErrCommentThreadNotFound)
	uc := comment.NewResolveCommentThreadUseCase(repo)

	_, err := uc.Execute(context.Background(), comment.ResolveCommentThreadInput{
		WorkspaceID: cWS, PageID: cPage, ThreadID: cThread, ResolvedByUserID: cAuthor,
	})

	require.ErrorIs(t, err, repository.ErrCommentThreadNotFound)
}

func Test_スレッド再開(t *testing.T) {
	repo := &mockCommentRepo{}
	want := &domain.CommentThread{ID: cThread, WorkspaceID: cWS, PageID: cPage}
	repo.On("ReopenCommentThread", mock.Anything, cWS, cPage, cThread).Return(want, nil)
	uc := comment.NewReopenCommentThreadUseCase(repo)

	out, err := uc.Execute(context.Background(), comment.ReopenCommentThreadInput{
		WorkspaceID: cWS, PageID: cPage, ThreadID: cThread,
	})

	require.NoError(t, err)
	assert.Equal(t, want, out)
	repo.AssertExpectations(t)
}

func Test_スレッド再開_見つからなければ伝播(t *testing.T) {
	repo := &mockCommentRepo{}
	repo.On("ReopenCommentThread", mock.Anything, cWS, cPage, cThread).
		Return(nil, repository.ErrCommentThreadNotFound)
	uc := comment.NewReopenCommentThreadUseCase(repo)

	_, err := uc.Execute(context.Background(), comment.ReopenCommentThreadInput{
		WorkspaceID: cWS, PageID: cPage, ThreadID: cThread,
	})

	require.ErrorIs(t, err, repository.ErrCommentThreadNotFound)
}
