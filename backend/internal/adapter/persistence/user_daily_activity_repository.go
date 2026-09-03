package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// userDailyActivityRepository は [repository.UserDailyActivityRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type userDailyActivityRepository struct {
	db *sql.DB
}

// NewUserDailyActivityRepository は UserDailyActivityRepository の実装を返す。
func NewUserDailyActivityRepository(db *sql.DB) repository.UserDailyActivityRepository {
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
	uid, ok := toInt64ID(userID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("user_id", userID)
	}
	// date を DATE 型へ切り詰め（時刻成分を捨てる）。
	d := date.UTC().Truncate(24 * time.Hour)
	return sqlcgen.New(r.db).IncrementUserDailyActivity(ctx, sqlcgen.IncrementUserDailyActivityParams{
		UserID:        uid,
		ActivityDate:  d,
		ExerciseCount: int32(delta.ExerciseCount),
		CorrectCount:  int32(delta.CorrectCount),
		ChapterCount:  int32(delta.LessonCount),
		NoteCount:     int32(delta.NoteCount),
	})
}
