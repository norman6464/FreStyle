package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type ScenarioBookmarkRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.ScenarioBookmark, error)
	Create(ctx context.Context, b *domain.ScenarioBookmark) error
	Delete(ctx context.Context, userID, scenarioID uint64) error
}

type scenarioBookmarkRepository struct{ db *gorm.DB }

func NewScenarioBookmarkRepository(db *gorm.DB) ScenarioBookmarkRepository {
	return &scenarioBookmarkRepository{db: db}
}

func (r *scenarioBookmarkRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.ScenarioBookmark, error) {
	var rows []domain.ScenarioBookmark
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *scenarioBookmarkRepository) Create(ctx context.Context, b *domain.ScenarioBookmark) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *scenarioBookmarkRepository) Delete(ctx context.Context, userID, scenarioID uint64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND scenario_id = ?", userID, scenarioID).
		Delete(&domain.ScenarioBookmark{}).Error
}
