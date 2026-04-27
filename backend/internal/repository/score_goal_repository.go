package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type ScoreGoalRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (*domain.ScoreGoal, error)
	Upsert(ctx context.Context, g *domain.ScoreGoal) error
}

type scoreGoalRepository struct{ db *gorm.DB }

func NewScoreGoalRepository(db *gorm.DB) ScoreGoalRepository { return &scoreGoalRepository{db: db} }

func (r *scoreGoalRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.ScoreGoal, error) {
	var g domain.ScoreGoal
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *scoreGoalRepository) Upsert(ctx context.Context, g *domain.ScoreGoal) error {
	return r.db.WithContext(ctx).Save(g).Error
}
