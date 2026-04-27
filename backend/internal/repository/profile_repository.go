package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (*domain.Profile, error)
	Upsert(ctx context.Context, p *domain.Profile) error
}

type profileRepository struct{ db *gorm.DB }

func NewProfileRepository(db *gorm.DB) ProfileRepository { return &profileRepository{db: db} }

func (r *profileRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.Profile, error) {
	var p domain.Profile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *profileRepository) Upsert(ctx context.Context, p *domain.Profile) error {
	return r.db.WithContext(ctx).Save(p).Error
}
