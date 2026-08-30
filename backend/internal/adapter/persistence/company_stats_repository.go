package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// companyStatsRepository は [repository.CompanyMemberCounter] の実装。
// users テーブルを company_id で GROUP BY し、会社ごとのメンバー集計を 1 クエリで返す。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type companyStatsRepository struct {
	db *sql.DB
}

// NewCompanyStatsRepository は会社メンバー集計 repository を生成する。
func NewCompanyStatsRepository(db *sql.DB) repository.CompanyMemberCounter {
	return &companyStatsRepository{db: db}
}

func (r *companyStatsRepository) CountMembersByCompany(ctx context.Context) ([]repository.CompanyMemberCount, error) {
	rows, err := sqlcgen.New(r.db).CountMembersByCompany(ctx, domain.RoleIDTrainee)
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

func (r *companyStatsRepository) CountMembersByWorkspace(ctx context.Context) ([]repository.WorkspaceMemberCount, error) {
	rows, err := sqlcgen.New(r.db).CountMembersByWorkspace(ctx, domain.RoleIDTrainee)
	if err != nil {
		return nil, err
	}
	out := make([]repository.WorkspaceMemberCount, 0, len(rows))
	for _, row := range rows {
		if !row.WorkspaceID.Valid {
			continue // WHERE workspace_id IS NOT NULL 済みだが、型上は NullUUID なので念のため。
		}
		out = append(out, repository.WorkspaceMemberCount{
			WorkspaceID: row.WorkspaceID.UUID.String(),
			Total:       int(row.Total),
			Active:      int(row.Active),
			Trainees:    int(row.Trainees),
		})
	}
	return out, nil
}
