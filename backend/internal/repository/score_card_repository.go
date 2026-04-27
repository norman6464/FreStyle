package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type ScoreCardRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.ScoreCard, error)
	Create(ctx context.Context, c *domain.ScoreCard) error
}

type scoreCardRepository struct{ db *gorm.DB }

func NewScoreCardRepository(db *gorm.DB) ScoreCardRepository { return &scoreCardRepository{db: db} }

func (r *scoreCardRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.ScoreCard, error) {
	var rows []domain.ScoreCard
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *scoreCardRepository) Create(ctx context.Context, c *domain.ScoreCard) error {
	return r.db.WithContext(ctx).Create(c).Error
}
