package persistence

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// masterExerciseRepository は [repository.MasterExerciseRepository] の GORM 実装。
type masterExerciseRepository struct {
	db *gorm.DB
}

// NewMasterExerciseRepository は GORM ベース の [repository.MasterExerciseRepository] を 返す。
func NewMasterExerciseRepository(db *gorm.DB) repository.MasterExerciseRepository {
	return &masterExerciseRepository{db: db}
}

func (r *masterExerciseRepository) ListByLanguage(language string) ([]domain.MasterExercise, error) {
	var exercises []domain.MasterExercise
	q := r.db.Where("is_published = ?", true)
	if language != "" {
		q = q.Where("language = ?", language)
	}
	if err := q.Order("order_index asc").Find(&exercises).Error; err != nil {
		return nil, err
	}
	return exercises, nil
}

func (r *masterExerciseRepository) GetByID(id uint64) (*domain.MasterExercise, error) {
	var exercise domain.MasterExercise
	if err := r.db.First(&exercise, id).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *masterExerciseRepository) GetBySlug(slug string) (*domain.MasterExercise, error) {
	var exercise domain.MasterExercise
	if err := r.db.Where("slug = ?", slug).First(&exercise).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}
