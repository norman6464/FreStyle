package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NoteRepository は notes テーブルへのアクセスを提供する。
type NoteRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Note, error)
	FindByID(ctx context.Context, id uint64) (*domain.Note, error)
	Create(ctx context.Context, n *domain.Note) error
	Update(ctx context.Context, n *domain.Note) error
	// Delete は WHERE で user_id を絞り、他人の note を消せないようにする。
	Delete(ctx context.Context, userID, id uint64) error
}
