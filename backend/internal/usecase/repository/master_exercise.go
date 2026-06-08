package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// MasterExerciseWithStatus は問題 + current user の状態（solved / in_progress / 未提出""）+ 全体集計のセット。
// 一覧ページ用の read model で、 ListWithStatusByLanguage が 1 クエリでまとめて返す。
type MasterExerciseWithStatus struct {
	domain.MasterExercise

	Status string                  `json:"status"`
	Stats  ExerciseSubmissionStats `json:"stats"`
}

// MasterExerciseRepository は運営マスタ演習問題の永続化を担う（言語フィルタは ListByLanguage）。
type MasterExerciseRepository interface {
	ListByLanguage(ctx context.Context, language string) ([]domain.MasterExercise, error)
	GetByID(ctx context.Context, id uint64) (*domain.MasterExercise, error)
	GetBySlug(ctx context.Context, slug string) (*domain.MasterExercise, error)

	// ListWithStatusByLanguage は公開済み問題を、 current user の提出状態 + 全体集計付きで
	// 1 クエリ（master_exercises ⟕ exercise_submissions 集計）で返す。userID=0 は未ログイン扱いで status="".
	ListWithStatusByLanguage(ctx context.Context, userID uint64, language string) ([]MasterExerciseWithStatus, error)
}
