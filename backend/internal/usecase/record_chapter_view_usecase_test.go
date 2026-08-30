package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_章閲覧記録_自社の公開教材はupsertする(t *testing.T) {
	mat, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 5, CompanyID: 10, CourseID: 99, WorkspaceID: strPtr(wsA), IsPublished: true,
	}})
	crs, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 99, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: true}})
	views := &mockChapterViewRepo{}
	views.On("UpsertView", mock.Anything, uint64(1), uint64(5), uint64(99)).Return(nil)
	uc := usecase.NewRecordChapterViewUseCase(views, mat, crs)

	err := uc.Execute(context.Background(), usecase.RecordChapterViewInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	require.NoError(t, err)
	views.AssertExpectations(t)
}

func Test_章閲覧記録_他社の教材はforbidden(t *testing.T) {
	mat, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 5, CompanyID: 10, CourseID: 99, WorkspaceID: strPtr(wsA), IsPublished: true,
	}})
	crs, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 99, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: true}})
	views := &mockChapterViewRepo{}
	uc := usecase.NewRecordChapterViewUseCase(views, mat, crs)

	err := uc.Execute(context.Background(), usecase.RecordChapterViewInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsB), ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	require.ErrorIs(t, err, usecase.ErrChapterViewForbidden)
	views.AssertNotCalled(t, "UpsertView", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_章閲覧記録_存在しない教材はNotFound(t *testing.T) {
	mat, _ := materialRepo(materialFakeConfig{getErr: domain.ErrNotFound})
	crs, _ := courseRepo(courseFakeConfig{})
	views := &mockChapterViewRepo{}
	uc := usecase.NewRecordChapterViewUseCase(views, mat, crs)

	err := uc.Execute(context.Background(), usecase.RecordChapterViewInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleTrainee, TeachingMaterialID: 404,
	})
	require.ErrorIs(t, err, domain.ErrNotFound)
}
