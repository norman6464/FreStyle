package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrLessonNotFound は完了対象の教材が存在しないときに返す。
var ErrLessonNotFound = errors.New("lesson_not_found")

// MarkLessonCompletedUseCase は教材を完了として記録する（trainee 自身の進捗）。
// 教材から course_id を解決して記録するため、クライアントが course を詐称できない。
type MarkLessonCompletedUseCase struct {
	progress  repository.LessonProgressRepository
	materials repository.TeachingMaterialRepository
}

func NewMarkLessonCompletedUseCase(
	p repository.LessonProgressRepository,
	m repository.TeachingMaterialRepository,
) *MarkLessonCompletedUseCase {
	return &MarkLessonCompletedUseCase{progress: p, materials: m}
}

func (u *MarkLessonCompletedUseCase) Execute(ctx context.Context, userID, materialID uint64) error {
	m, err := u.materials.GetByID(ctx, materialID)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrLessonNotFound
	}
	return u.progress.MarkCompleted(ctx, userID, materialID, m.CourseID)
}

// MarkLessonIncompleteUseCase は完了記録を取り消す。
type MarkLessonIncompleteUseCase struct {
	progress repository.LessonProgressRepository
}

func NewMarkLessonIncompleteUseCase(p repository.LessonProgressRepository) *MarkLessonIncompleteUseCase {
	return &MarkLessonIncompleteUseCase{progress: p}
}

func (u *MarkLessonIncompleteUseCase) Execute(ctx context.Context, userID, materialID uint64) error {
	return u.progress.MarkIncomplete(ctx, userID, materialID)
}

// ListLessonProgressUseCase は user の完了記録一覧を返す（進捗バー / 完了チェック表示用）。
type ListLessonProgressUseCase struct {
	progress repository.LessonProgressRepository
}

func NewListLessonProgressUseCase(p repository.LessonProgressRepository) *ListLessonProgressUseCase {
	return &ListLessonProgressUseCase{progress: p}
}

func (u *ListLessonProgressUseCase) Execute(ctx context.Context, userID uint64) ([]domain.UserLessonProgress, error) {
	return u.progress.ListByUser(ctx, userID)
}
