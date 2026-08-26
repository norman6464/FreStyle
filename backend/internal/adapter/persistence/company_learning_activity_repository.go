package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// companyLearningActivityRepository は [repository.CompanyLearningActivitySummarizer] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type companyLearningActivityRepository struct {
	db *sql.DB
}

// NewCompanyLearningActivityRepository は CompanyLearningActivitySummarizer の実装を返す。
func NewCompanyLearningActivityRepository(db *sql.DB) repository.CompanyLearningActivitySummarizer {
	return &companyLearningActivityRepository{db: db}
}

// ListMemberActivities は自社 trainee ごとの最終活動日と fromDate 以降の活動回数を 1 クエリで集計する
// (trainee ごとの個別クエリだと N+1 になるため集計 CTE + LEFT JOIN で一括取得する)。
func (r *companyLearningActivityRepository) ListMemberActivities(
	ctx context.Context,
	companyID uint64,
	fromDate time.Time,
) ([]repository.MemberLearningActivity, error) {
	cid, ok := toInt64ID(companyID)
	if !ok {
		return []repository.MemberLearningActivity{}, nil // 存在し得ない company_id = 0 件
	}
	// activity_date は DATE 列。時刻成分が残っていると比較が timestamp に昇格して
	// 境界日(fromDate 当日)の活動が漏れるため、日付に丸めてから比較する(ListByUser と同じ流儀)。
	from := fromDate.UTC().Truncate(24 * time.Hour)
	rows, err := sqlcgen.New(r.db).ListCompanyMemberActivities(ctx, sqlcgen.ListCompanyMemberActivitiesParams{
		FromDate:  from,
		CompanyID: cid,
		// trainee 判定は正規化後の正である role_id で行う（FRESTYLE-311）。
		TraineeRoleID: domain.RoleIDTrainee,
	})
	if err != nil {
		return nil, err
	}
	out := make([]repository.MemberLearningActivity, 0, len(rows))
	for _, row := range rows {
		activity := repository.MemberLearningActivity{
			UserID:              uint64(row.UserID),
			Name:                row.Name,
			RecentActivityCount: int(row.RecentActivityCount),
		}
		if row.LastActiveDate.Valid {
			t := row.LastActiveDate.Time
			activity.LastActiveDate = &t
		}
		out = append(out, activity)
	}
	return out, nil
}
