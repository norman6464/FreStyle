//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestUserDailyActivityRepository_Increment_Integration は日次サマリーの upsert 加算を
// 実 Postgres で固定する。初回は delta で INSERT、2 回目以降は各カウンタへ加算する。
func TestUserDailyActivityRepository_Increment_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserDailyActivityRepository(sqlDB)
	ctx := context.Background()

	testsupport.TruncateAll(t, sqlDB, "user_daily_activities")

	day1 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)

	// 初回 INSERT。
	require.NoError(t, repo.Increment(ctx, 1, day1, repository.UserDailyActivityIncrement{
		ExerciseCount: 1, CorrectCount: 1,
	}))
	// 同 (user, date) への 2 回目は各カウンタへ加算。LessonCount は列 chapter_count へ入る。
	require.NoError(t, repo.Increment(ctx, 1, day1, repository.UserDailyActivityIncrement{
		ExerciseCount: 2, LessonCount: 1, NoteCount: 1,
	}))
	// 別日は別行。
	require.NoError(t, repo.Increment(ctx, 1, day2, repository.UserDailyActivityIncrement{NoteCount: 5}))
	// 別 user は混ざらない。
	require.NoError(t, repo.Increment(ctx, 2, day1, repository.UserDailyActivityIncrement{ExerciseCount: 9}))

	t.Run("同一日の加算が各カウンタへ蓄積される", func(t *testing.T) {
		rows, err := repo.ListByUser(ctx, 1, day1, day1)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		got := rows[0]
		require.Equal(t, uint64(1), got.UserID)
		require.Equal(t, 3, got.ExerciseCount, "1 + 2")
		require.Equal(t, 1, got.CorrectCount)
		require.Equal(t, 1, got.LessonCount, "chapter_count 列へ加算される")
		require.Equal(t, 1, got.NoteCount)
		require.Equal(t, day1, got.ActivityDate.UTC())
	})

	t.Run("別 user の行は混ざらない", func(t *testing.T) {
		rows, err := repo.ListByUser(ctx, 2, day1, day1)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, 9, rows[0].ExerciseCount)
	})

	t.Run("時刻成分は日付へ切り詰められて同一行に集約される", func(t *testing.T) {
		// 同じ日の別時刻で加算しても day1 の行へ入る（新しい行を作らない）。
		withTime := time.Date(2026, 1, 10, 13, 45, 0, 0, time.UTC)
		require.NoError(t, repo.Increment(ctx, 1, withTime, repository.UserDailyActivityIncrement{NoteCount: 2}))
		rows, err := repo.ListByUser(ctx, 1, day1, day1)
		require.NoError(t, err)
		require.Len(t, rows, 1, "時刻違いでも新しい行は増えない")
		require.Equal(t, 3, rows[0].NoteCount, "1 + 2")
	})
}

// TestUserDailyActivityRepository_ListByUser_Integration は範囲取得の境界と昇順を固定する。
func TestUserDailyActivityRepository_ListByUser_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserDailyActivityRepository(sqlDB)
	ctx := context.Background()

	testsupport.TruncateAll(t, sqlDB, "user_daily_activities")

	d10 := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	d11 := time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC)
	d12 := time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC)

	// 投入順を降順にして、返却が activity_date 昇順で固定されることを確かめる。
	require.NoError(t, repo.Increment(ctx, 1, d12, repository.UserDailyActivityIncrement{NoteCount: 1}))
	require.NoError(t, repo.Increment(ctx, 1, d10, repository.UserDailyActivityIncrement{NoteCount: 1}))
	require.NoError(t, repo.Increment(ctx, 1, d11, repository.UserDailyActivityIncrement{NoteCount: 1}))

	t.Run("範囲内を activity_date 昇順で返す", func(t *testing.T) {
		rows, err := repo.ListByUser(ctx, 1, d10, d12)
		require.NoError(t, err)
		require.Len(t, rows, 3)
		require.Equal(t, d10, rows[0].ActivityDate.UTC())
		require.Equal(t, d11, rows[1].ActivityDate.UTC())
		require.Equal(t, d12, rows[2].ActivityDate.UTC())
	})

	t.Run("BETWEEN は両端を含む", func(t *testing.T) {
		rows, err := repo.ListByUser(ctx, 1, d11, d11)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, d11, rows[0].ActivityDate.UTC())
	})

	t.Run("範囲外は含めない", func(t *testing.T) {
		rows, err := repo.ListByUser(ctx, 1, d10, d11)
		require.NoError(t, err)
		require.Len(t, rows, 2, "d12 は範囲外")
	})

	t.Run("該当なしは空配列（nil ではない）", func(t *testing.T) {
		rows, err := repo.ListByUser(ctx, 999, d10, d12)
		require.NoError(t, err)
		require.NotNil(t, rows)
		require.Empty(t, rows)
	})
}
