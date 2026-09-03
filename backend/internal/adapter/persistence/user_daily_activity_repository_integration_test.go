//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// dailyActivityRow は検証用に user_daily_activities を直接読む1行分。
type dailyActivityRow struct {
	exerciseCount int
	correctCount  int
	chapterCount  int
	noteCount     int
}

func queryDailyActivityRow(t *testing.T, ctx context.Context, sqlDB *sql.DB, userID uint64, date time.Time) (dailyActivityRow, bool) {
	t.Helper()
	var row dailyActivityRow
	err := sqlDB.QueryRowContext(
		ctx,
		"SELECT exercise_count, correct_count, chapter_count, note_count FROM user_daily_activities WHERE user_id = $1 AND activity_date = $2",
		userID, date,
	).Scan(&row.exerciseCount, &row.correctCount, &row.chapterCount, &row.noteCount)
	if err != nil {
		return dailyActivityRow{}, false
	}
	return row, true
}

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
		row, ok := queryDailyActivityRow(t, ctx, sqlDB, 1, day1)
		require.True(t, ok)
		require.Equal(t, 3, row.exerciseCount, "1 + 2")
		require.Equal(t, 1, row.correctCount)
		require.Equal(t, 1, row.chapterCount, "chapter_count 列へ加算される")
		require.Equal(t, 1, row.noteCount)
	})

	t.Run("別日は別行", func(t *testing.T) {
		row, ok := queryDailyActivityRow(t, ctx, sqlDB, 1, day2)
		require.True(t, ok)
		require.Equal(t, 5, row.noteCount)
	})

	t.Run("別 user の行は混ざらない", func(t *testing.T) {
		row, ok := queryDailyActivityRow(t, ctx, sqlDB, 2, day1)
		require.True(t, ok)
		require.Equal(t, 9, row.exerciseCount)
	})

	t.Run("時刻成分は日付へ切り詰められて同一行に集約される", func(t *testing.T) {
		// 同じ日の別時刻で加算しても day1 の行へ入る（新しい行を作らない）。
		withTime := time.Date(2026, 1, 10, 13, 45, 0, 0, time.UTC)
		require.NoError(t, repo.Increment(ctx, 1, withTime, repository.UserDailyActivityIncrement{NoteCount: 2}))
		row, ok := queryDailyActivityRow(t, ctx, sqlDB, 1, day1)
		require.True(t, ok)
		require.Equal(t, 3, row.noteCount, "1 + 2")
	})
}
