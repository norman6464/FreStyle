package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// SessionNoteRepository は session_notes テーブルへのアクセスを提供する。
type SessionNoteRepository interface {
	FindBySessionID(ctx context.Context, sessionID uint64) (*domain.SessionNote, error)
	Upsert(ctx context.Context, n *domain.SessionNote) error
}
