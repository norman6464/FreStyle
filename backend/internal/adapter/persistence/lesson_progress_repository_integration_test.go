//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
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
		testsupport.TruncateAll(t, db, "user_chapter_progress")

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
		testsupport.TruncateAll(t, db, "user_chapter_progress")

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
		testsupport.TruncateAll(t, db, "user_chapter_progress")

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

// TestLessonProgressRepository_CountCompletedByUserGroupedByCourse_Integration は
// 完了章数のコース別集計が「現存する published 教材」のみを数えることを実 Postgres で検証する。
func TestLessonProgressRepository_CountCompletedByUserGroupedByCourse_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	progress := persistence.NewLessonProgressRepository(db)
	materials := persistence.NewTeachingMaterialRepository(db)
	ctx := context.Background()

	testsupport.TruncateAll(t, db, "user_chapter_progress")
	testsupport.TruncateAll(t, db, "course_chapters")

	mk := func(courseID uint64, title string, published bool) *domain.TeachingMaterial {
		m := &domain.TeachingMaterial{
			CompanyID: 1, CourseID: courseID, CreatedByUserID: 1,
			Title: title, Content: "本文", OrderInCourse: 1, IsPublished: published,
		}
		require.NoError(t, materials.Create(ctx, m))
		return m
	}

	pub1 := mk(10, "c10-pub-1", true)
	pub2 := mk(10, "c10-pub-2", true)
	draft := mk(10, "c10-draft", false)
	other := mk(20, "c20-pub", true)

	// user 1: published 2 章 + draft 1 章 + 別コース 1 章を完了。user 2 は集計に混ざらないことの確認用。
	for _, m := range []*domain.TeachingMaterial{pub1, pub2, draft, other} {
		_, err := progress.MarkCompleted(ctx, 1, m.ID, m.CourseID)
		require.NoError(t, err)
	}
	_, err := progress.MarkCompleted(ctx, 2, pub1.ID, pub1.CourseID)
	require.NoError(t, err)

	t.Run("published のみ数え draft の完了行は含めない", func(t *testing.T) {
		counts, err := progress.CountCompletedByUserGroupedByCourse(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, map[uint64]int{10: 2, 20: 1}, counts)
	})

	t.Run("教材が削除されると完了行が残っていても集計から落ちる", func(t *testing.T) {
		require.NoError(t, materials.Delete(ctx, pub2.ID))
		counts, err := progress.CountCompletedByUserGroupedByCourse(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, map[uint64]int{10: 1, 20: 1}, counts)
	})

	t.Run("完了記録が無い user は空 map", func(t *testing.T) {
		counts, err := progress.CountCompletedByUserGroupedByCourse(ctx, 999)
		require.NoError(t, err)
		require.Empty(t, counts)
	})
}
