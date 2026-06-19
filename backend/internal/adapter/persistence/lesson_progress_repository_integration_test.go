//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestLessonProgressRepository_Integration は完了記録の冪等 upsert / 取消 / user 絞り込みを
// 実 Postgres で検証する。
func TestLessonProgressRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewLessonProgressRepository(db)
	ctx := context.Background()

	t.Run("MarkCompleted は冪等（二重実行でも 1 件）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "user_lesson_progress")

		changed, err := repo.MarkCompleted(ctx, 1, 10, 100)
		require.NoError(t, err)
		require.True(t, changed, "初回は true")

		changed2, err := repo.MarkCompleted(ctx, 1, 10, 100) // 二重実行
		require.NoError(t, err)
		require.False(t, changed2, "重複は false")

		rows, err := repo.ListByUser(ctx, 1)
		require.NoError(t, err)
		require.Len(t, rows, 1, "(user, material) 一意制約で 1 件のまま")
		require.Equal(t, uint64(100), rows[0].CourseID)
	})

	t.Run("ListByUser は user で絞り込む", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "user_lesson_progress")

		_, err := repo.MarkCompleted(ctx, 1, 10, 100)
		require.NoError(t, err)
		_, err = repo.MarkCompleted(ctx, 1, 11, 100)
		require.NoError(t, err)
		_, err = repo.MarkCompleted(ctx, 2, 10, 100) // 別 user
		require.NoError(t, err)

		rows, err := repo.ListByUser(ctx, 1)
		require.NoError(t, err)
		require.Len(t, rows, 2, "user 1 の 2 件のみ")
	})

	t.Run("MarkIncomplete は行を削除する（未記録でもエラーにしない）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "user_lesson_progress")

		_, err := repo.MarkCompleted(ctx, 1, 10, 100)
		require.NoError(t, err)
		require.NoError(t, repo.MarkIncomplete(ctx, 1, 10))

		rows, err := repo.ListByUser(ctx, 1)
		require.NoError(t, err)
		require.Empty(t, rows)

		// 未記録に対する取消も冪等にエラーなし。
		require.NoError(t, repo.MarkIncomplete(ctx, 1, 999))
	})
}
