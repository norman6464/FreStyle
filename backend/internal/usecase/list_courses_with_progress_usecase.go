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
//
// **見せるコースは対象ごとの付与で決まる。** 事実をまとめて引き（コースごとに引かない）、
// ふるい落としは domain.ResolveMaterialPermission に通す。
//
// 分子(完了章数)は現存する published 章の完了行のみを数え、コース詳細ページの進捗バーと
// 同じ意味論にする。
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
	// 見せてよいコースだけに絞る。規則は domain が持つので、ここでは答えを使うだけ。
	rows := make([]domain.Course, 0, len(facts))
	// 下書きの章まで数えてよいのは、そのコースを編集できる人だけ。1 つでもあれば
	// 章数の集計は下書き込みで引き、コースごとに数え直さない。
	includeUnpublished := false
	for _, f := range facts {
		perm := domain.ResolveMaterialPermission(f.Facts)
		if !perm.CanView {
			continue
		}
		rows = append(rows, f.Course)
		if perm.CanEdit {
			includeUnpublished = true
		}
	}
	materialCounts, err := u.materials.CountByCourseForWorkspace(ctx, workspaceID, includeUnpublished)
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
