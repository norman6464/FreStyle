package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// RecordChapterViewUseCase はユーザーが章（教材）を開いたときに閲覧記録を残す。
// 「続きから」カードや閲覧履歴の基盤となる。完了（MarkLessonCompleted）とは別に、
// ページを開いただけでも記録するため、離脱した章も追跡できる。
// 読めるかは対象ごとの付与で決める（他テナントの教材への不正記録もここで塞がる）。
type RecordChapterViewUseCase struct {
	chapterViews repository.UserChapterViewRepository
	materials    repository.TeachingMaterialRepository
	perm         *CheckMaterialPermissionUseCase
}

func NewRecordChapterViewUseCase(
	cv repository.UserChapterViewRepository,
	m repository.TeachingMaterialRepository,
	perm *CheckMaterialPermissionUseCase,
) *RecordChapterViewUseCase {
	return &RecordChapterViewUseCase{chapterViews: cv, materials: m, perm: perm}
}

// RecordChapterViewInput は章閲覧記録の入力。actor のワークスペースで可視性を検証する。
type RecordChapterViewInput struct {
	UserID             uint64
	ActorWorkspace     domain.WorkspaceRef
	TeachingMaterialID uint64
}

// ErrChapterViewForbidden は閲覧権限のない教材を記録しようとしたときに返す。
var ErrChapterViewForbidden = errors.New("chapter_view_forbidden")

// Execute は course_id を教材から解決し、canRead を検証してから upsert する。
// エラー時は呼び出し元の処理を止めない想定（handler 側でベストエフォート扱い）。
func (u *RecordChapterViewUseCase) Execute(ctx context.Context, in RecordChapterViewInput) error {
	m, err := u.materials.GetByID(ctx, in.TeachingMaterialID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrLessonNotFound
	}
	workspaceID, affiliated := in.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return ErrChapterViewForbidden
	}
	perm, err := u.perm.Chapter(ctx, workspaceID, in.TeachingMaterialID, in.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrChapterViewForbidden
		}
		return err
	}
	if !perm.CanView {
		return ErrChapterViewForbidden
	}
	return u.chapterViews.UpsertView(ctx, in.UserID, in.TeachingMaterialID, m.CourseID)
}
