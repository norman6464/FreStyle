package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type FriendshipRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Friendship, error)
	Create(ctx context.Context, f *domain.Friendship) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
}

type friendshipRepository struct{ db *gorm.DB }

func NewFriendshipRepository(db *gorm.DB) FriendshipRepository { return &friendshipRepository{db: db} }

func (r *friendshipRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.Friendship, error) {
	var rows []domain.Friendship
	err := r.db.WithContext(ctx).
		Where("requester_id = ? OR addressee_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *friendshipRepository) Create(ctx context.Context, f *domain.Friendship) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *friendshipRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&domain.Friendship{}).Where("id = ?", id).Update("status", status).Error
}
