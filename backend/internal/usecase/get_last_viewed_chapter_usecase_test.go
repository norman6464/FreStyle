package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// chapterViewRepo は UserChapterViewRepository の mock に、このテストが使う
// GetLastViewedByUserAndCourse の応答だけを設定して返す。
func chapterViewRepo(lastViewed *domain.UserChapterView, getErr error) *mockChapterViewRepo {
	repo := &mockChapterViewRepo{}
	repo.On("GetLastViewedByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).
		Return(lastViewed, getErr).Maybe()
	return repo
}

func newLastViewedUC(cfg materialFactsConfig, view *domain.UserChapterView) *usecase.GetLastViewedChapterUseCase {
	_, perm := materialPerm(cfg)
	return usecase.NewGetLastViewedChapterUseCase(chapterViewRepo(view, nil), perm)
}

func Test_最終閲覧章_公開コースなら履歴を返す(t *testing.T) {
	view := &domain.UserChapterView{UserID: 1, TeachingMaterialID: 42, CourseID: 5, LastViewedAt: time.Now()}
	uc := newLastViewedUC(materialFactsConfig{member: true, published: true}, view)

	got, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		MaterialActor: actorIn(wsA), CourseID: 5,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, uint64(42), got.TeachingMaterialID)
}

func Test_最終閲覧章_履歴なしはnilを返す(t *testing.T) {
	uc := newLastViewedUC(materialFactsConfig{member: true, published: true}, nil)

	got, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		MaterialActor: actorIn(wsA), CourseID: 5,
	})
	require.NoError(t, err)
	assert.Nil(t, got, "初めて開くコースは履歴なし = 正常系")
}

func Test_最終閲覧章_読めないコースは実在を教えない(t *testing.T) {
	// 履歴の有無からコースの実在が読めてはいけない。どの理由でも同じ ErrNotFound。
	for _, c := range []struct {
		name string
		cfg  materialFactsConfig
	}{
		{"別テナントのコース", materialFactsConfig{notFound: true}},
		{"付与の無い下書き", materialFactsConfig{member: true, published: false}},
		{"所属していない", materialFactsConfig{member: false, published: true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			uc := newLastViewedUC(c.cfg, nil)
			_, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
				MaterialActor: actorIn(wsA), CourseID: 5,
			})
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func Test_最終閲覧章_未所属は読めない(t *testing.T) {
	uc := newLastViewedUC(materialFactsConfig{member: true, published: true}, nil)
	_, err := uc.Execute(context.Background(), usecase.GetLastViewedChapterInput{
		MaterialActor: usecase.MaterialActor{ActorUserID: 1}, CourseID: 5,
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
