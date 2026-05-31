package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// masterExerciseRepository は [repository.MasterExerciseRepository] の GORM 実装。
type masterExerciseRepository struct {
	db *gorm.DB
}

func NewMasterExerciseRepository(db *gorm.DB) repository.MasterExerciseRepository {
	return &masterExerciseRepository{db: db}
}

func (r *masterExerciseRepository) ListByLanguage(ctx context.Context, language string) ([]domain.MasterExercise, error) {
	var exercises []domain.MasterExercise
	q := r.db.WithContext(ctx).Where("is_published = ?", true)
	if language != "" {
		q = q.Where("language = ?", language)
	}
	if err := q.Order("order_index asc").Find(&exercises).Error; err != nil {
		return nil, err
	}
	return exercises, nil
}

func (r *masterExerciseRepository) GetByID(ctx context.Context, id uint64) (*domain.MasterExercise, error) {
	var exercise domain.MasterExercise
	if err := r.db.WithContext(ctx).First(&exercise, id).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *masterExerciseRepository) GetBySlug(ctx context.Context, slug string) (*domain.MasterExercise, error) {
	var exercise domain.MasterExercise
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&exercise).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}
