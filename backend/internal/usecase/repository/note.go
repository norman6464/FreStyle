package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NoteRepository は notes テーブル へ の アクセス を 提供 する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// noteRepository (GORM)。
type NoteRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Note, error)
	FindByID(ctx context.Context, id uint64) (*domain.Note, error)
	Create(ctx context.Context, n *domain.Note) error
	Update(ctx context.Context, n *domain.Note) error
	// Delete は所有者検証込みで note を削除する。
	// WHERE で user_id を絞って他人の note を消せないようにする。
	Delete(ctx context.Context, userID, id uint64) error
}
