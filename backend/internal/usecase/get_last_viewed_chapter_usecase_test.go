package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository/repofakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// chapterViewRepo は UserChapterViewRepository の生成 fake に、このテストが使う
// GetLastViewedByUserAndCourse だけを差し込んで返す。
func chapterViewRepo(lastViewed *domain.UserChapterView, getErr error) *repofakes.FakeUserChapterViewRepository {
	return &repofakes.FakeUserChapterViewRepository{
		GetLastViewedByUserAndCourseFunc: func(context.Context, uint64, uint64) (*domain.UserChapterView, error) {
			if getErr != nil {
				return nil, getErr
			}
			return lastViewed, nil
		},
	}
}

func Test_最終閲覧章_履歴があれば返す(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}})
	view := &domain.UserChapterView{UserID: 7, TeachingMaterialID: 42, CourseID: 5, LastViewedAt: time.Now()}
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, chapterViewRepo(view, nil))

	got, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, uint64(42), got.TeachingMaterialID)
}

func Test_最終閲覧章_履歴なしはnilを返す(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}})
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, chapterViewRepo(nil, nil))

	got, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.NoError(t, err)
	assert.Nil(t, got, "初めて開くコースは履歴なし = 正常系")
}

func Test_最終閲覧章_他社コースは禁止(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}})
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, chapterViewRepo(nil, nil))

	_, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 99, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_最終閲覧章_traineeは未公開コース禁止(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}})
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, chapterViewRepo(nil, nil))

	_, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_最終閲覧章_コースが無ければNotFound(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{getErr: gorm.ErrRecordNotFound})
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, chapterViewRepo(nil, nil))

	_, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
