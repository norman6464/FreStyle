package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type GetUserStatsUseCase struct {
	stats repository.UserStatsRepository
}

func NewGetUserStatsUseCase(s repository.UserStatsRepository) *GetUserStatsUseCase {
	return &GetUserStatsUseCase{stats: s}
}

func (u *GetUserStatsUseCase) Execute(ctx context.Context, userID uint64) (*domain.UserStats, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.stats.Compute(ctx, userID)
}
