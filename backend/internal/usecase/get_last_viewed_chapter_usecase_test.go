package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// fakeChapterViewRepo は UserChapterViewRepository の fake。
type fakeChapterViewRepo struct {
	lastViewed *domain.UserChapterView
	getErr     error
}

func (f *fakeChapterViewRepo) UpsertView(context.Context, uint64, uint64, uint64) error { return nil }

func (f *fakeChapterViewRepo) ListRecentByUser(context.Context, uint64, int) ([]domain.UserChapterView, error) {
	return nil, nil
}

func (f *fakeChapterViewRepo) GetLastViewedByUserAndCourse(context.Context, uint64, uint64) (*domain.UserChapterView, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.lastViewed, nil
}

func Test_最終閲覧章_履歴があれば返す(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	view := &domain.UserChapterView{UserID: 7, TeachingMaterialID: 42, CourseID: 5, LastViewedAt: time.Now()}
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, &fakeChapterViewRepo{lastViewed: view})

	got, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, uint64(42), got.TeachingMaterialID)
}

func Test_最終閲覧章_履歴なしはnilを返す(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, &fakeChapterViewRepo{lastViewed: nil})

	got, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.NoError(t, err)
	assert.Nil(t, got, "初めて開くコースは履歴なし = 正常系")
}

func Test_最終閲覧章_他社コースは禁止(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, &fakeChapterViewRepo{})

	_, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 99, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_最終閲覧章_traineeは未公開コース禁止(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}}
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, &fakeChapterViewRepo{})

	_, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_最終閲覧章_コースが無ければNotFound(t *testing.T) {
	crepo := &fakeCourseRepo{getErr: gorm.ErrRecordNotFound}
	uc := usecase.NewGetLastViewedChapterUseCase(crepo, &fakeChapterViewRepo{})

	_, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		UserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleTrainee, CourseID: 5,
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
