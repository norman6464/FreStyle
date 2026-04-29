package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type GetScoreGoalUseCase struct {
	repo repository.ScoreGoalRepository
}

func NewGetScoreGoalUseCase(r repository.ScoreGoalRepository) *GetScoreGoalUseCase {
	return &GetScoreGoalUseCase{repo: r}
}

func (u *GetScoreGoalUseCase) Execute(ctx context.Context, userID uint64) (*domain.ScoreGoal, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.FindByUserID(ctx, userID)
}

type UpsertScoreGoalUseCase struct {
	repo repository.ScoreGoalRepository
}

func NewUpsertScoreGoalUseCase(r repository.ScoreGoalRepository) *UpsertScoreGoalUseCase {
	return &UpsertScoreGoalUseCase{repo: r}
}

func (u *UpsertScoreGoalUseCase) Execute(ctx context.Context, userID uint64, target float64) (*domain.ScoreGoal, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	if target < 0 || target > 10 {
		return nil, errors.New("target must be between 0 and 10")
	}
	g := &domain.ScoreGoal{UserID: userID, TargetScore: target}
	if err := u.repo.Upsert(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}
