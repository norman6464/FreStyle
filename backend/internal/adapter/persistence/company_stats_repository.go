package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyStatsRepository は [repository.CompanyMemberCounter] の実装。
// users テーブルを company_id で GROUP BY し、会社ごとのメンバー集計を 1 クエリで返す。
type companyStatsRepository struct {
	db *gorm.DB
}

// NewCompanyStatsRepository は会社メンバー集計 repository を生成する。
func NewCompanyStatsRepository(db *gorm.DB) repository.CompanyMemberCounter {
	return &companyStatsRepository{db: db}
}

func (r *companyStatsRepository) CountMembersByCompany(ctx context.Context) ([]repository.CompanyMemberCount, error) {
	rows := make([]repository.CompanyMemberCount, 0)
	err := r.db.WithContext(ctx).
		Table("users").
		Select(
			"company_id AS company_id, "+
				"COUNT(*) AS total, "+
				"COUNT(*) FILTER (WHERE is_active) AS active, "+
				// trainee 判定は正規化後の正である role_id で行う（FRESTYLE-311）。
				"COUNT(*) FILTER (WHERE role_id = ?) AS trainees",
			domain.RoleIDTrainee,
		).
		Where("company_id IS NOT NULL AND deleted_at IS NULL").
		Group("company_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
