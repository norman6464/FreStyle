package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

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

type GetLastViewedChapterInput struct {
	MaterialActor
	CourseID uint64
}

func (u *GetLastViewedChapterUseCase) Execute(ctx context.Context, in GetLastViewedChapterInput) (*domain.UserChapterView, error) {
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
