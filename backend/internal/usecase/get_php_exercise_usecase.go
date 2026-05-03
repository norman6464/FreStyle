package usecase

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// GetPhpExerciseUseCase は指定 ID の PHP 演習問題を返す。
type GetPhpExerciseUseCase struct {
	repo repository.PhpExerciseRepository
}

func NewGetPhpExerciseUseCase(repo repository.PhpExerciseRepository) *GetPhpExerciseUseCase {
	return &GetPhpExerciseUseCase{repo: repo}
}

func (uc *GetPhpExerciseUseCase) Execute(id uint) (*domain.PhpExercise, error) {
	return uc.repo.GetByID(id)
}
