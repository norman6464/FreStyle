package persistence

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

type userDailyActivityRepository struct {
	db *gorm.DB
}

// NewUserDailyActivityRepository は UserDailyActivityRepository の GORM 実装を返す。
func NewUserDailyActivityRepository(db *gorm.DB) repository.UserDailyActivityRepository {
	return &userDailyActivityRepository{db: db}
}

// Increment は user_daily_activities を upsert し各カウンタを delta 分だけ加算する。
// PostgreSQL ON CONFLICT DO UPDATE で原子的に実行するため、アプリ側でのロックは不要。
func (r *userDailyActivityRepository) Increment(
	ctx context.Context,
	userID uint64,
	date time.Time,
	delta repository.UserDailyActivityIncrement,
) error {
	// date を DATE 型へ切り詰め（時刻成分を捨てる）。
	d := date.UTC().Truncate(24 * time.Hour)
	// Expand フェーズ: 旧 lesson_count と新 chapter_count の両方へ同じ増分を加算する(dual-write)。
	// 読みは当面 lesson_count。Contract で chapter_count へ切り替えて旧列を削除する。
	sql := `
INSERT INTO user_daily_activities
  (user_id, activity_date, exercise_count, correct_count, lesson_count, chapter_count, ai_chat_count, note_count)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (user_id, activity_date) DO UPDATE SET
  exercise_count = user_daily_activities.exercise_count + EXCLUDED.exercise_count,
  correct_count  = user_daily_activities.correct_count  + EXCLUDED.correct_count,
  lesson_count   = user_daily_activities.lesson_count   + EXCLUDED.lesson_count,
  chapter_count  = user_daily_activities.chapter_count  + EXCLUDED.chapter_count,
  ai_chat_count  = user_daily_activities.ai_chat_count  + EXCLUDED.ai_chat_count,
  note_count     = user_daily_activities.note_count     + EXCLUDED.note_count
`
	return r.db.WithContext(ctx).Exec(
		sql,
		userID, d,
		delta.ExerciseCount,
		delta.CorrectCount,
		delta.LessonCount,
		delta.LessonCount,
		delta.AiChatCount,
		delta.NoteCount,
	).Error
}

func (r *userDailyActivityRepository) ListByUser(
	ctx context.Context,
	userID uint64,
	from, to time.Time,
) ([]domain.UserDailyActivity, error) {
	var rows []domain.UserDailyActivity
	fromDate := from.UTC().Truncate(24 * time.Hour)
	toDate := to.UTC().Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND activity_date BETWEEN ? AND ?", userID, fromDate, toDate).
		Order("activity_date ASC").
		Find(&rows).Error
	return rows, err
}
