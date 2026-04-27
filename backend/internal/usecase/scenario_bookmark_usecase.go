package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListScenarioBookmarksUseCase struct {
	repo repository.ScenarioBookmarkRepository
}

func NewListScenarioBookmarksUseCase(r repository.ScenarioBookmarkRepository) *ListScenarioBookmarksUseCase {
	return &ListScenarioBookmarksUseCase{repo: r}
}

func (u *ListScenarioBookmarksUseCase) Execute(ctx context.Context, userID uint64) ([]domain.ScenarioBookmark, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

type AddScenarioBookmarkUseCase struct {
	repo repository.ScenarioBookmarkRepository
}

func NewAddScenarioBookmarkUseCase(r repository.ScenarioBookmarkRepository) *AddScenarioBookmarkUseCase {
	return &AddScenarioBookmarkUseCase{repo: r}
}

func (u *AddScenarioBookmarkUseCase) Execute(ctx context.Context, userID, scenarioID uint64) (*domain.ScenarioBookmark, error) {
	if userID == 0 || scenarioID == 0 {
		return nil, errors.New("userID and scenarioID are required")
	}
	b := &domain.ScenarioBookmark{UserID: userID, ScenarioID: scenarioID}
	if err := u.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

type RemoveScenarioBookmarkUseCase struct {
	repo repository.ScenarioBookmarkRepository
}

func NewRemoveScenarioBookmarkUseCase(r repository.ScenarioBookmarkRepository) *RemoveScenarioBookmarkUseCase {
	return &RemoveScenarioBookmarkUseCase{repo: r}
}

func (u *RemoveScenarioBookmarkUseCase) Execute(ctx context.Context, userID, scenarioID uint64) error {
	if userID == 0 || scenarioID == 0 {
		return errors.New("userID and scenarioID are required")
	}
	return u.repo.Delete(ctx, userID, scenarioID)
}
