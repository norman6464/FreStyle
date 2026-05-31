package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// UserStatsRepository は user_stats を集計して返す。
type UserStatsRepository interface {
	Compute(ctx context.Context, userID uint64) (*domain.UserStats, error)
}
