package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_章閲覧記録_読める教材はupsertする(t *testing.T) {
	mat, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 5, CourseID: 99, WorkspaceID: strPtr(wsA), IsPublished: true,
	}})
	views := &mockChapterViewRepo{}
	views.On("UpsertView", mock.Anything, uint64(1), uint64(5), uint64(99)).Return(nil)
	_, perm := materialPerm(materialFactsConfig{member: true, published: true})
	uc := usecase.NewRecordChapterViewUseCase(views, mat, perm)

	err := uc.Execute(context.Background(), usecase.RecordChapterViewInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), TeachingMaterialID: 5,
	})
	require.NoError(t, err)
	views.AssertExpectations(t)
}

func Test_章閲覧記録_読めない教材は記録しない(t *testing.T) {
	// 閲覧記録は「開いた」ことの記録なので、開けない教材に残ってはいけない
	// （残ると「続きから」に他人の教材が並ぶ）。
	for _, c := range []struct {
		name string
		cfg  materialFactsConfig
	}{
		{"別テナントの教材", materialFactsConfig{notFound: true}},
		{"付与の無い下書き", materialFactsConfig{member: true, published: false}},
		{"所属していない", materialFactsConfig{member: false, published: true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			mat, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
				ID: 5, CourseID: 99, WorkspaceID: strPtr(wsA), IsPublished: true,
			}})
			views := &mockChapterViewRepo{}
			_, perm := materialPerm(c.cfg)
			uc := usecase.NewRecordChapterViewUseCase(views, mat, perm)

			err := uc.Execute(context.Background(), usecase.RecordChapterViewInput{
				UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), TeachingMaterialID: 5,
			})
			require.ErrorIs(t, err, usecase.ErrChapterViewForbidden)
			views.AssertNotCalled(t, "UpsertView", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func Test_章閲覧記録_存在しない教材はNotFound(t *testing.T) {
	mat, _ := materialRepo(materialFakeConfig{getErr: domain.ErrNotFound})
	views := &mockChapterViewRepo{}
	_, perm := materialPerm(materialFactsConfig{member: true, published: true})
	uc := usecase.NewRecordChapterViewUseCase(views, mat, perm)

	err := uc.Execute(context.Background(), usecase.RecordChapterViewInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), TeachingMaterialID: 404,
	})
	require.ErrorIs(t, err, domain.ErrNotFound)
}
