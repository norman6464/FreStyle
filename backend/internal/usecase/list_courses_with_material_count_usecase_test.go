package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_コース一覧章数付き_各コースに章数が合成される(t *testing.T) {
	crepo := &fakeCourseRepo{rows: []domain.Course{{ID: 1, Title: "Git"}, {ID: 2, Title: "Docker"}}}
	mrepo := &fakeTeachingMaterialRepo{countsByCourse: map[uint64]int{1: 3, 2: 12}}
	uc := usecase.NewListCoursesWithMaterialCountUseCase(crepo, mrepo)

	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithMaterialCountInput{
		ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
	})
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, uint64(1), out[0].ID)
	assert.Equal(t, 3, out[0].MaterialCount)
	assert.Equal(t, 12, out[1].MaterialCount)
}

func Test_コース一覧章数付き_集計に無いコースは0章(t *testing.T) {
	crepo := &fakeCourseRepo{rows: []domain.Course{{ID: 7, Title: "空のコース"}}}
	mrepo := &fakeTeachingMaterialRepo{countsByCourse: map[uint64]int{}}
	uc := usecase.NewListCoursesWithMaterialCountUseCase(crepo, mrepo)

	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithMaterialCountInput{
		ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, 0, out[0].MaterialCount)
}

func Test_コース一覧章数付き_会社未所属は空スライス(t *testing.T) {
	uc := usecase.NewListCoursesWithMaterialCountUseCase(&fakeCourseRepo{}, &fakeTeachingMaterialRepo{})
	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithMaterialCountInput{
		ActorCompanyID: 0, ActorRole: domain.RoleSuperAdmin,
	})
	require.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out)
}

func Test_コース一覧章数付き_0件でもnilではなく空スライス(t *testing.T) {
	// GORM の Find は 0 件時に nil スライスを返し JSON で null になるため正規化する(FRESTYLE-70)。
	crepo := &fakeCourseRepo{rows: nil}
	uc := usecase.NewListCoursesWithMaterialCountUseCase(crepo, &fakeTeachingMaterialRepo{})
	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithMaterialCountInput{
		ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
	})
	require.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out)
}

func Test_コース一覧章数付き_traineeは公開のみで集計(t *testing.T) {
	crepo := &fakeCourseRepo{rows: []domain.Course{{ID: 1}}}
	mrepo := &fakeTeachingMaterialRepo{countsByCourse: map[uint64]int{1: 5}}
	uc := usecase.NewListCoursesWithMaterialCountUseCase(crepo, mrepo)

	_, err := uc.Execute(context.Background(), usecase.ListCoursesWithMaterialCountInput{
		ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
	})
	require.NoError(t, err)
	require.NotNil(t, mrepo.lastCountIncludeUnpublished)
	assert.False(t, *mrepo.lastCountIncludeUnpublished, "trainee の分母は published のみ")
}

func Test_コース一覧章数付き_管理者は下書き込みで集計(t *testing.T) {
	crepo := &fakeCourseRepo{rows: []domain.Course{{ID: 1}}}
	mrepo := &fakeTeachingMaterialRepo{countsByCourse: map[uint64]int{1: 5}}
	uc := usecase.NewListCoursesWithMaterialCountUseCase(crepo, mrepo)

	_, err := uc.Execute(context.Background(), usecase.ListCoursesWithMaterialCountInput{
		ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, mrepo.lastCountIncludeUnpublished)
	assert.True(t, *mrepo.lastCountIncludeUnpublished, "admin は下書き章も分母に含む")
}

func Test_コース一覧章数付き_集計エラーはそのまま返す(t *testing.T) {
	crepo := &fakeCourseRepo{rows: []domain.Course{{ID: 1}}}
	mrepo := &fakeTeachingMaterialRepo{countErr: context.DeadlineExceeded}
	uc := usecase.NewListCoursesWithMaterialCountUseCase(crepo, mrepo)

	_, err := uc.Execute(context.Background(), usecase.ListCoursesWithMaterialCountInput{
		ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
	})
	require.Error(t, err)
}
