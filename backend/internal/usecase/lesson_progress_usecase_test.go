package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeLessonProgressRepo は LessonProgressRepository の fake。
type fakeLessonProgressRepo struct {
	completed   map[uint64]uint64 // materialID -> courseID
	listRows    []domain.UserLessonProgress
	completeErr error
}

func newFakeLessonProgressRepo() *fakeLessonProgressRepo {
	return &fakeLessonProgressRepo{completed: map[uint64]uint64{}}
}

func (f *fakeLessonProgressRepo) MarkCompleted(_ context.Context, _, materialID, courseID uint64) error {
	if f.completeErr != nil {
		return f.completeErr
	}
	f.completed[materialID] = courseID
	return nil
}

func (f *fakeLessonProgressRepo) MarkIncomplete(_ context.Context, _, materialID uint64) error {
	delete(f.completed, materialID)
	return nil
}

func (f *fakeLessonProgressRepo) ListByUser(context.Context, uint64) ([]domain.UserLessonProgress, error) {
	return f.listRows, nil
}

// fakeMaterialRepoForProgress は TeachingMaterialRepository の最小 fake（GetByID のみ意味を持つ）。
type fakeMaterialRepoForProgress struct {
	material *domain.TeachingMaterial
}

func (f *fakeMaterialRepoForProgress) GetByID(context.Context, uint64) (*domain.TeachingMaterial, error) {
	return f.material, nil
}

func (f *fakeMaterialRepoForProgress) ListByCompany(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (f *fakeMaterialRepoForProgress) ListByCourse(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (f *fakeMaterialRepoForProgress) Create(context.Context, *domain.TeachingMaterial) error {
	return nil
}

func (f *fakeMaterialRepoForProgress) Update(context.Context, *domain.TeachingMaterial) error {
	return nil
}

func (f *fakeMaterialRepoForProgress) Delete(context.Context, uint64) error { return nil }

func (f *fakeMaterialRepoForProgress) DeleteByCourse(context.Context, uint64) error { return nil }

func Test_レッスン完了_教材からcourse_idを解決して記録する(t *testing.T) {
	progress := newFakeLessonProgressRepo()
	materials := &fakeMaterialRepoForProgress{material: &domain.TeachingMaterial{ID: 5, CourseID: 99}}
	uc := usecase.NewMarkLessonCompletedUseCase(progress, materials)

	require.NoError(t, uc.Execute(context.Background(), 1, 5))
	assert.Equal(t, uint64(99), progress.completed[5]) // 教材の course_id が使われる
}

func Test_レッスン完了_存在しない教材は404相当(t *testing.T) {
	uc := usecase.NewMarkLessonCompletedUseCase(
		newFakeLessonProgressRepo(),
		&fakeMaterialRepoForProgress{material: nil},
	)
	err := uc.Execute(context.Background(), 1, 404)
	assert.ErrorIs(t, err, usecase.ErrLessonNotFound)
}

func Test_レッスン完了_記録失敗を伝播(t *testing.T) {
	progress := newFakeLessonProgressRepo()
	progress.completeErr = errors.New("db")
	uc := usecase.NewMarkLessonCompletedUseCase(
		progress,
		&fakeMaterialRepoForProgress{material: &domain.TeachingMaterial{ID: 5, CourseID: 1}},
	)
	assert.Error(t, uc.Execute(context.Background(), 1, 5))
}

func Test_レッスン完了取消_行を削除する(t *testing.T) {
	progress := newFakeLessonProgressRepo()
	progress.completed[5] = 1
	uc := usecase.NewMarkLessonIncompleteUseCase(progress)
	require.NoError(t, uc.Execute(context.Background(), 1, 5))
	_, ok := progress.completed[5]
	assert.False(t, ok)
}

func Test_学習進捗一覧_完了記録を返す(t *testing.T) {
	progress := newFakeLessonProgressRepo()
	progress.listRows = []domain.UserLessonProgress{{TeachingMaterialID: 1, CourseID: 9}}
	uc := usecase.NewListLessonProgressUseCase(progress)
	rows, err := uc.Execute(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, uint64(9), rows[0].CourseID)
}
