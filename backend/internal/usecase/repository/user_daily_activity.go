package repository

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// UserDailyActivityIncrement は Increment 呼び出し1回分の加算量を表す。
// 0 のフィールドは SQL 上も 0 を加算するだけなので呼び出し側で気にしなくてよい。
type UserDailyActivityIncrement struct {
	ExerciseCount int
	CorrectCount  int
	LessonCount   int
	AiChatCount   int
	NoteCount     int
}

// UserDailyActivityRepository は日次学習サマリーの永続化 port。
type UserDailyActivityRepository interface {
	// Increment は (user_id, activity_date) の行を upsert し各カウンタに delta を加算する。
	// 行が存在しない場合は delta の値で INSERT する。
	Increment(ctx context.Context, userID uint64, date time.Time, delta UserDailyActivityIncrement) error

	// ListByUser は from〜to（UTC DATE）の範囲でサマリー一覧を返す。カレンダー表示用。
	ListByUser(ctx context.Context, userID uint64, from, to time.Time) ([]domain.UserDailyActivity, error)
}
