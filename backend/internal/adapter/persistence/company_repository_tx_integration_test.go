//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// workspaceSettings は写し先（workspaces）のテナント設定を読む。
func workspaceSettings(t *testing.T, db *gorm.DB, companyID int64) (aiChat, active sql.NullBool) {
	t.Helper()
	require.NoError(t, db.Raw(
		`SELECT w.ai_chat_enabled_for_trainees, w.is_active
		 FROM workspaces w JOIN companies c ON c.workspace_id = w.id
		 WHERE c.id = ?`, companyID,
	).Row().Scan(&aiChat, &active))
	return aiChat, active
}

// blockWorkspaceWrites は workspaces への「false を書く」更新を必ず失敗させる CHECK 制約を
// 一時的に張る。写し先だけが失敗したときに companies 側も巻き戻るかを確かめるための足場。
func blockWorkspaceWrites(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(
		`ALTER TABLE workspaces ADD CONSTRAINT tmp_ck_workspaces_no_false
		 CHECK (is_active IS NOT FALSE AND ai_chat_enabled_for_trainees IS NOT FALSE)`,
	).Error)
	t.Cleanup(func() {
		require.NoError(t, db.Exec(
			`ALTER TABLE workspaces DROP CONSTRAINT IF EXISTS tmp_ck_workspaces_no_false`,
		).Error)
	})
}

// TestCompanyRepositoryMirrorAtomicity_Integration は companies の更新と workspaces への
// 写しが 1 つのトランザクションで不可分に行われることを実 PostgreSQL で固定する。
//
// 移行期間中はテナント設定という 1 つの事実を companies と workspaces の 2 か所で持つ。
// 片方だけがコミットされると、どちらを読むかで答えが変わる（AI チャットが使えるはずの会社で
// 使えない・停止したはずの会社が生きている）。だから「写しに失敗したら会社側も巻き戻る」
// ことが安全弁の本体で、mirror をトランザクションの外へ出すとここが破れる。
func TestCompanyRepositoryMirrorAtomicity_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewCompanyRepository(sqlDB)
	ctx := context.Background()

	t.Run("UpdateActive: 写しに失敗したら会社の更新も巻き戻る", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "会社 A", true, true)
		runStartupBackfill(ctx, t, db)
		blockWorkspaceWrites(t, db)

		require.Error(t, repo.UpdateActive(ctx, 1, false), "写し先の制約違反はそのままエラーになる")

		got, err := repo.FindByID(ctx, 1)
		require.NoError(t, err)
		require.True(t, got.IsActive, "companies 側の更新も巻き戻っている")
		_, active := workspaceSettings(t, db, 1)
		require.Equal(t, sql.NullBool{Bool: true, Valid: true}, active)
	})

	t.Run("UpdateAiChatEnabled: 写しに失敗したら会社の更新も巻き戻る", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "会社 A", true, true)
		runStartupBackfill(ctx, t, db)
		blockWorkspaceWrites(t, db)

		require.Error(t, repo.UpdateAiChatEnabled(ctx, 1, false))

		got, err := repo.FindByID(ctx, 1)
		require.NoError(t, err)
		require.True(t, got.AiChatEnabledForTrainees, "companies 側の更新も巻き戻っている")
		aiChat, _ := workspaceSettings(t, db, 1)
		require.Equal(t, sql.NullBool{Bool: true, Valid: true}, aiChat)
	})

	t.Run("成功時は companies と workspaces が同時に更新される", func(t *testing.T) {
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		insertCompany(t, db, 1, "会社 A", true, true)
		runStartupBackfill(ctx, t, db)

		require.NoError(t, repo.UpdateActive(ctx, 1, false))
		require.NoError(t, repo.UpdateAiChatEnabled(ctx, 1, false))

		got, err := repo.FindByID(ctx, 1)
		require.NoError(t, err)
		require.False(t, got.IsActive)
		require.False(t, got.AiChatEnabledForTrainees)
		aiChat, active := workspaceSettings(t, db, 1)
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, aiChat)
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, active)
	})

	t.Run("存在しない会社への UpdateAiChatEnabled は 0 件更新で成功する", func(t *testing.T) {
		// UpdateActive だけが not-found をエラーにする（UpdateAiChatEnabled は従来から
		// 件数を見ていない）。移行でこの非対称を変えない。
		testsupport.TruncateAll(t, db, tenantBridgeTables...)
		require.NoError(t, repo.UpdateAiChatEnabled(ctx, 999, false))
		require.ErrorIs(t, repo.UpdateActive(ctx, 999, false), domain.ErrNotFound)
	})
}
