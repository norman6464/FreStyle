package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// RecordChapterViewUseCase はユーザーが章（教材）を開いたときに閲覧記録を残す。
// 「続きから」カードや閲覧履歴の基盤となる。完了（MarkLessonCompleted）とは別に、
// ページを開いただけでも記録するため、離脱した章も追跡できる。
type RecordChapterViewUseCase struct {
	chapterViews  repository.UserChapterViewRepository
	materials     repository.TeachingMaterialRepository
}

func NewRecordChapterViewUseCase(
	cv repository.UserChapterViewRepository,
	m repository.TeachingMaterialRepository,
) *RecordChapterViewUseCase {
	return &RecordChapterViewUseCase{chapterViews: cv, materials: m}
}

// RecordChapterViewInput は章閲覧記録の入力。
type RecordChapterViewInput struct {
	UserID             uint64
	TeachingMaterialID uint64
}

// Execute は course_id を教材から解決し upsert する。
// エラー時は呼び出し元の処理を止めない想定（handler 側でベストエフォート扱い）。
func (u *RecordChapterViewUseCase) Execute(ctx context.Context, in RecordChapterViewInput) error {
	m, err := u.materials.GetByID(ctx, in.TeachingMaterialID)
	if err != nil || m == nil {
		return err
	}
	return u.chapterViews.UpsertView(ctx, in.UserID, in.TeachingMaterialID, m.CourseID)
}
