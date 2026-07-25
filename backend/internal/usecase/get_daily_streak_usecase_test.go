package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
)

// fakeDailyActivityRepo は UserDailyActivityRepository の in-memory fake。
type fakeDailyActivityRepo struct {
	activities []domain.UserDailyActivity
}

func (f *fakeDailyActivityRepo) Increment(_ context.Context, _ uint64, _ time.Time, _ repository.UserDailyActivityIncrement) error {
	return nil
}

func (f *fakeDailyActivityRepo) ListByUser(_ context.Context, _ uint64, _, _ time.Time) ([]domain.UserDailyActivity, error) {
	return f.activities, nil
}

func day(offset int) time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, offset)
}

func act(offset, exercises int) domain.UserDailyActivity {
	return domain.UserDailyActivity{ActivityDate: day(offset), ExerciseCount: exercises}
}

func Test_連続学習統計_今日まで連続していればcurrentStreakに数える(t *testing.T) {
	uc := NewGetDailyStreakUseCase(&fakeDailyActivityRepo{activities: []domain.UserDailyActivity{
		act(-2, 1), act(-1, 1), act(0, 1),
	}})
	out, err := uc.Execute(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 3, out.CurrentStreak)
	assert.Equal(t, 3, out.LongestStreak)
	assert.Equal(t, 3, out.TotalAchievedDays)
}

func Test_連続学習統計_途切れた過去の連続はlongestStreakにだけ残る(t *testing.T) {
	// 5 日前〜3 日前の 3 連続 + 今日のみ → current=1, longest=3, total=4
	uc := NewGetDailyStreakUseCase(&fakeDailyActivityRepo{activities: []domain.UserDailyActivity{
		act(-5, 1), act(-4, 1), act(-3, 1), act(0, 1),
	}})
	out, err := uc.Execute(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 1, out.CurrentStreak)
	assert.Equal(t, 3, out.LongestStreak)
	assert.Equal(t, 4, out.TotalAchievedDays)
}

func Test_連続学習統計_全カウンタ0の行は学習日に数えない(t *testing.T) {
	uc := NewGetDailyStreakUseCase(&fakeDailyActivityRepo{activities: []domain.UserDailyActivity{
		act(-1, 0), act(0, 1),
	}})
	out, err := uc.Execute(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 1, out.CurrentStreak)
	assert.Equal(t, 1, out.LongestStreak)
	assert.Equal(t, 1, out.TotalAchievedDays)
}

func Test_連続学習統計_活動なしは全て0(t *testing.T) {
	uc := NewGetDailyStreakUseCase(&fakeDailyActivityRepo{})
	out, err := uc.Execute(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 0, out.CurrentStreak)
	assert.Equal(t, 0, out.LongestStreak)
	assert.Equal(t, 0, out.TotalAchievedDays)
}

func Test_連続学習統計_userID必須(t *testing.T) {
	uc := NewGetDailyStreakUseCase(&fakeDailyActivityRepo{})
	_, err := uc.Execute(context.Background(), 0)
	assert.Error(t, err)
}
