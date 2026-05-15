package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// SessionNoteRepository は session_notes テーブル へ の アクセス を 提供 する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// sessionNoteRepository (GORM)。
type SessionNoteRepository interface {
	FindBySessionID(ctx context.Context, sessionID uint64) (*domain.SessionNote, error)
	Upsert(ctx context.Context, n *domain.SessionNote) error
}
