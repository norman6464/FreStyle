//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestUserChapterViewRepository_GetLastViewedByUserAndCourse_Integration は
// (user, course) 内の last_viewed_at 最大 1 件の取得を実 Postgres で検証する。
func TestUserChapterViewRepository_GetLastViewedByUserAndCourse_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewUserChapterViewRepository(db)
	ctx := context.Background()

	testsupport.TruncateAll(t, db, "user_chapter_views")

	// user 1 が course 10 の章 101 → 102 の順に閲覧(102 が最後)。course 20 は章 201 のみ。
	require.NoError(t, repo.UpsertView(ctx, 1, 101, 10))
	require.NoError(t, repo.UpsertView(ctx, 1, 102, 10))
	require.NoError(t, repo.UpsertView(ctx, 1, 201, 20))
	// 別 user の閲覧は混ざらないことの確認用。
	require.NoError(t, repo.UpsertView(ctx, 2, 101, 10))

	t.Run("course 内で last_viewed_at 最大の 1 件を返す", func(t *testing.T) {
		got, err := repo.GetLastViewedByUserAndCourse(ctx, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, uint64(102), got.TeachingMaterialID, "最後に閲覧した章 102 が返る")
	})

	t.Run("再閲覧で last_viewed_at が更新され最新扱いになる", func(t *testing.T) {
		require.NoError(t, repo.UpsertView(ctx, 1, 101, 10)) // 101 を読み直す
		got, err := repo.GetLastViewedByUserAndCourse(ctx, 1, 10)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, uint64(101), got.TeachingMaterialID)
	})

	t.Run("履歴の無い course は nil を返す", func(t *testing.T) {
		got, err := repo.GetLastViewedByUserAndCourse(ctx, 1, 999)
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("別 user の閲覧は混ざらない", func(t *testing.T) {
		got, err := repo.GetLastViewedByUserAndCourse(ctx, 2, 20)
		require.NoError(t, err)
		require.Nil(t, got, "user 2 は course 20 を閲覧していない")
	})
}
