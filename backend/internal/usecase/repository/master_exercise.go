package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// MasterExerciseRepository は運営マスタ演習問題の永続化を担う（言語フィルタは ListByLanguage）。
type MasterExerciseRepository interface {
	ListByLanguage(ctx context.Context, language string) ([]domain.MasterExercise, error)
	GetByID(ctx context.Context, id uint64) (*domain.MasterExercise, error)
	GetBySlug(ctx context.Context, slug string) (*domain.MasterExercise, error)
}
