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

func (r *companyLearningActivityRepository) ListMemberActivitiesByWorkspace(
	ctx context.Context,
	workspaceID string,
	fromDate time.Time,
) ([]repository.MemberLearningActivity, error) {
	wid, ok := toNullUUID(workspaceID)
	if !ok {
		return []repository.MemberLearningActivity{}, nil // 不正 / 空の ID は該当なしと同じ扱い
	}
	rows, err := sqlcgen.New(r.db).ListCompanyMemberActivitiesByWorkspace(ctx, sqlcgen.ListCompanyMemberActivitiesByWorkspaceParams{
		FromDate:      truncateToUTCDate(fromDate),
		WorkspaceID:   wid,
		TraineeRoleID: domain.RoleIDTrainee,
	})
	if err != nil {
		return nil, err
	}
	return toMemberLearningActivitiesFromWorkspaceRows(rows), nil
}

// truncateToUTCDate は activity_date（DATE 列）との比較のために時刻成分を落とす。
// 時刻成分が残っていると比較が timestamp に昇格して境界日(fromDate 当日)の活動が漏れる
// （ListByUser と同じ流儀）。
func truncateToUTCDate(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

func toMemberLearningActivitiesFromWorkspaceRows(rows []sqlcgen.ListCompanyMemberActivitiesByWorkspaceRow) []repository.MemberLearningActivity {
	out := make([]repository.MemberLearningActivity, 0, len(rows))
	for _, row := range rows {
		out = append(out, toMemberLearningActivity(row.UserID, row.Name, row.LastActiveDate, row.RecentActivityCount))
	}
	return out
}

func toMemberLearningActivity(userID int64, name string, lastActiveDate sql.NullTime, recentActivityCount int64) repository.MemberLearningActivity {
	activity := repository.MemberLearningActivity{
		UserID:              uint64(userID),
		Name:                name,
		RecentActivityCount: int(recentActivityCount),
	}
	if lastActiveDate.Valid {
		t := lastActiveDate.Time
		activity.LastActiveDate = &t
	}
	return activity
}
