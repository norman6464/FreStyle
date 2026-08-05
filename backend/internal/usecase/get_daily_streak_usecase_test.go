package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// dailyActivityRepo は指定の活動履歴を返す mock を組み立てる。
func dailyActivityRepo(activities ...domain.UserDailyActivity) *mockDailyActivityRepo {
	repo := &mockDailyActivityRepo{}
	repo.On("ListByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(activities, nil).Maybe()
	return repo
}

func streakDay(offset int) time.Time {
	return time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, offset)
}

func streakAct(offset, exercises int) domain.UserDailyActivity {
	return domain.UserDailyActivity{ActivityDate: streakDay(offset), ExerciseCount: exercises}
}

func Test_連続学習統計_今日まで連続していればcurrentStreakに数える(t *testing.T) {
	uc := usecase.NewGetDailyStreakUseCase(dailyActivityRepo(streakAct(-2, 1), streakAct(-1, 1), streakAct(0, 1)))
	out, err := uc.Execute(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 3, out.CurrentStreak)
	assert.Equal(t, 3, out.LongestStreak)
	assert.Equal(t, 3, out.TotalAchievedDays)
}

func Test_連続学習統計_途切れた過去の連続はlongestStreakにだけ残る(t *testing.T) {
	// 5 日前〜3 日前の 3 連続 + 今日のみ → current=1, longest=3, total=4
	uc := usecase.NewGetDailyStreakUseCase(dailyActivityRepo(streakAct(-5, 1), streakAct(-4, 1), streakAct(-3, 1), streakAct(0, 1)))
	out, err := uc.Execute(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 1, out.CurrentStreak)
	assert.Equal(t, 3, out.LongestStreak)
	assert.Equal(t, 4, out.TotalAchievedDays)
}

func Test_連続学習統計_全カウンタ0の行は学習日に数えない(t *testing.T) {
	uc := usecase.NewGetDailyStreakUseCase(dailyActivityRepo(streakAct(-1, 0), streakAct(0, 1)))
	out, err := uc.Execute(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 1, out.CurrentStreak)
	assert.Equal(t, 1, out.LongestStreak)
	assert.Equal(t, 1, out.TotalAchievedDays)
}

func Test_連続学習統計_活動なしは全て0(t *testing.T) {
	uc := usecase.NewGetDailyStreakUseCase(dailyActivityRepo())
	out, err := uc.Execute(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, 0, out.CurrentStreak)
	assert.Equal(t, 0, out.LongestStreak)
	assert.Equal(t, 0, out.TotalAchievedDays)
}

func Test_連続学習統計_userID必須(t *testing.T) {
	repo := dailyActivityRepo()
	uc := usecase.NewGetDailyStreakUseCase(repo)
	_, err := uc.Execute(context.Background(), 0)
	assert.Error(t, err)
	// 入口で弾いているので repository には到達しない。
	repo.AssertNotCalled(t, "ListByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
