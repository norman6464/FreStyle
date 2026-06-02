package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListMasterExercisesUseCase は指定言語の運営マスタ演習問題一覧を返す。
type ListMasterExercisesUseCase struct {
	repo repository.MasterExerciseRepository
}

func NewListMasterExercisesUseCase(repo repository.MasterExerciseRepository) *ListMasterExercisesUseCase {
	return &ListMasterExercisesUseCase{repo: repo}
}

// Execute は language 指定があれば該当言語のみ、空文字なら全言語の問題を返す。
func (uc *ListMasterExercisesUseCase) Execute(ctx context.Context, language string) ([]domain.MasterExercise, error) {
	return uc.repo.ListByLanguage(ctx, language)
}
