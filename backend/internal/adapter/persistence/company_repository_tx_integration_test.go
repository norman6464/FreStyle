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
)

// workspaceSettings は写し先（workspaces）のテナント設定を読む。
func workspaceSettings(t *testing.T, db *sql.DB, companyID int64) (active sql.NullBool) {
	t.Helper()
	require.NoError(t, db.QueryRow(
		`SELECT w.is_active
		 FROM workspaces w JOIN companies c ON c.workspace_id = w.id
		 WHERE c.id = $1`, companyID,
	).Scan(&active))
	return active
}

// blockWorkspaceWrites は workspaces への「false を書く」更新を必ず失敗させる CHECK 制約を
// 一時的に張る。写し先だけが失敗したときに companies 側も巻き戻るかを確かめるための足場。
func blockWorkspaceWrites(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(
		`ALTER TABLE workspaces ADD CONSTRAINT tmp_ck_workspaces_no_false
		 CHECK (is_active IS NOT FALSE)`,
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(
			`ALTER TABLE workspaces DROP CONSTRAINT IF EXISTS tmp_ck_workspaces_no_false`,
		)
		require.NoError(t, err)
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
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCompanyRepository(sqlDB)
	ctx := context.Background()

	t.Run("UpdateActive: 写しに失敗したら会社の更新も巻き戻る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		runStartupBackfill(ctx, t, sqlDB)
		blockWorkspaceWrites(t, sqlDB)

		require.Error(t, repo.UpdateActive(ctx, 1, false), "写し先の制約違反はそのままエラーになる")

		got, err := repo.FindByID(ctx, 1)
		require.NoError(t, err)
		require.True(t, got.IsActive, "companies 側の更新も巻き戻っている")
		active := workspaceSettings(t, sqlDB, 1)
		require.Equal(t, sql.NullBool{Bool: true, Valid: true}, active)
	})

	t.Run("成功時は companies と workspaces が同時に更新される", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		runStartupBackfill(ctx, t, sqlDB)

		require.NoError(t, repo.UpdateActive(ctx, 1, false))

		got, err := repo.FindByID(ctx, 1)
		require.NoError(t, err)
		require.False(t, got.IsActive)
		active := workspaceSettings(t, sqlDB, 1)
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, active)
	})

	// 期待値を「0 件更新で成功」から not-found へ更新した理由:
	//   会社行が無くても handler が 200 と要求どおりの値を返すと、管理者の画面では設定が
	//   切り替わったように見えて実際は何も保存されていない（次に開いたときだけ元へ戻り、
	//   どこで失われたのか分からない）。0 件更新は not-found として返す。
	t.Run("存在しない会社への設定更新は not-found", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		require.ErrorIs(t, repo.UpdateActive(ctx, 999, false), domain.ErrNotFound)
	})
}
