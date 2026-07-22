package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetExerciseLanguageSummaryUseCase は公開済み演習を言語ごとに集計するユースケース。
// コード学習の言語選択カード（問題数 + 完了数）に使う（FRESTYLE-152）。
type GetExerciseLanguageSummaryUseCase struct {
	exercises repository.MasterExerciseRepository
}

// NewGetExerciseLanguageSummaryUseCase は GetExerciseLanguageSummaryUseCase を生成する。
func NewGetExerciseLanguageSummaryUseCase(exercises repository.MasterExerciseRepository) *GetExerciseLanguageSummaryUseCase {
	return &GetExerciseLanguageSummaryUseCase{exercises: exercises}
}

// Execute は言語別の集計を返す。userID=0（未ログイン）は solved が 0 になる。
func (u *GetExerciseLanguageSummaryUseCase) Execute(ctx context.Context, userID uint64) ([]repository.ExerciseLanguageSummary, error) {
	return u.exercises.SummaryByLanguage(ctx, userID)
}
