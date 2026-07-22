package usecase

import (
	"context"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetLastViewedChapterUseCase はコース内で actor 自身が最後に閲覧した章の閲覧記録を返す。
// コース詳細を開いたときの「続きから表示」(FRESTYLE-99)用。履歴が無い場合は nil を返す(エラーではない)。
// canReadCourse で他社・未公開(trainee 不可)コースへのアクセスを防ぐ。
type GetLastViewedChapterUseCase struct {
	courses      repository.CourseRepository
	chapterViews repository.UserChapterViewRepository
}

// NewGetLastViewedChapterUseCase は GetLastViewedChapterUseCase を組み立てる。
func NewGetLastViewedChapterUseCase(
	courses repository.CourseRepository,
	chapterViews repository.UserChapterViewRepository,
) *GetLastViewedChapterUseCase {
	return &GetLastViewedChapterUseCase{courses: courses, chapterViews: chapterViews}
}

// GetLastViewedChapterInput は取得対象コースと actor 情報(認証 context 由来)。
type GetLastViewedChapterInput struct {
	UserID         uint64
	ActorCompanyID uint64
	ActorRole      string
	CourseID       uint64
}

// Execute はコースの可視性を検証してから最終閲覧記録を返す。履歴なしは (nil, nil)。
func (u *GetLastViewedChapterUseCase) Execute(ctx context.Context, in GetLastViewedChapterInput) (*domain.UserChapterView, error) {
	course, err := u.courses.GetByID(ctx, in.CourseID)
	if err != nil {
		return nil, err
	}
	if !canReadCourse(course, in.ActorCompanyID, in.ActorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	return u.chapterViews.GetLastViewedByUserAndCourse(ctx, in.UserID, in.CourseID)
}
