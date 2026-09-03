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
	MaterialCount  int  `json:"materialCount"`
	CompletedCount int  `json:"completedCount"`
	CanEdit        bool `json:"canEdit"`
	CanManage      bool `json:"canManage"`
}

// ListCoursesWithProgressUseCase はコース一覧に章数と完了章数を付けて返す。
type ListCoursesWithProgressUseCase struct {
	materials repository.TeachingMaterialRepository
	progress  repository.LessonProgressRepository
	perm      repository.MaterialPermissionRepository
}

// NewListCoursesWithProgressUseCase は ListCoursesWithProgressUseCase を組み立てる。
func NewListCoursesWithProgressUseCase(
	materials repository.TeachingMaterialRepository,
	progress repository.LessonProgressRepository,
	perm repository.MaterialPermissionRepository,
) *ListCoursesWithProgressUseCase {
	return &ListCoursesWithProgressUseCase{materials: materials, progress: progress, perm: perm}
}

// ListCoursesWithProgressInput は一覧取得の actor 情報（認証 context 由来）。
type ListCoursesWithProgressInput struct {
	MaterialActor
}

// Execute はコース一覧を返す。ワークスペース未所属の actor は(super_admin でも)空スライス。
func (u *ListCoursesWithProgressUseCase) Execute(ctx context.Context, in ListCoursesWithProgressInput) ([]CourseWithProgress, error) {
	// 自社のコースを並べる画面なので、所属ワークスペースが無ければ数える対象そのものが無い。
	workspaceID, affiliated := in.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return []CourseWithProgress{}, nil
	}
	facts, err := u.perm.ListCourseFactsForUser(ctx, workspaceID, in.ActorUserID)
	if err != nil {
		return nil, err
	}

	rows := make([]domain.Course, 0, len(facts))
	editable := make(map[uint64]bool, len(facts))
	manageable := make(map[uint64]bool, len(facts))
	for _, f := range facts {
		perm := domain.ResolveMaterialPermission(f.Facts)
		if !perm.CanView {
			continue
		}
		rows = append(rows, f.Course)
		editable[f.Course.ID] = perm.CanEdit
		manageable[f.Course.ID] = perm.CanManage
	}

	publishedCounts, err := u.materials.CountByCourseForWorkspace(ctx, workspaceID, false)
	if err != nil {
		return nil, err
	}
	allCounts := publishedCounts
	if anyEditable(editable) {
		allCounts, err = u.materials.CountByCourseForWorkspace(ctx, workspaceID, true)
		if err != nil {
			return nil, err
		}
	}
	completedCounts, err := u.progress.CountCompletedByUserGroupedByCourse(ctx, in.ActorUserID)
	if err != nil {
		return nil, err
	}
	out := make([]CourseWithProgress, 0, len(rows))
	for _, c := range rows {
		count := publishedCounts[c.ID]
		if editable[c.ID] {
			count = allCounts[c.ID]
		}
		out = append(out, CourseWithProgress{
			Course:         c,
			MaterialCount:  count,
			CompletedCount: completedCounts[c.ID],
			CanEdit:        editable[c.ID],
			CanManage:      manageable[c.ID],
		})
	}
	return out, nil
}

// anyEditable は編集できるコースが 1 つでもあるかを返す。
func anyEditable(editable map[uint64]bool) bool {
	for _, ok := range editable {
		if ok {
			return true
		}
	}
	return false
}
