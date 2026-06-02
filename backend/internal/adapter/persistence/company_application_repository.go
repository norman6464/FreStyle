package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyApplicationRepository は [repository.CompanyApplicationRepository] の GORM 実装。
type companyApplicationRepository struct{ db *gorm.DB }

func NewCompanyApplicationRepository(db *gorm.DB) repository.CompanyApplicationRepository {
	return &companyApplicationRepository{db: db}
}

func (r *companyApplicationRepository) Create(ctx context.Context, app *domain.CompanyApplication) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *companyApplicationRepository) ListAll(ctx context.Context) ([]domain.CompanyApplication, error) {
	var rows []domain.CompanyApplication
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *companyApplicationRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).
		Model(&domain.CompanyApplication{}).
		Where("id = ?", id).
		Update("status", status).Error
}
