package usecase

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetUserDashboardUseCase はパーソナライズダッシュボードに必要な集計データを返す。
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
