package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	ListAll(ctx context.Context) ([]domain.Company, error)
	FindByID(ctx context.Context, id uint64) (*domain.Company, error)
}

type companyRepository struct{ db *gorm.DB }

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) ListAll(ctx context.Context) ([]domain.Company, error) {
	var rows []domain.Company
	err := r.db.WithContext(ctx).Order("name ASC").Find(&rows).Error
	return rows, err
}

func (r *companyRepository) FindByID(ctx context.Context, id uint64) (*domain.Company, error) {
	var c domain.Company
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
