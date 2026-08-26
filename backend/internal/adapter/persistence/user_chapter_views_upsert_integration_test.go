//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// fetchChapterView は (user, chapter) の生の行を直接読み出す（repository を介さず事実を確認する）。
func fetchChapterView(t *testing.T, db *gorm.DB, userID, chapterID uint64) domain.UserChapterView {
	t.Helper()
	var row domain.UserChapterView
	require.NoError(t, db.WithContext(context.Background()).
		Where("user_id = ? AND chapter_id = ?", userID, chapterID).
		Take(&row).Error)
	return row
}

// TestUserChapterViewRepository_UpsertView_Integration は章閲覧の upsert を実 Postgres で固定する。
// 初回は view_count=1・first/last_viewed_at をセット、再閲覧は view_count 加算・last_viewed_at 更新で
// first_viewed_at は保持、course_id は最新に更新されることを主張する。
func TestUserChapterViewRepository_UpsertView_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewUserChapterViewRepository(db)
	ctx := context.Background()

	testsupport.TruncateAll(t, db, "user_chapter_views")

	require.NoError(t, repo.UpsertView(ctx, 1, 100, 10))
	first := fetchChapterView(t, db, 1, 100)

	t.Run("初回は view_count=1 で first/last がセットされる", func(t *testing.T) {
		require.Equal(t, 1, first.ViewCount)
		require.False(t, first.FirstViewedAt.IsZero())
		require.False(t, first.LastViewedAt.IsZero())
		require.Equal(t, uint64(10), first.CourseID)
	})

	t.Run("再閲覧で view_count が加算され last_viewed_at が進み first_viewed_at は保持される", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // last_viewed_at が確実に進むよう僅かに待つ
		require.NoError(t, repo.UpsertView(ctx, 1, 100, 10))
		second := fetchChapterView(t, db, 1, 100)

		require.Equal(t, 2, second.ViewCount, "view_count は +1")
		require.Equal(t, first.FirstViewedAt.UTC(), second.FirstViewedAt.UTC(), "first_viewed_at は初回のまま")
		require.True(t, second.LastViewedAt.After(first.LastViewedAt), "last_viewed_at は前進する")
	})

	t.Run("再閲覧で course_id は最新値へ更新される", func(t *testing.T) {
		require.NoError(t, repo.UpsertView(ctx, 1, 100, 20)) // 別コース経由で開いた
		row := fetchChapterView(t, db, 1, 100)
		require.Equal(t, uint64(20), row.CourseID)
		require.Equal(t, 3, row.ViewCount)
	})
}
