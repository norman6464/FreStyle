package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrCommentThreadNotFound はスレッドが存在しない・別ページ/別テナントのものだったときに返す。
var ErrCommentThreadNotFound = errors.New("comment thread not found")

// CommentRepository は comment_threads / comments テーブルへのアクセスを提供する。
type CommentRepository interface {
	CreateCommentThread(ctx context.Context, workspaceID, pageID string, createdByUserID uint64) (*domain.CommentThread, error)
	CreateComment(ctx context.Context, threadID string, authorUserID uint64, body string) (*domain.Comment, error)
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
