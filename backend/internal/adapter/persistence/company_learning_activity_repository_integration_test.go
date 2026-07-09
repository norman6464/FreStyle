//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestCompanyLearningActivityRepository_Integration は自社 trainee の学習アクティビティ集計
// (会社/ロール絞り込み・論理削除除外・最終活動日・期間内活動回数・並び順)を実 Postgres で検証する。
func TestCompanyLearningActivityRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewCompanyLearningActivityRepository(db)
	activities := persistence.NewUserDailyActivityRepository(db)
	ctx := context.Background()

	testsupport.TruncateAll(t, db, "user_daily_activities")
	testsupport.TruncateAll(t, db, "users")

	companyID := uint64(1)
	otherCompany := uint64(2)
	mkUser := func(id uint64, name, role string, company uint64, deleted bool) {
		u := &domain.User{
			ID: id, CognitoSub: name, Email: name + "@example.com", Name: name,
			CompanyID: &company, Role: role, IsActive: true,
		}
		if deleted {
			now := time.Now().UTC()
			u.DeletedAt = &now
		}
		require.NoError(t, db.Create(u).Error)
	}

	mkUser(11, "active-today", domain.RoleTrainee, companyID, false)
	mkUser(12, "active-old", domain.RoleTrainee, companyID, false)
	mkUser(13, "never-active", domain.RoleTrainee, companyID, false)
	mkUser(14, "deleted-trainee", domain.RoleTrainee, companyID, true)
	mkUser(15, "admin-not-counted", domain.RoleCompanyAdmin, companyID, false)
	mkUser(16, "other-company", domain.RoleTrainee, otherCompany, false)

	today := time.Now().UTC()
	tenDaysAgo := today.AddDate(0, 0, -10)
	inc := repository.UserDailyActivityIncrement{LessonCount: 1}
	require.NoError(t, activities.Increment(ctx, 11, today, inc))
	require.NoError(t, activities.Increment(ctx, 11, today, inc)) // 同日 2 回目
	require.NoError(t, activities.Increment(ctx, 12, tenDaysAgo, inc))
	require.NoError(t, activities.Increment(ctx, 16, today, inc)) // 他社

	fromDate := today.AddDate(0, 0, -6)
	rows, err := repo.ListMemberActivities(ctx, companyID, fromDate)
	require.NoError(t, err)
	require.Len(t, rows, 3, "自社 trainee のみ(論理削除・admin・他社は除外)")

	// 並び順: 最終活動日の新しい順 → 未活動は末尾。
	require.Equal(t, uint64(11), rows[0].UserID)
	require.NotNil(t, rows[0].LastActiveDate)
	require.Equal(t, today.Format("2006-01-02"), rows[0].LastActiveDate.Format("2006-01-02"))
	require.Equal(t, 2, rows[0].RecentActivityCount, "期間内の活動回数が合算される")

	require.Equal(t, uint64(12), rows[1].UserID)
	require.Equal(t, 0, rows[1].RecentActivityCount, "期間外の活動は数えない")
	require.NotNil(t, rows[1].LastActiveDate)

	require.Equal(t, uint64(13), rows[2].UserID)
	require.Nil(t, rows[2].LastActiveDate)
	require.Equal(t, 0, rows[2].RecentActivityCount)

	// 誰も居ない会社は空スライス。
	empty, err := repo.ListMemberActivities(ctx, 999, fromDate)
	require.NoError(t, err)
	require.Empty(t, empty)
}
