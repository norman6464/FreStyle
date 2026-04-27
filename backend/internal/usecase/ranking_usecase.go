package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type GetRankingUseCase struct{ repo repository.RankingRepository }

func NewGetRankingUseCase(r repository.RankingRepository) *GetRankingUseCase {
	return &GetRankingUseCase{repo: r}
}

func (u *GetRankingUseCase) Execute(ctx context.Context, limit int) ([]domain.RankingEntry, error) {
	return u.repo.TopByAverageScore(ctx, limit)
}
