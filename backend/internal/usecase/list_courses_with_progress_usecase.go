package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CourseWithProgress はコース一覧用にコースへ章数と actor 自身の完了章数を合成した出力。
// 埋め込みにより JSON は Course のフィールドへ materialCount / completedCount が加わったフラットな形になる。
type CourseWithProgress struct {
	domain.Course
	// MaterialCount はコース内の章数。trainee は published のみ、admin 系は下書き込み。
	MaterialCount int `json:"materialCount"`
	// CompletedCount は actor 自身が完了した章数(現存する published 章のみ。常に MaterialCount 以下)。
	CompletedCount int `json:"completedCount"`
}

// ListCoursesWithProgressUseCase はコース一覧に章数と完了章数を付けて返す。
// 分母(章数)は trainee=published のみ / admin=下書き込み。分子(完了章数)は現存する published
// 章の完了行のみを数え、コース詳細ページの進捗バーと同じ意味論にする。
// 管理ロールは完了記録を持たない(UI も進捗を出さない)ため分子の集計をスキップする。
type ListCoursesWithProgressUseCase struct {
	courses   repository.CourseRepository
	materials repository.TeachingMaterialRepository
	progress  repository.LessonProgressRepository
}

// NewListCoursesWithProgressUseCase は ListCoursesWithProgressUseCase を組み立てる。
func NewListCoursesWithProgressUseCase(
	courses repository.CourseRepository,
	materials repository.TeachingMaterialRepository,
	progress repository.LessonProgressRepository,
) *ListCoursesWithProgressUseCase {
	return &ListCoursesWithProgressUseCase{courses: courses, materials: materials, progress: progress}
}

// ListCoursesWithProgressInput は一覧取得の actor 情報(認証 context 由来)。
type ListCoursesWithProgressInput struct {
	ActorUserID    uint64
	ActorCompanyID uint64
	ActorRole      string
}

// Execute はコース一覧を返す。ActorCompanyID=0(会社未所属)は空スライス。
func (u *ListCoursesWithProgressUseCase) Execute(ctx context.Context, in ListCoursesWithProgressInput) ([]CourseWithProgress, error) {
	if in.ActorCompanyID == 0 {
		return []CourseWithProgress{}, nil
	}
	includeUnpublished := canManage(in.ActorRole)
	rows, err := u.courses.ListByCompany(ctx, in.ActorCompanyID, includeUnpublished)
	if err != nil {
		return nil, err
	}
	materialCounts, err := u.materials.CountByCourseForCompany(ctx, in.ActorCompanyID, includeUnpublished)
	if err != nil {
		return nil, err
	}
	completedCounts := map[uint64]int{}
	if !includeUnpublished {
		// 完了記録を持つのは受講者のみ。管理ロールでは 1 クエリ節約する。
		completedCounts, err = u.progress.CountCompletedByUserGroupedByCourse(ctx, in.ActorUserID)
		if err != nil {
			return nil, err
		}
	}
	// 0 件時も JSON で null にならないよう必ず空スライスを返す(FRESTYLE-70 と同じ理由)。
	out := make([]CourseWithProgress, 0, len(rows))
	for _, c := range rows {
		out = append(out, CourseWithProgress{
			Course:         c,
			MaterialCount:  materialCounts[c.ID],
			CompletedCount: completedCounts[c.ID],
		})
	}
	return out, nil
}
