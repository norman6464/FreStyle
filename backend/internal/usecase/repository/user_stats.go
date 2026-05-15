package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// UserStatsRepository は user_stats を集計 して 返す。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// userStatsRepository (GORM Raw SQL)。
type UserStatsRepository interface {
	Compute(ctx context.Context, userID uint64) (*domain.UserStats, error)
}
