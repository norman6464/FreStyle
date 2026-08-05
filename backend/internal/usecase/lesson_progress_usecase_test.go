package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository/repofakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// progressStore は進捗 fake が記録する状態(course クラスタのテストで共有)。
type progressStore struct {
	completed   map[uint64]uint64 // materialID -> courseID
	countCalled bool
}

// progressFakeConfig は進捗 fake の応答設定。ゼロ値はすべて「空を返す」。
type progressFakeConfig struct {
	listRows    []domain.UserLessonProgress
	completeErr error
	counts      map[uint64]int
	countErr    error
}

// progressRepo は LessonProgressRepository の生成 fake に、このクラスタが使う
// メソッドだけを差し込んで返す。
func progressRepo(cfg progressFakeConfig) (*repofakes.FakeLessonProgressRepository, *progressStore) {
	st := &progressStore{completed: map[uint64]uint64{}}
	repo := &repofakes.FakeLessonProgressRepository{
		MarkCompletedFunc: func(_ context.Context, _, materialID, courseID uint64) (bool, error) {
			if cfg.completeErr != nil {
				return false, cfg.completeErr
			}
			_, alreadyDone := st.completed[materialID]
			st.completed[materialID] = courseID
			return !alreadyDone, nil
		},
		MarkIncompleteFunc: func(_ context.Context, _, materialID uint64) error {
			delete(st.completed, materialID)
			return nil
		},
		ListByUserFunc: func(context.Context, uint64) ([]domain.UserLessonProgress, error) {
			return cfg.listRows, nil
		},
		CountCompletedByUserGroupedByCourseFunc: func(context.Context, uint64) (map[uint64]int, error) {
			st.countCalled = true
			if cfg.countErr != nil {
				return nil, cfg.countErr
			}
			return cfg.counts, nil
		},
	}
	return repo, st
}

// publishedSetup は「自社・公開教材・公開コース」の正常に完了できる組み合わせを作る。
func publishedSetup(materialID, companyID, courseID uint64) (*repofakes.FakeTeachingMaterialRepository, *repofakes.FakeCourseRepository) {
	mat, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: materialID, CompanyID: companyID, CourseID: courseID, IsPublished: true,
	}})
	crs, _ := courseRepo(courseFakeConfig{get: &domain.Course{
		ID: courseID, CompanyID: companyID, IsPublished: true,
	}})
	return mat, crs
}

func Test_レッスン完了_自社の公開教材はcourse_idを解決して記録する(t *testing.T) {
	progress, pstore := progressRepo(progressFakeConfig{})
	mat, crs := publishedSetup(5, 10, 99)
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(99), pstore.completed[5]) // 教材の course_id が使われる
}

func Test_レッスン完了_他社の教材は403相当で弾く(t *testing.T) {
	progress, pstore := progressRepo(progressFakeConfig{})
	mat, crs := publishedSetup(5, 10, 99) // company 10 の教材
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 20, ActorRole: domain.RoleTrainee, TeachingMaterialID: 5, // 別 company
	})
	assert.ErrorIs(t, err, usecase.ErrLessonForbidden)
	assert.Empty(t, pstore.completed)
}

func Test_レッスン完了_trainee_に未公開の教材は403相当(t *testing.T) {
	progress, _ := progressRepo(progressFakeConfig{})
	mat, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 5, CompanyID: 10, CourseID: 99, IsPublished: false, // 下書き
	}})
	crs, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 99, CompanyID: 10, IsPublished: true}})
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
	})
	assert.ErrorIs(t, err, usecase.ErrLessonForbidden)
}

func Test_レッスン完了_存在しない教材は404相当(t *testing.T) {
	progress, _ := progressRepo(progressFakeConfig{})
	mat, _ := materialRepo(materialFakeConfig{getErr: gorm.ErrRecordNotFound})
	crs, _ := courseRepo(courseFakeConfig{})
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})
	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, TeachingMaterialID: 404,
	})
	assert.ErrorIs(t, err, usecase.ErrLessonNotFound)
}

func Test_レッスン完了_記録失敗を伝播(t *testing.T) {
	wantErr := errors.New("db")
	progress, _ := progressRepo(progressFakeConfig{completeErr: wantErr})
	mat, crs := publishedSetup(5, 10, 1)
	uc := usecase.NewMarkLessonCompletedUseCase(progress, mat, crs, &nopActivityRepo{})

	err := uc.Execute(context.Background(), usecase.MarkLessonCompletedInput{
		UserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, TeachingMaterialID: 5,
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
