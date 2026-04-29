package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListAdminScenariosUseCase struct {
	repo repository.AdminScenarioRepository
}

func NewListAdminScenariosUseCase(r repository.AdminScenarioRepository) *ListAdminScenariosUseCase {
	return &ListAdminScenariosUseCase{repo: r}
}

func (u *ListAdminScenariosUseCase) Execute(ctx context.Context) ([]domain.PracticeScenario, error) {
	return u.repo.List(ctx)
}

type CreateAdminScenarioUseCase struct {
	repo repository.AdminScenarioRepository
}

func NewCreateAdminScenarioUseCase(r repository.AdminScenarioRepository) *CreateAdminScenarioUseCase {
	return &CreateAdminScenarioUseCase{repo: r}
}

func (u *CreateAdminScenarioUseCase) Execute(ctx context.Context, s *domain.PracticeScenario) (*domain.PracticeScenario, error) {
	if s == nil || s.Title == "" {
		return nil, errors.New("title is required")
	}
	if err := u.repo.Create(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

type UpdateAdminScenarioUseCase struct {
	repo repository.AdminScenarioRepository
}

func NewUpdateAdminScenarioUseCase(r repository.AdminScenarioRepository) *UpdateAdminScenarioUseCase {
	return &UpdateAdminScenarioUseCase{repo: r}
}

func (u *UpdateAdminScenarioUseCase) Execute(ctx context.Context, s *domain.PracticeScenario) (*domain.PracticeScenario, error) {
	if s == nil || s.ID == 0 {
		return nil, errors.New("id is required")
	}
	if err := u.repo.Update(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

type DeleteAdminScenarioUseCase struct {
	repo repository.AdminScenarioRepository
}

func NewDeleteAdminScenarioUseCase(r repository.AdminScenarioRepository) *DeleteAdminScenarioUseCase {
	return &DeleteAdminScenarioUseCase{repo: r}
}

func (u *DeleteAdminScenarioUseCase) Execute(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("id is required")
	}
	return u.repo.Delete(ctx, id)
}
