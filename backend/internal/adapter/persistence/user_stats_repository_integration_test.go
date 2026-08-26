//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// score_cards は AI 評価スコアの旧テーブル（対応 domain 構造体は撤去済）。テーブル自体は
// 起動時と同じ DDL（infra/database/schema/core.sql）が作るので、ここでは中身だけ空にする。
func ensureScoreCards(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE score_cards RESTART IDENTITY`)
	require.NoError(t, err)
}

// TestUserStatsRepository_Integration は score_cards からの提出数(COUNT)・平均スコア(AVG)集計を
// 実 Postgres で固定する。user_id での絞り込み・NULL スコアの扱い・0 件時の COALESCE を検証する。
func TestUserStatsRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserStatsRepository(sqlDB)
	ctx := context.Background()
	ensureScoreCards(t, sqlDB)

	insert := func(userID uint64, score any) {
		_, err := sqlDB.Exec(
			`INSERT INTO score_cards (user_id, overall_score, created_at) VALUES ($1, $2, NOW())`,
			userID, score,
		)
		require.NoError(t, err)
	}

	// user 7: 80 / 90 / 100 → count=3, avg=90。他ユーザー(user 8)の 50 は混ざらない。
	insert(7, 80)
	insert(7, 90)
	insert(7, 100)
	insert(8, 50)

	t.Run("COUNT と AVG を user で絞って返す", func(t *testing.T) {
		stats, err := repo.Compute(ctx, 7)
		require.NoError(t, err)
		require.Equal(t, uint64(7), stats.UserID)
		require.Equal(t, 3, stats.TotalSessions, "自分の提出のみ数える")
		require.InDelta(t, 90.0, stats.AverageScore, 1e-9, "80/90/100 の平均")
	})

	t.Run("小数平均も保つ", func(t *testing.T) {
		ensureScoreCards(t, sqlDB)
		insert(9, 70)
		insert(9, 75)
		stats, err := repo.Compute(ctx, 9)
		require.NoError(t, err)
		require.Equal(t, 2, stats.TotalSessions)
		require.InDelta(t, 72.5, stats.AverageScore, 1e-9)
	})

	t.Run("行はあるがスコアが全て NULL なら AVG は NULL→COALESCE で 0、COUNT は行数", func(t *testing.T) {
		ensureScoreCards(t, sqlDB)
		insert(10, nil)
		insert(10, nil)
		stats, err := repo.Compute(ctx, 10)
		require.NoError(t, err)
		require.Equal(t, 2, stats.TotalSessions, "NULL スコアでも提出は数える")
		require.Equal(t, 0.0, stats.AverageScore, "AVG が NULL のときは COALESCE で 0")
	})

	t.Run("1 件も無いユーザーは count=0 / avg=0", func(t *testing.T) {
		ensureScoreCards(t, sqlDB)
		stats, err := repo.Compute(ctx, 7)
		require.NoError(t, err)
		require.Equal(t, 0, stats.TotalSessions)
		require.Equal(t, 0.0, stats.AverageScore)
	})
}
