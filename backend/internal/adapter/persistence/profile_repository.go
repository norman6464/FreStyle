package persistence

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// profileRepository は [repository.ProfileRepository] の GORM 実装。
type profileRepository struct{ db *gorm.DB }

// NewProfileRepository は GORM ベース の [repository.ProfileRepository] を 返す。
func NewProfileRepository(db *gorm.DB) repository.ProfileRepository {
	return &profileRepository{db: db}
}

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
