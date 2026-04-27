package repository

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type DailyGoalRepository interface {
	FindByUserAndDate(ctx context.Context, userID uint64, date time.Time) (*domain.DailyGoal, error)
	Upsert(ctx context.Context, g *domain.DailyGoal) error
}

type dailyGoalRepository struct{ db *gorm.DB }

func NewDailyGoalRepository(db *gorm.DB) DailyGoalRepository { return &dailyGoalRepository{db: db} }

func (r *dailyGoalRepository) FindByUserAndDate(ctx context.Context, userID uint64, date time.Time) (*domain.DailyGoal, error) {
	var g domain.DailyGoal
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND goal_date = ?", userID, date.Format("2006-01-02")).
		First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *dailyGoalRepository) Upsert(ctx context.Context, g *domain.DailyGoal) error {
	return r.db.WithContext(ctx).Save(g).Error
}
