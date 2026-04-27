package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type PracticeScenarioRepository interface {
	ListActive(ctx context.Context) ([]domain.PracticeScenario, error)
	FindByID(ctx context.Context, id uint64) (*domain.PracticeScenario, error)
}

type practiceScenarioRepository struct{ db *gorm.DB }

func NewPracticeScenarioRepository(db *gorm.DB) PracticeScenarioRepository {
	return &practiceScenarioRepository{db: db}
}

func (r *practiceScenarioRepository) ListActive(ctx context.Context) ([]domain.PracticeScenario, error) {
	var rows []domain.PracticeScenario
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("difficulty_level, id").
		Find(&rows).Error
	return rows, err
}

func (r *practiceScenarioRepository) FindByID(ctx context.Context, id uint64) (*domain.PracticeScenario, error) {
	var s domain.PracticeScenario
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
