package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_コース一覧進捗付き_各コースに章数と完了章数が合成される(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{rows: []domain.Course{{ID: 1, Title: "Git"}, {ID: 2, Title: "Docker"}}})
	mrepo, _ := materialRepo(materialFakeConfig{counts: map[uint64]int{1: 3, 2: 12}})
	prepo, _ := progressRepo(progressFakeConfig{counts: map[uint64]int{1: 2}})
	uc := usecase.NewListCoursesWithProgressUseCase(crepo, mrepo, prepo)

	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		ActorUserID: 5, ActorWorkspace: domain.WorkspaceRefOf("0198a000-0000-7000-8000-0000000000c1"), ActorRole: domain.RoleTrainee,
	})
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, uint64(1), out[0].ID)
	assert.Equal(t, 3, out[0].MaterialCount)
	assert.Equal(t, 2, out[0].CompletedCount)
	assert.Equal(t, 12, out[1].MaterialCount)
	assert.Equal(t, 0, out[1].CompletedCount, "完了記録が無いコースは 0")
}

func Test_コース一覧進捗付き_集計に無いコースは0章(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{rows: []domain.Course{{ID: 7, Title: "空のコース"}}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	prepo, _ := progressRepo(progressFakeConfig{})
	uc := usecase.NewListCoursesWithProgressUseCase(crepo, mrepo, prepo)

	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		ActorUserID: 5, ActorWorkspace: domain.WorkspaceRefOf("0198a000-0000-7000-8000-0000000000c1"), ActorRole: domain.RoleTrainee,
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, 0, out[0].MaterialCount)
	assert.Equal(t, 0, out[0].CompletedCount)
}

func Test_コース一覧進捗付き_会社未所属は空スライス(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{})
	mrepo, _ := materialRepo(materialFakeConfig{})
	prepo, _ := progressRepo(progressFakeConfig{})
	uc := usecase.NewListCoursesWithProgressUseCase(crepo, mrepo, prepo)
	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		ActorUserID: 5, ActorWorkspace: domain.NoWorkspace(), ActorRole: domain.RoleSuperAdmin,
	})
	require.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out)
}

func Test_コース一覧進捗付き_0件でもnilではなく空スライス(t *testing.T) {
	// GORM の Find は 0 件時に nil スライスを返し JSON で null になるため正規化する(FRESTYLE-70)。
	crepo, _ := courseRepo(courseFakeConfig{rows: nil})
	mrepo, _ := materialRepo(materialFakeConfig{})
	prepo, _ := progressRepo(progressFakeConfig{})
	uc := usecase.NewListCoursesWithProgressUseCase(crepo, mrepo, prepo)
	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		ActorUserID: 5, ActorWorkspace: domain.WorkspaceRefOf("0198a000-0000-7000-8000-0000000000c1"), ActorRole: domain.RoleTrainee,
	})
	require.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out)
}

func Test_コース一覧進捗付き_traineeは公開のみで集計(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{rows: []domain.Course{{ID: 1}}})
	mrepo, mstore := materialRepo(materialFakeConfig{counts: map[uint64]int{1: 5}})
	prepo, pstore := progressRepo(progressFakeConfig{counts: map[uint64]int{1: 1}})
	uc := usecase.NewListCoursesWithProgressUseCase(crepo, mrepo, prepo)

	_, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		ActorUserID: 5, ActorWorkspace: domain.WorkspaceRefOf("0198a000-0000-7000-8000-0000000000c1"), ActorRole: domain.RoleTrainee,
	})
	require.NoError(t, err)
	require.NotNil(t, mstore.lastCountIncludeUnpublished)
	assert.False(t, *mstore.lastCountIncludeUnpublished, "trainee の分母は published のみ")
	assert.True(t, pstore.countCalled, "trainee は完了章数も集計する")
}

func Test_コース一覧進捗付き_管理ロールは完了集計をスキップ(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{rows: []domain.Course{{ID: 1}}})
	mrepo, mstore := materialRepo(materialFakeConfig{counts: map[uint64]int{1: 5}})
	prepo, pstore := progressRepo(progressFakeConfig{counts: map[uint64]int{1: 3}})
	uc := usecase.NewListCoursesWithProgressUseCase(crepo, mrepo, prepo)

	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		ActorUserID: 5, ActorWorkspace: domain.WorkspaceRefOf("0198a000-0000-7000-8000-0000000000c1"), ActorRole: domain.RoleCompanyAdmin,
	})
	require.NoError(t, err)
	require.NotNil(t, mstore.lastCountIncludeUnpublished)
	assert.True(t, *mstore.lastCountIncludeUnpublished, "admin は下書き章も分母に含む")
	assert.False(t, pstore.countCalled, "管理ロールは完了記録を持たないため集計しない")
	assert.Equal(t, 0, out[0].CompletedCount)
}

func Test_コース一覧進捗付き_章数集計エラーはそのまま返す(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{rows: []domain.Course{{ID: 1}}})
	mrepo, _ := materialRepo(materialFakeConfig{countErr: context.DeadlineExceeded})
	prepo, _ := progressRepo(progressFakeConfig{})
	uc := usecase.NewListCoursesWithProgressUseCase(crepo, mrepo, prepo)

	_, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		ActorUserID: 5, ActorWorkspace: domain.WorkspaceRefOf("0198a000-0000-7000-8000-0000000000c1"), ActorRole: domain.RoleTrainee,
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func Test_コース一覧進捗付き_完了集計エラーはそのまま返す(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{rows: []domain.Course{{ID: 1}}})
	mrepo, _ := materialRepo(materialFakeConfig{counts: map[uint64]int{1: 5}})
	prepo, _ := progressRepo(progressFakeConfig{countErr: context.DeadlineExceeded})
	uc := usecase.NewListCoursesWithProgressUseCase(crepo, mrepo, prepo)

	_, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		ActorUserID: 5, ActorWorkspace: domain.WorkspaceRefOf("0198a000-0000-7000-8000-0000000000c1"), ActorRole: domain.RoleTrainee,
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
