package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// RecordChapterViewUseCase はユーザーが章（教材）を開いたときに閲覧記録を残す。
// 「続きから」カードや閲覧履歴の基盤となる。完了（MarkLessonCompleted）とは別に、
// ページを開いただけでも記録するため、離脱した章も追跡できる。
// canRead で actor の会社・ロールを検証し、他社教材への不正記録を防ぐ。
type RecordChapterViewUseCase struct {
	chapterViews repository.UserChapterViewRepository
	materials    repository.TeachingMaterialRepository
	courses      repository.CourseRepository
}

func NewRecordChapterViewUseCase(
	cv repository.UserChapterViewRepository,
	m repository.TeachingMaterialRepository,
	c repository.CourseRepository,
) *RecordChapterViewUseCase {
	return &RecordChapterViewUseCase{chapterViews: cv, materials: m, courses: c}
}

// RecordChapterViewInput は章閲覧記録の入力。actor の会社・ロールで可視性を検証する。
type RecordChapterViewInput struct {
	UserID             uint64
	ActorCompanyID     uint64
	ActorRole          string
	TeachingMaterialID uint64
}

// ErrChapterViewForbidden は閲覧権限のない教材を記録しようとしたときに返す。
var ErrChapterViewForbidden = errors.New("chapter_view_forbidden")

// Execute は course_id を教材から解決し、canRead を検証してから upsert する。
// エラー時は呼び出し元の処理を止めない想定（handler 側でベストエフォート扱い）。
func (u *RecordChapterViewUseCase) Execute(ctx context.Context, in RecordChapterViewInput) error {
	m, err := u.materials.GetByID(ctx, in.TeachingMaterialID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrLessonNotFound
	}
	course, err := u.courses.GetByID(ctx, m.CourseID)
	if err != nil {
		return err
	}
	if course == nil {
		return ErrLessonNotFound
	}
	if !canRead(m, course, in.ActorCompanyID, in.ActorRole) {
		return ErrChapterViewForbidden
	}
	return u.chapterViews.UpsertView(ctx, in.UserID, in.TeachingMaterialID, m.CourseID)
}
