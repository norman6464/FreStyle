package usecase

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetUserDashboardUseCase はパーソナライズダッシュボードに必要な集計データを返す。
//
// 返す情報:
//   - streak        : 今日を起点とした連続学習日数（user_daily_activities ベース）
//   - recentActivity: 過去 90 日分の日次サマリー（カレンダーヒートマップ用）
//   - chapterViews  : 直近に開いた章 5 件（「続きから」カード用）
type GetUserDashboardUseCase struct {
	activity     repository.UserDailyActivityRepository
	chapterViews repository.UserChapterViewRepository
}

func NewGetUserDashboardUseCase(
	a repository.UserDailyActivityRepository,
	cv repository.UserChapterViewRepository,
) *GetUserDashboardUseCase {
	return &GetUserDashboardUseCase{activity: a, chapterViews: cv}
}

// GetUserDashboardOutput はダッシュボード API のレスポンス型。
type GetUserDashboardOutput struct {
	Streak             int                        `json:"streak"`
	TotalExercises     int                        `json:"totalExercises"`
	TotalCorrect       int                        `json:"totalCorrect"`
	TotalLessons       int                        `json:"totalLessons"`
	RecentActivity     []domain.UserDailyActivity `json:"recentActivity"`
	RecentChapterViews []domain.UserChapterView   `json:"recentChapterViews"`
}

func (u *GetUserDashboardUseCase) Execute(ctx context.Context, userID uint64) (*GetUserDashboardOutput, error) {
	now := time.Now().UTC()
	// 過去 90 日分を取得してカレンダー表示と streak 計算を両立する。
	from := now.AddDate(0, 0, -89).Truncate(24 * time.Hour)

	activities, err := u.activity.ListByUser(ctx, userID, from, now)
	if err != nil {
		return nil, err
	}

	views, err := u.chapterViews.ListRecentByUser(ctx, userID, 5)
	if err != nil {
		return nil, err
	}

	out := &GetUserDashboardOutput{
		Streak:             computeStreak(activities, now),
		RecentActivity:     activities,
		RecentChapterViews: views,
	}
	for _, a := range activities {
		out.TotalExercises += a.ExerciseCount
		out.TotalCorrect += a.CorrectCount
		out.TotalLessons += a.LessonCount
	}
	return out, nil
}

// computeStreak は今日から遡って何日連続で学習したかを返す。
// 学習あり = ExerciseCount + LessonCount + AiChatCount + NoteCount のいずれかが 1 以上の日。
func computeStreak(activities []domain.UserDailyActivity, now time.Time) int {
	// date → activity のマップを作る。
	actMap := make(map[string]bool, len(activities))
	for _, a := range activities {
		if a.ExerciseCount+a.LessonCount+a.AiChatCount+a.NoteCount > 0 {
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
