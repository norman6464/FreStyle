package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyRepository は [repository.CompanyRepository] の GORM 実装。
type companyRepository struct{ db *gorm.DB }

// NewCompanyRepository は GORM ベース の [repository.CompanyRepository] を 返す。
func NewCompanyRepository(db *gorm.DB) repository.CompanyRepository {
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
