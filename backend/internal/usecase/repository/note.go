package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NoteRepository は notes テーブルへのアクセスを提供する。
type NoteRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Note, error)
	// FindByID は所有者本人の note だけを返す。SQL の WHERE で user_id まで絞るため、
	// 他人の note は「存在しない」（domain.ErrNotFound）と区別が付かない。
	// 引数の並びは Delete と同じ (userID, id)。
	FindByID(ctx context.Context, userID, id uint64) (*domain.Note, error)
	Create(ctx context.Context, n *domain.Note) error
	Update(ctx context.Context, n *domain.Note) error
	// Delete は WHERE で user_id を絞り、他人の note を消せないようにする。
	Delete(ctx context.Context, userID, id uint64) error
}
