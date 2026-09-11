package repository

import (
	"context"
	"errors"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// ErrCommentThreadNotFound はスレッドが存在しない・別ページ/別テナントのものだったときに返す。
var ErrCommentThreadNotFound = errors.New("comment thread not found")

// CommentAnchor はスレッドが錨付けされている、ページ内の特定ブロック・特定文字範囲。
// 4 つとも nil なら page-level（PR2 までの書き込み経路と同じ）を意味する。
// 4 つとも非 nil のときだけ意味を持つ組み合わせで、その検証は domain.ValidateCommentAnchor
// （usecase 経由）が行う。
type CommentAnchor struct {
	BlockID    *string
	AnchorFrom *int
	AnchorTo   *int
	Quote      *string
}

// CommentRepository は comment_threads / comments テーブルへのアクセスを提供する。
type CommentRepository interface {
	CreateCommentThread(ctx context.Context, workspaceID, pageID string, createdByUserID uint64, anchor CommentAnchor) (*domain.CommentThread, error)
	CreateComment(ctx context.Context, threadID string, authorUserID uint64, body string) (*domain.Comment, error)
	// BlockExistsInPage は block_id が、実際にその workspace/page に属するブロックとして
	// 実在するかを返す。錨付きコメントの CreateCommentThread が block_id を受け取る前に、
	// usecase から呼ぶ（他ページ・他テナントのブロックへ錨を張れてしまう穴を塞ぐ）。
	BlockExistsInPage(ctx context.Context, workspaceID, pageID, blockID string) (bool, error)
	// GetCommentThread は workspace_id/page_id で絞って 1 件返す。無ければ ErrCommentThreadNotFound。
	GetCommentThread(ctx context.Context, workspaceID, pageID, threadID string) (*domain.CommentThread, error)
	// ListCommentThreadsByPage はページのスレッド一覧を作成日時昇順で返す（0件なら空スライス）。
	ListCommentThreadsByPage(ctx context.Context, workspaceID, pageID string) ([]domain.CommentThread, error)
	// ListCommentsByThreads は複数スレッドぶんの返信をまとめて返す（N+1回避。作成日時昇順）。
	ListCommentsByThreads(ctx context.Context, threadIDs []string) ([]domain.Comment, error)
	// ResolveCommentThread / ReopenCommentThread は workspace_id/page_id/id で絞って更新する。
	// 対象が無ければ ErrCommentThreadNotFound。
	ResolveCommentThread(ctx context.Context, workspaceID, pageID, threadID string, resolvedByUserID uint64) (*domain.CommentThread, error)
	ReopenCommentThread(ctx context.Context, workspaceID, pageID, threadID string) (*domain.CommentThread, error)
}
