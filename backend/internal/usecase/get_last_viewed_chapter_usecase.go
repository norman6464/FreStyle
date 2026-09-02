package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetLastViewedChapterUseCase はコース内で actor 自身が最後に閲覧した章の閲覧記録を返す。
// コース詳細を開いたときの「続きから表示」(FRESTYLE-99)用。履歴が無い場合は nil を返す(エラーではない)。
// コースを読めるかは対象ごとの付与で決める（CheckMaterialPermissionUseCase）。
type GetLastViewedChapterUseCase struct {
	chapterViews repository.UserChapterViewRepository
	perm         *CheckMaterialPermissionUseCase
}

// NewGetLastViewedChapterUseCase は GetLastViewedChapterUseCase を組み立てる。
func NewGetLastViewedChapterUseCase(
	chapterViews repository.UserChapterViewRepository,
	perm *CheckMaterialPermissionUseCase,
) *GetLastViewedChapterUseCase {
	return &GetLastViewedChapterUseCase{chapterViews: chapterViews, perm: perm}
}

// GetLastViewedChapterInput は取得対象コースと actor 情報(認証 context 由来)。
type GetLastViewedChapterInput struct {
	MaterialActor
	CourseID uint64
}

// Execute はコースの可視性を検証してから最終閲覧記録を返す。履歴なしは (nil, nil)。
func (u *GetLastViewedChapterUseCase) Execute(ctx context.Context, in GetLastViewedChapterInput) (*domain.UserChapterView, error) {
	// 読み取りの経路なので、見えない相手には実在を教えない（どの理由でも同じ ErrNotFound）。
	// 履歴の有無からコースの実在が読めてしまうのを塞ぐ。
	workspaceID, affiliated := in.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return nil, domain.ErrNotFound
	}
	perm, err := u.perm.Course(ctx, workspaceID, in.CourseID, in.ActorUserID)
	if err != nil {
		return nil, err
	}
	if !perm.CanView {
		return nil, domain.ErrNotFound
	}
	return u.chapterViews.GetLastViewedByUserAndCourse(ctx, in.ActorUserID, in.CourseID)
}
