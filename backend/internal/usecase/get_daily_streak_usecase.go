package usecase

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetDailyStreakUseCase は user_daily_activities から連続学習日数の統計を算出する。
// 設定画面のプロフィール統計 (GET /daily-goals/streak) 用。
type GetDailyStreakUseCase struct {
	activity repository.UserDailyActivityRepository
}

func NewGetDailyStreakUseCase(a repository.UserDailyActivityRepository) *GetDailyStreakUseCase {
	return &GetDailyStreakUseCase{activity: a}
}

// GetDailyStreakOutput は streak API のレスポンス型。JSON キーはフロント
// (entities/user の ProfileStats) の契約に合わせる。
type GetDailyStreakOutput struct {
	CurrentStreak     int `json:"currentStreak"`
	LongestStreak     int `json:"longestStreak"`
	TotalAchievedDays int `json:"totalAchievedDays"`
}

// dailyActivityEpoch は全期間集計の下限。サービス開始 (2026-04) より十分過去に置く。
var dailyActivityEpoch = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func (u *GetDailyStreakUseCase) Execute(ctx context.Context, userID uint64) (*GetDailyStreakOutput, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	now := time.Now().UTC()
	// 最長連続・累計は全履歴が必要なため全期間を取得する（1 user 1 日 1 行なので高々数百行）。
	activities, err := u.activity.ListByUser(ctx, userID, dailyActivityEpoch, now)
	if err != nil {
		return nil, err
	}

	// 学習ありの日を日単位で集める（判定条件は computeStreak（下記）と同じ）。
	daySet := make(map[string]time.Time, len(activities))
	for _, a := range activities {
		if a.ExerciseCount+a.LessonCount+a.NoteCount > 0 {
			d := a.ActivityDate.UTC().Truncate(24 * time.Hour)
			daySet[d.Format("2006-01-02")] = d
		}
	}
	days := make([]time.Time, 0, len(daySet))
	for _, d := range daySet {
		days = append(days, d)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })

	longest, run := 0, 0
	for i, d := range days {
		if i > 0 && days[i-1].AddDate(0, 0, 1).Equal(d) {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}

	return &GetDailyStreakOutput{
		CurrentStreak:     computeStreak(activities, now),
		LongestStreak:     longest,
		TotalAchievedDays: len(days),
	}, nil
}

// computeStreak は今日を起点に、活動が途切れるまで遡って連続学習日数を数える。
// 学習あり = ExerciseCount / LessonCount / NoteCount のいずれかが 1 以上。
func computeStreak(activities []domain.UserDailyActivity, now time.Time) int {
	actMap := make(map[string]bool, len(activities))
	for _, a := range activities {
		if a.ExerciseCount+a.LessonCount+a.NoteCount > 0 {
			actMap[a.ActivityDate.UTC().Format("2006-01-02")] = true
		}
	}
	streak := 0
	today := now.UTC().Truncate(24 * time.Hour)
	for d := today; ; d = d.AddDate(0, 0, -1) {
		if !actMap[d.Format("2006-01-02")] {
			break
		}
		streak++
	}
	return streak
}
