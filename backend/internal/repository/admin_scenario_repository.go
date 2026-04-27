package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// AdminScenarioRepository は管理者向けにシナリオを CRUD する。
// PracticeScenario 構造体を再利用するが、書き込み権限を分離するため別 interface とする。
type AdminScenarioRepository interface {
	List(ctx context.Context) ([]domain.PracticeScenario, error)
	Create(ctx context.Context, s *domain.PracticeScenario) error
	Update(ctx context.Context, s *domain.PracticeScenario) error
	Delete(ctx context.Context, id uint64) error
}

type adminScenarioRepository struct{ db *gorm.DB }

func NewAdminScenarioRepository(db *gorm.DB) AdminScenarioRepository {
	return &adminScenarioRepository{db: db}
}

func (r *adminScenarioRepository) List(ctx context.Context) ([]domain.PracticeScenario, error) {
	var rows []domain.PracticeScenario
	err := r.db.WithContext(ctx).Order("difficulty_level, id").Find(&rows).Error
	return rows, err
}

func (r *adminScenarioRepository) Create(ctx context.Context, s *domain.PracticeScenario) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *adminScenarioRepository) Update(ctx context.Context, s *domain.PracticeScenario) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *adminScenarioRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.PracticeScenario{}, id).Error
}
