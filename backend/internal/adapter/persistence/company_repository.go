package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyRepository は [repository.CompanyRepository] の GORM 実装。
type companyRepository struct{ db *gorm.DB }

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

// UpdateAiChatEnabled は ai_chat_enabled_for_trainees を更新する（生 SQL 直書き / updated_at も更新）。
func (r *companyRepository) UpdateAiChatEnabled(ctx context.Context, companyID uint64, enabled bool) error {
	const q = `UPDATE companies SET ai_chat_enabled_for_trainees = ?, updated_at = NOW() WHERE id = ?`
	return r.db.WithContext(ctx).Exec(q, enabled, companyID).Error
}
