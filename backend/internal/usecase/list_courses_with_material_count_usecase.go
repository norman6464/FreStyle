package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CourseWithMaterialCount はコース一覧用にコースへ教材(章)数を合成した出力。
// 埋め込みにより JSON は Course のフィールドへ materialCount が加わったフラットな形になる。
type CourseWithMaterialCount struct {
	domain.Course
	MaterialCount int `json:"materialCount"`
}

// ListCoursesWithMaterialCountUseCase はコース一覧に章数(教材数)を付けて返す。
// trainee は published のみ(一覧・章数とも)で、コース詳細の進捗バーの分母と一致させる。
type ListCoursesWithMaterialCountUseCase struct {
	courses   repository.CourseRepository
	materials repository.TeachingMaterialRepository
}

func NewListCoursesWithMaterialCountUseCase(courses repository.CourseRepository, materials repository.TeachingMaterialRepository) *ListCoursesWithMaterialCountUseCase {
	return &ListCoursesWithMaterialCountUseCase{courses: courses, materials: materials}
}

type ListCoursesWithMaterialCountInput struct {
	ActorCompanyID uint64
	ActorRole      string
}

func (u *ListCoursesWithMaterialCountUseCase) Execute(ctx context.Context, in ListCoursesWithMaterialCountInput) ([]CourseWithMaterialCount, error) {
	if in.ActorCompanyID == 0 {
		return []CourseWithMaterialCount{}, nil
	}
	includeUnpublished := canManage(in.ActorRole)
	rows, err := u.courses.ListByCompany(ctx, in.ActorCompanyID, includeUnpublished)
	if err != nil {
		return nil, err
	}
	counts, err := u.materials.CountByCourseForCompany(ctx, in.ActorCompanyID, includeUnpublished)
	if err != nil {
		return nil, err
	}
	// 0 件時も JSON で null にならないよう必ず空スライスを返す(FRESTYLE-70 と同じ理由)。
	out := make([]CourseWithMaterialCount, 0, len(rows))
	for _, c := range rows {
		out = append(out, CourseWithMaterialCount{Course: c, MaterialCount: counts[c.ID]})
	}
	return out, nil
}
