package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

var (
	// ErrLessonNotFound は完了対象の教材が存在しないときに返す。
	ErrLessonNotFound = errors.New("lesson_not_found")
	// ErrLessonForbidden は他社の教材 / 閲覧不可な教材を完了にしようとしたときに返す。
	ErrLessonForbidden = errors.New("lesson_forbidden")
)

// MarkLessonCompletedUseCase は教材を完了として記録する（trainee 自身の進捗）。
// 教材から course_id を解決して記録するため、 クライアントが course を詐称できない。
// さらに actor の company / role で可視性を検証し、 他社・非公開教材の完了を弾く（IDOR 対策）。
// 完了後に user_daily_activities をベストエフォートでインクリメントする。
type MarkLessonCompletedUseCase struct {
	progress  repository.LessonProgressRepository
	materials repository.TeachingMaterialRepository
	courses   repository.CourseRepository
	activity  repository.UserDailyActivityRepository
}

func NewMarkLessonCompletedUseCase(
	p repository.LessonProgressRepository,
	m repository.TeachingMaterialRepository,
	c repository.CourseRepository,
	activity repository.UserDailyActivityRepository,
) *MarkLessonCompletedUseCase {
	return &MarkLessonCompletedUseCase{progress: p, materials: m, courses: c, activity: activity}
}

// MarkLessonCompletedInput は完了記録の入力。 actor の company / role で可視性を検証する。
type MarkLessonCompletedInput struct {
	UserID             uint64
	ActorCompanyID     uint64
	ActorRole          string
	TeachingMaterialID uint64
}

func (u *MarkLessonCompletedUseCase) Execute(ctx context.Context, in MarkLessonCompletedInput) error {
	m, err := u.materials.GetByID(ctx, in.TeachingMaterialID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLessonNotFound
		}
		return err
	}
	if m == nil {
		return ErrLessonNotFound
	}
	course, err := u.courses.GetByID(ctx, m.CourseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLessonNotFound
		}
		return err
	}
	// 自社かつ閲覧可能な教材のみ完了にできる（他社教材 / trainee に未公開の教材を弾く）。
	// 既存の単一教材取得 (TeachingMaterialUseCase.Get) と同じ canRead で判定する。
	if !canRead(m, course, in.ActorCompanyID, in.ActorRole) {
		return ErrLessonForbidden
	}
	changed, err := u.progress.MarkCompleted(ctx, in.UserID, in.TeachingMaterialID, m.CourseID)
	if err != nil {
		return err
	}
	// 初回完了時のみカウントを加算する（重複リクエストでは changed=false なのでスキップ）。
	if changed {
		if err := u.activity.Increment(ctx, in.UserID, time.Now().UTC(), repository.UserDailyActivityIncrement{
			LessonCount: 1,
		}); err != nil {
			slog.WarnContext(ctx, "user_daily_activities increment failed", "userID", in.UserID, "err", err)
		}
	}
	return nil
}

// MarkLessonIncompleteUseCase は完了記録を取り消す。
// 対象は (user, material) 一致の自分の行のみで、 他人の進捗には触れられない。
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
