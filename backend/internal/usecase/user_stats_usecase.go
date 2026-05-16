package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetUserStatsUseCase は user の マイページ 集計 (合計 セッション 数 / 平均 スコア 等) を 返す。
// 依存 port: [repository.UserStatsRepository] (Raw SQL で 集計)。
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
