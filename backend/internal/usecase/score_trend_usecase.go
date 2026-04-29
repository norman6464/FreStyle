package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type GetScoreTrendUseCase struct {
	repo repository.ScoreTrendRepository
}

func NewGetScoreTrendUseCase(r repository.ScoreTrendRepository) *GetScoreTrendUseCase {
	return &GetScoreTrendUseCase{repo: r}
}

func (u *GetScoreTrendUseCase) Execute(ctx context.Context, userID uint64, days int) (*domain.ScoreTrend, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	points, err := u.repo.AggregateDaily(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	return &domain.ScoreTrend{UserID: userID, Points: points}, nil
}
