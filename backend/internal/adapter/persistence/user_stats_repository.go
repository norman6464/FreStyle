package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// userStatsRepository は [repository.UserStatsRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type userStatsRepository struct{ db *gorm.DB }

func NewUserStatsRepository(db *gorm.DB) repository.UserStatsRepository {
	return &userStatsRepository{db: db}
}

// Compute は score_cards から提出数 / 平均スコアを集計する。
func (r *userStatsRepository) Compute(ctx context.Context, userID uint64) (*domain.UserStats, error) {
	stats := &domain.UserStats{UserID: userID}
	uid, ok := toInt64ID(userID)
	if !ok {
		return stats, nil // 存在し得ない user_id = 0 件（count=0 / avg=0）
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).ComputeUserStats(ctx, uid)
	if err != nil {
		return nil, err
	}
	stats.TotalSessions = int(row.TotalSessions)
	stats.AverageScore = row.AverageScore
	return stats, nil
}
