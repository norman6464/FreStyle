//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestCompanyLearningActivityRepository_Integration は自ワークスペース trainee の学習アクティビティ集計
// (ワークスペース/ロール絞り込み・論理削除除外・最終活動日・期間内活動回数・並び順)を実 Postgres で検証する。
func TestCompanyLearningActivityRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCompanyLearningActivityRepository(sqlDB)
	activities := persistence.NewUserDailyActivityRepository(sqlDB)
	ctx := context.Background()

	testsupport.TruncateAll(t, sqlDB,
		append([]string{"user_daily_activities", "users"}, tenantBridgeTables...)...)
	insertCompany(t, sqlDB, 1, "会社 A", true)
	insertCompany(t, sqlDB, 2, "会社 B", true)
	runStartupBackfill(ctx, t, sqlDB)
	workspaceID := companyWorkspaceID(t, sqlDB, 1).UUID.String()
	otherWorkspace := companyWorkspaceID(t, sqlDB, 2).UUID.String()

	userRepo := persistence.NewUserRepository(sqlDB)
	mkUser := func(id uint64, name string, role domain.RoleName, workspace string, deleted bool) {
		u := &domain.User{
			ID: id, Email: name + "@example.com", Name: name,
			WorkspaceID: &workspace, Role: role, IsActive: true,
		}
		if deleted {
			now := time.Now().UTC()
			u.DeletedAt = &now
		}
		// role_id の解決（resolveRoleID）と identity 作成を通すため repository 経由で作成する。
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, name))
	}

	mkUser(11, "active-today", domain.RoleTrainee, workspaceID, false)
	mkUser(12, "active-old", domain.RoleTrainee, workspaceID, false)
	mkUser(13, "never-active", domain.RoleTrainee, workspaceID, false)
	mkUser(14, "deleted-trainee", domain.RoleTrainee, workspaceID, true)
	mkUser(15, "admin-not-counted", domain.RoleCompanyAdmin, workspaceID, false)
	mkUser(16, "other-workspace", domain.RoleTrainee, otherWorkspace, false)

	today := time.Now().UTC()
	boundaryDay := today.AddDate(0, 0, -6) // 集計ウィンドウの境界日(7 日間の初日)
	tenDaysAgo := today.AddDate(0, 0, -10)
	inc := repository.UserDailyActivityIncrement{LessonCount: 1}
	require.NoError(t, activities.Increment(ctx, 11, today, inc))
	require.NoError(t, activities.Increment(ctx, 11, boundaryDay, inc)) // 境界日の活動も数える
	require.NoError(t, activities.Increment(ctx, 12, tenDaysAgo, inc))
	require.NoError(t, activities.Increment(ctx, 16, today, inc)) // 他ワークスペース

	// fromDate に時刻成分が残っていても境界日(当日)の活動が漏れないこと(日付へ丸めて比較)。
	fromDate := boundaryDay
	rows, err := repo.ListMemberActivitiesByWorkspace(ctx, workspaceID, fromDate)
	require.NoError(t, err)
	require.Len(t, rows, 3, "自ワークスペースの trainee のみ(論理削除・admin・他ワークスペースは除外)")

	// 並び順: 最終活動日の新しい順 → 未活動は末尾。
	require.Equal(t, uint64(11), rows[0].UserID)
	require.NotNil(t, rows[0].LastActiveDate)
	require.Equal(t, today.Format("2006-01-02"), rows[0].LastActiveDate.Format("2006-01-02"))
	require.Equal(t, 2, rows[0].RecentActivityCount, "今日 + 境界日の活動が合算される")

	require.Equal(t, uint64(12), rows[1].UserID)
	require.Equal(t, 0, rows[1].RecentActivityCount, "期間外の活動は数えない")
	require.NotNil(t, rows[1].LastActiveDate)

	require.Equal(t, uint64(13), rows[2].UserID)
	require.Nil(t, rows[2].LastActiveDate)
	require.Equal(t, 0, rows[2].RecentActivityCount)

	// 誰も居ないワークスペースは空スライス。
	empty, err := repo.ListMemberActivitiesByWorkspace(ctx, uuid.NewString(), fromDate)
	require.NoError(t, err)
	require.Empty(t, empty)
}

// TestCompanyLearningActivityRepository_ByWorkspace_Integration は
// ListMemberActivitiesByWorkspace が workspace_id で正しく絞り、他ワークスペースの
// メンバー集計が混ざらないことを実 Postgres で検証する。
func TestCompanyLearningActivityRepository_ByWorkspace_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCompanyLearningActivityRepository(sqlDB)
	activities := persistence.NewUserDailyActivityRepository(sqlDB)
	userRepo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()

	testsupport.TruncateAll(t, sqlDB, append([]string{"user_daily_activities", "users"}, tenantBridgeTables...)...)
	insertCompany(t, sqlDB, 1, "会社 A", true)
	insertCompany(t, sqlDB, 2, "会社 B", true)
	runStartupBackfill(ctx, t, sqlDB)
	ws1 := companyWorkspaceID(t, sqlDB, 1)
	ws2 := companyWorkspaceID(t, sqlDB, 2)
	require.True(t, ws1.Valid)
	require.True(t, ws2.Valid)
	require.NotEqual(t, ws1.UUID, ws2.UUID)

	mkUser := func(id uint64, name string, role domain.RoleName, workspace string) {
		u := &domain.User{
			ID: id, Email: name + "@example.com", Name: name,
			WorkspaceID: &workspace, Role: role, IsActive: true,
		}
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, name))
	}
	mkUser(21, "ws1-trainee", domain.RoleTrainee, ws1.UUID.String())
	mkUser(22, "ws2-trainee", domain.RoleTrainee, ws2.UUID.String())

	today := time.Now().UTC()
	inc := repository.UserDailyActivityIncrement{LessonCount: 1}
	require.NoError(t, activities.Increment(ctx, 21, today, inc))
	require.NoError(t, activities.Increment(ctx, 22, today, inc))

	fromDate := today.AddDate(0, 0, -6)
	rows, err := repo.ListMemberActivitiesByWorkspace(ctx, ws1.UUID.String(), fromDate)
	require.NoError(t, err)
	require.Len(t, rows, 1, "他ワークスペースのメンバーが混ざらない")
	require.Equal(t, uint64(21), rows[0].UserID)
	require.Equal(t, 1, rows[0].RecentActivityCount)

	// 不正 / 空の workspace_id は該当なし扱い（toNullUUID の失敗経路を両方とも確認する）。
	for _, invalid := range []string{"", "not-a-uuid"} {
		empty, err := repo.ListMemberActivitiesByWorkspace(ctx, invalid, fromDate)
		require.NoError(t, err)
		require.Empty(t, empty, "workspaceID=%q", invalid)
	}
}
