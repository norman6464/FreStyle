package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// MasterExerciseWithStatus は問題 + current user の状態（solved / in_progress / 未提出""）+ 全体集計のセット。
type MasterExerciseWithStatus struct {
	domain.MasterExercise
	Status string                  `json:"status"`
	Stats  ExerciseSubmissionStats `json:"stats"`
}

// ListWithStatusInput は ListWithStatusByLanguage の入力パラメータ。
type ListWithStatusInput struct {
	UserID   uint64
	Language string
	Offset   int
	Limit    int
}

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
	SummaryByLanguage(ctx context.Context, userID uint64) ([]ExerciseLanguageSummary, error)
	ListWithStatusByLanguage(ctx context.Context, in ListWithStatusInput) ([]MasterExerciseWithStatus, error)
}
