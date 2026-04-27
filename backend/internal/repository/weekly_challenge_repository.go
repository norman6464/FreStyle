package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type WeeklyChallengeRepository interface {
	CurrentChallenge(ctx context.Context) (*domain.WeeklyChallenge, error)
	UpsertProgress(ctx context.Context, p *domain.WeeklyChallengeProgress) error
}

type weeklyChallengeRepository struct{ db *gorm.DB }

func NewWeeklyChallengeRepository(db *gorm.DB) WeeklyChallengeRepository {
	return &weeklyChallengeRepository{db: db}
}

func (r *weeklyChallengeRepository) CurrentChallenge(ctx context.Context) (*domain.WeeklyChallenge, error) {
	var c domain.WeeklyChallenge
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("week_start DESC").
		First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *weeklyChallengeRepository) UpsertProgress(ctx context.Context, p *domain.WeeklyChallengeProgress) error {
	return r.db.WithContext(ctx).Save(p).Error
}
