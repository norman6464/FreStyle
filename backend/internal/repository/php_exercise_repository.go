package repository

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// PhpExerciseRepository は PHP 演習問題の永続化を担う。
type PhpExerciseRepository interface {
	List() ([]domain.PhpExercise, error)
	GetByID(id uint) (*domain.PhpExercise, error)
}

type phpExerciseRepository struct {
	db *gorm.DB
}

func NewPhpExerciseRepository(db *gorm.DB) PhpExerciseRepository {
	return &phpExerciseRepository{db: db}
}

func (r *phpExerciseRepository) List() ([]domain.PhpExercise, error) {
	var exercises []domain.PhpExercise
	if err := r.db.Order("order_index asc").Find(&exercises).Error; err != nil {
		return nil, err
	}
	return exercises, nil
}

func (r *phpExerciseRepository) GetByID(id uint) (*domain.PhpExercise, error) {
	var exercise domain.PhpExercise
	if err := r.db.First(&exercise, id).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}
