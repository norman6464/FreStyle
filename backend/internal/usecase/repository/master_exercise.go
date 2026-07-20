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

// ListWithStatusInput は ListWithStatusByLanguage の入力パラメータ。
// Limit=0 はページネーション無効（全件）を表す。Offset は 0-based。
type ListWithStatusInput struct {
	UserID   uint64
	Language string
	Offset   int
	Limit    int
}

// ExerciseLanguageSummary は言語ごとの「公開済み問題数」と「current user が正解済みの問題数」。
// コード学習の言語選択カード（FRESTYLE-152）の進捗表示に使う read model。
type ExerciseLanguageSummary struct {
	Language string `json:"language"`
	Total    int64  `json:"total"`
	Solved   int64  `json:"solved"`
}

// MasterExerciseRepository は運営マスタ演習問題の永続化を担う（言語フィルタは ListByLanguage）。
type MasterExerciseRepository interface {
	ListByLanguage(ctx context.Context, language string) ([]domain.MasterExercise, error)
	GetByID(ctx context.Context, id uint64) (*domain.MasterExercise, error)
	GetBySlug(ctx context.Context, slug string) (*domain.MasterExercise, error)

	// SummaryByLanguage は公開済み問題を言語ごとに集計し、問題数と current user の正解済み件数を返す。
	// userID=0（未ログイン）は solved が全て 0 になる。言語の昇順で返す。
	SummaryByLanguage(ctx context.Context, userID uint64) ([]ExerciseLanguageSummary, error)

	// ListWithStatusByLanguage は公開済み問題を、 current user の提出状態 + 全体集計付きで
	// 1 クエリ（master_exercises ⟕ exercise_submissions 集計）で返す。userID=0 は未ログイン扱いで status="".
	// Limit > 0 のとき LIMIT/OFFSET を適用する。Limit=0 は全件取得。
	ListWithStatusByLanguage(ctx context.Context, in ListWithStatusInput) ([]MasterExerciseWithStatus, error)
}
