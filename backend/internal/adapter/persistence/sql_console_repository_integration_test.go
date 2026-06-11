//go:build integration

package persistence_test

import (
	"context"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestSQLConsoleRepository_Integration は実 Postgres で read-only SQL コンソールの
// ①SELECT が結果を返す ②maxRows で打ち切る ③書き込みは read-only トランザクションが拒否する
// を検証する（DB レベルの最終防壁の確認）。
func TestSQLConsoleRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewSQLConsoleRepository(db)
	ctx := context.Background()

	t.Run("SELECT は列と行を返す", func(t *testing.T) {
		res, err := repo.RunReadOnly(ctx, "SELECT generate_series(1, 3) AS n", 1000)
		require.NoError(t, err)
		require.Equal(t, []string{"n"}, res.Columns)
		require.Len(t, res.Rows, 3)
		require.False(t, res.Truncated)
	})

	t.Run("maxRows を超えると truncated", func(t *testing.T) {
		res, err := repo.RunReadOnly(ctx, "SELECT generate_series(1, 10) AS n", 4)
		require.NoError(t, err)
		require.Len(t, res.Rows, 4)
		require.True(t, res.Truncated)
	})

	t.Run("書き込みは read-only トランザクションが拒否する", func(t *testing.T) {
		// master_exercises は AutoMigrate で存在する。read-only tx なので DELETE は失敗する。
		_, err := repo.RunReadOnly(ctx, "DELETE FROM master_exercises", 1000)
		require.Error(t, err)
		require.Contains(t, strings.ToLower(err.Error()), "read-only")
	})
}
