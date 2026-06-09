package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyApplicationRepository は [repository.CompanyApplicationRepository] の実装。
// 読み取り（ListAll）は sqlc 生成コード（生 SQL）、書き込みは GORM。
type companyApplicationRepository struct{ db *gorm.DB }

func NewCompanyApplicationRepository(db *gorm.DB) repository.CompanyApplicationRepository {
	return &companyApplicationRepository{db: db}
}

func (r *companyApplicationRepository) Create(ctx context.Context, app *domain.CompanyApplication) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *companyApplicationRepository) ListAll(ctx context.Context) ([]domain.CompanyApplication, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListCompanyApplications(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.CompanyApplication, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.CompanyApplication{
			ID:            uint64(row.ID),
			CompanyName:   row.CompanyName,
			ApplicantName: row.ApplicantName,
			Email:         row.Email,
			Message:       row.Message,
			Status:        row.Status,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *companyApplicationRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).
		Model(&domain.CompanyApplication{}).
		Where("id = ?", id).
		Update("status", status).Error
}
