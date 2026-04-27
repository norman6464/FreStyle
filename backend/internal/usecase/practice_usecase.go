package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListPracticeScenariosUseCase struct {
	repo repository.PracticeScenarioRepository
}

func NewListPracticeScenariosUseCase(r repository.PracticeScenarioRepository) *ListPracticeScenariosUseCase {
	return &ListPracticeScenariosUseCase{repo: r}
}

func (u *ListPracticeScenariosUseCase) Execute(ctx context.Context) ([]domain.PracticeScenario, error) {
	return u.repo.ListActive(ctx)
}

type GetPracticeScenarioUseCase struct {
	repo repository.PracticeScenarioRepository
}

func NewGetPracticeScenarioUseCase(r repository.PracticeScenarioRepository) *GetPracticeScenarioUseCase {
	return &GetPracticeScenarioUseCase{repo: r}
}

func (u *GetPracticeScenarioUseCase) Execute(ctx context.Context, id uint64) (*domain.PracticeScenario, error) {
	if id == 0 {
		return nil, errors.New("id is required")
	}
	return u.repo.FindByID(ctx, id)
}
