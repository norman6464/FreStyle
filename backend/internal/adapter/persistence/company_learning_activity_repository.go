package persistence

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// companyLearningActivityRepository は [repository.CompanyLearningActivitySummarizer] の GORM 実装。
// 読み取り集計のみのため生 SQL 直書き(db.Raw)。
type companyLearningActivityRepository struct {
	db *gorm.DB
}

// NewCompanyLearningActivityRepository は CompanyLearningActivitySummarizer の GORM 実装を返す。
func NewCompanyLearningActivityRepository(db *gorm.DB) repository.CompanyLearningActivitySummarizer {
	return &companyLearningActivityRepository{db: db}
}

// ListMemberActivities は自社 trainee ごとの最終活動日と fromDate 以降の活動回数を 1 クエリで集計する
// (trainee ごとの個別クエリだと N+1 になるため LEFT JOIN + GROUP BY で一括取得)。
func (r *companyLearningActivityRepository) ListMemberActivities(
	ctx context.Context,
	companyID uint64,
	fromDate time.Time,
) ([]repository.MemberLearningActivity, error) {
	const q = `
SELECT u.id AS user_id,
       u.name,
       MAX(a.activity_date) AS last_active_date,
       COALESCE(SUM(
         CASE WHEN a.activity_date >= ?
              THEN a.exercise_count + a.chapter_count + a.ai_chat_count + a.note_count
              ELSE 0 END
       ), 0) AS recent_activity_count
FROM users u
LEFT JOIN user_daily_activities a ON a.user_id = u.id
WHERE u.company_id = ? AND u.role = ? AND u.deleted_at IS NULL
GROUP BY u.id, u.name
ORDER BY MAX(a.activity_date) DESC NULLS LAST, u.id ASC`
	var rows []struct {
		UserID              uint64
		Name                string
		LastActiveDate      *time.Time
		RecentActivityCount int
	}
	// activity_date は DATE 列。時刻成分が残っていると比較が timestamp に昇格して
	// 境界日(fromDate 当日)の活動が漏れるため、日付に丸めてから比較する(ListByUser と同じ流儀)。
	from := fromDate.UTC().Truncate(24 * time.Hour)
	if err := r.db.WithContext(ctx).
		Raw(q, from, companyID, domain.RoleTrainee).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]repository.MemberLearningActivity, 0, len(rows))
	for _, row := range rows {
		out = append(out, repository.MemberLearningActivity{
			UserID:              row.UserID,
			Name:                row.Name,
			LastActiveDate:      row.LastActiveDate,
			RecentActivityCount: row.RecentActivityCount,
		})
	}
	return out, nil
}
