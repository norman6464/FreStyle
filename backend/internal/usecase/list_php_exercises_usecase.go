package usecase

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// ListPhpExercisesUseCase は PHP 演習問題一覧を返す。
type ListPhpExercisesUseCase struct {
	repo repository.PhpExerciseRepository
}

func NewListPhpExercisesUseCase(repo repository.PhpExerciseRepository) *ListPhpExercisesUseCase {
	return &ListPhpExercisesUseCase{repo: repo}
}

func (uc *ListPhpExercisesUseCase) Execute() ([]domain.PhpExercise, error) {
	return uc.repo.List()
}
