package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

func (f *fakeLessonProgressRepo) MarkCompleted(_ context.Context, _, materialID, courseID uint64) (bool, error) {
	if f.completeErr != nil {
		return false, f.completeErr
	}
	_, alreadyDone := f.completed[materialID]
	f.completed[materialID] = courseID
	return !alreadyDone, nil
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
	getErr   error
}

func (f *fakeMaterialRepoForProgress) GetByID(context.Context, uint64) (*domain.TeachingMaterial, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.material, nil
}

func (f *fakeMaterialRepoForProgress) ListByCompany(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (f *fakeMaterialRepoForProgress) ListByCourse(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (f *fakeMaterialRepoForProgress) CountByCourseForCompany(context.Context, uint64, bool) (map[uint64]int, error) {
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

// fakeCourseRepoForProgress は CourseRepository の最小 fake（GetByID のみ意味を持つ）。
type fakeCourseRepoForProgress struct {
	course *domain.Course
	getErr error
}

func (f *fakeCourseRepoForProgress) GetByID(context.Context, uint64) (*domain.Course, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.course, nil
}

func (f *fakeCourseRepoForProgress) ListByCompany(context.Context, uint64, bool) ([]domain.Course, error) {
	return nil, nil
}
func (f *fakeCourseRepoForProgress) Create(context.Context, *domain.Course) error { return nil }
func (f *fakeCourseRepoForProgress) Update(context.Context, *domain.Course) error { return nil }
func (f *fakeCourseRepoForProgress) Delete(context.Context, uint64) error         { return nil }

// publishedSetup は「自社・公開教材・公開コース」の正常に完了できる組み合わせを作る。
func publishedSetup(materialID, companyID, courseID uint64) (*fakeMaterialRepoForProgress, *fakeCourseRepoForProgress) {
	mat := &fakeMaterialRepoForProgress{material: &domain.TeachingMaterial{
		ID: materialID, CompanyID: companyID, CourseID: courseID, IsPublished: true,
	}}
	crs := &fakeCourseRepoForProgress{course: &domain.Course{
		ID: courseID, CompanyID: companyID, IsPublished: true,
	}}
	return mat, crs
}

func Test_レッスン完了_自社の公開教材はcourse_idを解決して記録する(t *testing.T) {
	progress := newFakeLessonProgressRepo()
	mat, crs := publishedSetup(5, 10, 99)
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(99), progress.completed[5]) // 教材の course_id が使われる
}

func Test_レッスン完了_他社の教材は403相当で弾く(t *testing.T) {
	progress := newFakeLessonProgressRepo()
	mat, crs := publishedSetup(5, 10, 99) // company 10 の教材
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 20, ActorRole: domain.RoleTrainee, TeachingMaterialID: 5, // 別 company
	})
	assert.ErrorIs(t, err, usecase.ErrLessonForbidden)
	assert.Empty(t, progress.completed)
}

func Test_レッスン完了_trainee_に未公開の教材は403相当(t *testing.T) {
	progress := newFakeLessonProgressRepo()
	mat := &fakeMaterialRepoForProgress{material: &domain.TeachingMaterial{
		ID: 5, CompanyID: 10, CourseID: 99, IsPublished: false, // 下書き
	}}
	crs := &fakeCourseRepoForProgress{course: &domain.Course{ID: 99, CompanyID: 10, IsPublished: true}}
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	assert.ErrorIs(t, err, usecase.ErrLessonForbidden)
}

func Test_レッスン完了_存在しない教材は404相当(t *testing.T) {
	uc := usecase.NewMarkLessonCompletedUseCase(
		newFakeLessonProgressRepo(),
		&fakeMaterialRepoForProgress{getErr: gorm.ErrRecordNotFound},
		&fakeCourseRepoForProgress{},
		&nopActivityRepo{},
	)
	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, TeachingMaterialID: 404,
	})
	assert.ErrorIs(t, err, usecase.ErrLessonNotFound)
}

func Test_レッスン完了_記録失敗を伝播(t *testing.T) {
	progress := newFakeLessonProgressRepo()
	progress.completeErr = errors.New("db")
	mat, crs := publishedSetup(5, 10, 1)
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	assert.Error(t, err)
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
