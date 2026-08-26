//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// tenantBridgeTables は会社↔ワークスペースの橋渡しが触るテーブル（TRUNCATE 対象）。
// workspaces を参照する FK があるため CASCADE で companies / users も一緒に消えるが、
// 依存を読み手に見せるために全部並べる。
var tenantBridgeTables = []string{"users", "user_oidc_identities", "companies", "workspaces"}

// workspaceSlugPattern は自動採番した slug の形（ws- + UUID の 32 桁 hex）。
var workspaceSlugPattern = regexp.MustCompile(`^ws-[0-9a-f]{32}$`)

// insertCompany は会社を 1 件作る（id を明示して固定する）。
func insertCompany(t *testing.T, db *gorm.DB, id int64, name string, aiChat, active bool) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO companies (id, name, ai_chat_enabled_for_trainees, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, NOW(), NOW())`, id, name, aiChat, active,
	).Error)
}

// companyWorkspaceID は会社に紐づいたワークスペース ID を返す（未紐付けなら Valid=false）。
func companyWorkspaceID(t *testing.T, db *gorm.DB, companyID int64) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.Raw(`SELECT workspace_id FROM companies WHERE id = ?`, companyID).Row().Scan(&got))
	return got
}

// userWorkspaceID はユーザーの workspace_id を返す（未設定なら Valid=false）。
func userWorkspaceID(t *testing.T, db *gorm.DB, userID uint64) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.Raw(`SELECT workspace_id FROM users WHERE id = ?`, userID).Row().Scan(&got))
	return got
}

// runStartupBackfill は起動時に走る処理（スキーマ適用 → バックフィル）と同じ順で 1 回分を流す。
func runStartupBackfill(ctx context.Context, t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, database.ApplyTenantBridgeSchema(ctx, sqlDB))
	require.NoError(t, database.BackfillWorkspacesFromCompanies(ctx, sqlDB))
}

// TestTenantBridgeBackfill_Integration は「会社を workspaces へ映す」バックフィルを実 Postgres で固定する。
// 読み取りは何も変えていない段なので、ここで確かめるのは新しい列の中身だけ。
func TestTenantBridgeBackfill_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	ctx := context.Background()

	t.Run("会社ごとにワークスペースが 1 つでき、所属ユーザーへ写る", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "株式会社ふれすたいる", true, true)
		insertCompany(t, db, 2, "Second Corp", false, false)

		repo := persistence.NewUserRepository(sqlDB)
		c1, c2 := uint64(1), uint64(2)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "a@example.com", Name: "A", Role: domain.RoleTrainee, CompanyID: &c1,
		}, domain.OidcProviderCognito, "sub-a"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "b@example.com", Name: "B", Role: domain.RoleTrainee, CompanyID: &c2,
		}, domain.OidcProviderCognito, "sub-b"))
		// 未所属（company_id IS NULL）。運営管理者のようにどの会社にも属さない利用者。
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "Root", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "sub-root"))

		runStartupBackfill(ctx, t, db)

		var count int64
		require.NoError(t, db.Raw(`SELECT count(*) FROM workspaces`).Scan(&count).Error)
		require.Equal(t, int64(2), count, "会社 1 件につきワークスペースは 1 つ")

		ws1 := companyWorkspaceID(t, db, 1)
		ws2 := companyWorkspaceID(t, db, 2)
		require.True(t, ws1.Valid)
		require.True(t, ws2.Valid)
		require.NotEqual(t, ws1.UUID, ws2.UUID, "会社ごとに別のワークスペース")

		// 会社名と設定が写っていること（設定は移行期間中 companies が正本）。
		var name string
		var aiChat, active sql.NullBool
		require.NoError(t, db.Raw(
			`SELECT name, ai_chat_enabled_for_trainees, is_active FROM workspaces WHERE id = ?`, ws1.UUID,
		).Row().Scan(&name, &aiChat, &active))
		require.Equal(t, "株式会社ふれすたいる", name)
		require.Equal(t, sql.NullBool{Bool: true, Valid: true}, aiChat)
		require.Equal(t, sql.NullBool{Bool: true, Valid: true}, active)
		require.NoError(t, db.Raw(
			`SELECT ai_chat_enabled_for_trainees, is_active FROM workspaces WHERE id = ?`, ws2.UUID,
		).Row().Scan(&aiChat, &active))
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, aiChat)
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, active)

		userA, err := repo.FindByCognitoSub(ctx, "sub-a")
		require.NoError(t, err)
		userB, err := repo.FindByCognitoSub(ctx, "sub-b")
		require.NoError(t, err)
		userRoot, err := repo.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)

		require.Equal(t, ws1, userWorkspaceID(t, db, userA.ID))
		require.Equal(t, ws2, userWorkspaceID(t, db, userB.ID))
		require.False(t, userWorkspaceID(t, db, userRoot.ID).Valid, "未所属ユーザーは NULL のまま")

		// 読み取りは company_id のまま。バックフィルで見え方が変わっていないこと。
		require.Equal(t, uint64(1), *userA.CompanyID)
		require.Nil(t, userRoot.CompanyID)
	})

	t.Run("何度流してもワークスペースが増えず ID もぶれない", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "会社 A", true, true)
		insertCompany(t, db, 2, "会社 B", true, true)

		runStartupBackfill(ctx, t, db)
		first1 := companyWorkspaceID(t, db, 1)
		first2 := companyWorkspaceID(t, db, 2)

		for i := 0; i < 2; i++ { // 起動 2 回目・3 回目
			runStartupBackfill(ctx, t, db)
			var count int64
			require.NoError(t, db.Raw(`SELECT count(*) FROM workspaces`).Scan(&count).Error)
			require.Equal(t, int64(2), count, "再実行でワークスペースが増えない")
			require.Equal(t, first1, companyWorkspaceID(t, db, 1), "再実行で紐付け先が変わらない")
			require.Equal(t, first2, companyWorkspaceID(t, db, 2))
		}
	})

	t.Run("slug はグローバル一意で長さ制約を満たす", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "同名の会社", true, true)
		insertCompany(t, db, 2, "同名の会社", true, true) // companies.name に一意制約は無い

		runStartupBackfill(ctx, t, db)

		rows, err := db.Raw(`SELECT slug FROM workspaces ORDER BY slug`).Rows()
		require.NoError(t, err)
		defer func() { require.NoError(t, rows.Close()) }()
		slugs := make([]string, 0, 2)
		for rows.Next() {
			var slug string
			require.NoError(t, rows.Scan(&slug))
			require.Regexp(t, workspaceSlugPattern, slug)
			require.LessOrEqual(t, len(slug), 64)
			slugs = append(slugs, slug)
		}
		require.NoError(t, rows.Err())
		require.Len(t, slugs, 2)
		require.NotEqual(t, slugs[0], slugs[1], "会社名が同じでも slug は衝突しない")
	})

	t.Run("存在しないワークスペースは指せない", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "会社", true, true)

		err := db.Exec(`UPDATE companies SET workspace_id = ? WHERE id = 1`, uuid.New()).Error
		require.Error(t, err, "FK が無いまま company_id の轍を踏まない")
	})
}

// TestTenantBridgeDualWrite_Integration は所属を書く経路が company_id と workspace_id の
// 両方を埋めることを固定する。読み取りは company_id のままで変えていない。
func TestTenantBridgeDualWrite_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	ctx := context.Background()

	t.Run("招待からのユーザー作成で両方の列が埋まる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "会社 A", true, true)
		runStartupBackfill(ctx, t, db)
		ws1 := companyWorkspaceID(t, db, 1)

		repo := persistence.NewUserRepository(sqlDB)
		cid := uint64(1)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "new@example.com", Name: "新入社員", Role: domain.RoleTrainee, CompanyID: &cid,
		}, domain.OidcProviderCognito, "sub-new"))

		got, err := repo.FindByCognitoSub(ctx, "sub-new")
		require.NoError(t, err)
		require.Equal(t, uint64(1), *got.CompanyID)
		require.Equal(t, ws1, userWorkspaceID(t, db, got.ID))
	})

	t.Run("未所属のまま作られたユーザーは workspace_id も NULL", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "運営", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "sub-root"))

		got, err := repo.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)
		require.Nil(t, got.CompanyID)
		require.False(t, userWorkspaceID(t, db, got.ID).Valid)
	})

	t.Run("会社の付け替えで workspace_id も追随する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "会社 A", true, true)
		insertCompany(t, db, 2, "会社 B", true, true)
		runStartupBackfill(ctx, t, db)
		ws1 := companyWorkspaceID(t, db, 1)
		ws2 := companyWorkspaceID(t, db, 2)

		repo := persistence.NewUserRepository(sqlDB)
		cid := uint64(1)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "move@example.com", Name: "異動", Role: domain.RoleTrainee, CompanyID: &cid,
		}, domain.OidcProviderCognito, "sub-move"))
		got, err := repo.FindByCognitoSub(ctx, "sub-move")
		require.NoError(t, err)
		require.Equal(t, ws1, userWorkspaceID(t, db, got.ID))

		require.NoError(t, repo.UpdateCompanyID(ctx, got.ID, 2))

		moved, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.Equal(t, uint64(2), *moved.CompanyID)
		require.Equal(t, ws2, userWorkspaceID(t, db, got.ID))
	})

	t.Run("会社設定の更新がワークスペースへ写る", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "会社 A", true, true)
		runStartupBackfill(ctx, t, db)
		ws1 := companyWorkspaceID(t, db, 1)

		companies := persistence.NewCompanyRepository(sqlDB)
		require.NoError(t, companies.UpdateAiChatEnabled(ctx, 1, false))
		require.NoError(t, companies.UpdateActive(ctx, 1, false))

		var aiChat, active sql.NullBool
		require.NoError(t, db.Raw(
			`SELECT ai_chat_enabled_for_trainees, is_active FROM workspaces WHERE id = ?`, ws1.UUID,
		).Row().Scan(&aiChat, &active))
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, aiChat)
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, active)

		// 読み取り（companies）は従来どおり。
		company, err := companies.FindByID(ctx, 1)
		require.NoError(t, err)
		require.False(t, company.AiChatEnabledForTrainees)
		require.False(t, company.IsActive)
	})

	t.Run("存在しない会社の有効/無効更新は not found のまま", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		companies := persistence.NewCompanyRepository(sqlDB)
		require.ErrorIs(t, companies.UpdateActive(ctx, 999, false), domain.ErrNotFound)
	})

	t.Run("ワークスペース未紐付けの会社でも設定更新は成功する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "未紐付けの会社", true, true) // バックフィル前
		companies := persistence.NewCompanyRepository(sqlDB)
		require.NoError(t, companies.UpdateActive(ctx, 1, false))
		require.NoError(t, companies.UpdateAiChatEnabled(ctx, 1, false))
		require.False(t, companyWorkspaceID(t, db, 1).Valid)
	})
}
