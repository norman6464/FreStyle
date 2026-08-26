package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyStatsRepository は [repository.CompanyMemberCounter] の実装。
// users テーブルを company_id で GROUP BY し、会社ごとのメンバー集計を 1 クエリで返す。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type companyStatsRepository struct {
	db *gorm.DB
}

// NewCompanyStatsRepository は会社メンバー集計 repository を生成する。
func NewCompanyStatsRepository(db *gorm.DB) repository.CompanyMemberCounter {
	return &companyStatsRepository{db: db}
}

func (r *companyStatsRepository) CountMembersByCompany(ctx context.Context) ([]repository.CompanyMemberCount, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).CountMembersByCompany(ctx, int16(domain.RoleIDTrainee))
	if err != nil {
		return nil, err
	}
	out := make([]repository.CompanyMemberCount, 0, len(rows))
	for _, row := range rows {
		out = append(out, repository.CompanyMemberCount{
			CompanyID: uint64(row.CompanyID),
			Total:     int(row.Total),
			Active:    int(row.Active),
			Trainees:  int(row.Trainees),
		})
	}
	return out, nil
}
