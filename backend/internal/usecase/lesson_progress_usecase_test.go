package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// progressStore は進捗 mock が記録する状態(course クラスタのテストで共有)。
type progressStore struct {
	completed   map[uint64]uint64 // materialID -> courseID
	countCalled bool
}

// progressFakeConfig は進捗 mock の応答設定。ゼロ値はすべて「空を返す」。
type progressFakeConfig struct {
	listRows    []domain.UserLessonProgress
	completeErr error
	counts      map[uint64]int
	countErr    error
}

// progressRepo は LessonProgressRepository の mock に、このクラスタが使う応答を
// 設定して返す。
func progressRepo(cfg progressFakeConfig) (*mockProgressRepo, *progressStore) {
	st := &progressStore{completed: map[uint64]uint64{}}
	repo := &mockProgressRepo{}
	repo.On("MarkCompleted", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			if cfg.completeErr == nil {
				st.completed[args.Get(2).(uint64)] = args.Get(3).(uint64)
			}
		}).Return(true, cfg.completeErr).Maybe()
	repo.On("MarkIncomplete", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			delete(st.completed, args.Get(2).(uint64))
		}).Return(nil).Maybe()
	repo.On("ListByUser", mock.Anything, mock.Anything).Return(cfg.listRows, nil).Maybe()
	repo.On("CountCompletedByUserGroupedByCourse", mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { st.countCalled = true }).
		Return(cfg.counts, cfg.countErr).Maybe()
	return repo, st
}

// publishedSetup は「actor と同じワークスペース（wsA）・公開教材・公開コース」の正常に完了できる組み合わせを作る。
func publishedSetup(materialID, courseID uint64) (*mockMaterialRepo, *mockCourseRepo) {
	mat, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: materialID, CourseID: courseID, WorkspaceID: strPtr(wsA), IsPublished: true,
	}})
	crs, _ := courseRepo(courseFakeConfig{get: &domain.Course{
		ID: courseID, WorkspaceID: strPtr(wsA), IsPublished: true,
	}})
	return mat, crs
}

// nopActivityRepo は UserDailyActivityRepository の何もしない stub。
type nopActivityRepo struct{}

func (n *nopActivityRepo) Increment(_ context.Context, _ uint64, _ time.Time, _ repository.UserDailyActivityIncrement) error {
	return nil
}

func (n *nopActivityRepo) ListByUser(_ context.Context, _ uint64, _, _ time.Time) ([]domain.UserDailyActivity, error) {
	return nil, nil
}

func Test_レッスン完了_同じワークスペースの公開教材はcourse_idを解決して記録する(t *testing.T) {
	progress, pstore := progressRepo(progressFakeConfig{})
	mat, crs := publishedSetup(5, 99)
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(99), pstore.completed[5]) // 教材の course_id が使われる
}

func Test_レッスン完了_別ワークスペースの教材は403相当で弾く(t *testing.T) {
	progress, pstore := progressRepo(progressFakeConfig{})
	mat, crs := publishedSetup(5, 99) // wsA の教材
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsB), ActorRole: domain.RoleTrainee, TeachingMaterialID: 5, // 別 workspace
	})
	assert.ErrorIs(t, err, usecase.ErrLessonForbidden)
	assert.Empty(t, pstore.completed)
}

func Test_レッスン完了_trainee_に未公開の教材は403相当(t *testing.T) {
	progress, _ := progressRepo(progressFakeConfig{})
	mat, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 5, CourseID: 99, WorkspaceID: strPtr(wsA), IsPublished: false, // 下書き
	}})
	crs, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 99, WorkspaceID: strPtr(wsA), IsPublished: true}})
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	assert.ErrorIs(t, err, usecase.ErrLessonForbidden)
}

func Test_レッスン完了_存在しない教材は404相当(t *testing.T) {
	progress, _ := progressRepo(progressFakeConfig{})
	mat, _ := materialRepo(materialFakeConfig{getErr: domain.ErrNotFound})
	crs, _ := courseRepo(courseFakeConfig{})
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})
	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleTrainee, TeachingMaterialID: 404,
	})
	assert.ErrorIs(t, err, usecase.ErrLessonNotFound)
}

func Test_レッスン完了_記録失敗を伝播(t *testing.T) {
	wantErr := errors.New("db")
	progress, _ := progressRepo(progressFakeConfig{completeErr: wantErr})
	mat, crs := publishedSetup(5, 1)
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	assert.ErrorIs(t, err, wantErr, "repository のエラーを別のエラーに置き換えず伝播する")
}

func Test_レッスン完了取消_行を削除する(t *testing.T) {
	progress, pstore := progressRepo(progressFakeConfig{})
	pstore.completed[5] = 1
	uc := usecase.NewMarkLessonIncompleteUseCase(progress)
	require.NoError(t, uc.Execute(context.Background(), 1, 5))
	_, ok := pstore.completed[5]
	assert.False(t, ok)
}

func Test_学習進捗一覧_完了記録を返す(t *testing.T) {
	progress, _ := progressRepo(progressFakeConfig{
		listRows: []domain.UserLessonProgress{{TeachingMaterialID: 1, CourseID: 9}},
	})
	uc := usecase.NewListLessonProgressUseCase(progress)
	rows, err := uc.Execute(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, uint64(9), rows[0].CourseID)
}
