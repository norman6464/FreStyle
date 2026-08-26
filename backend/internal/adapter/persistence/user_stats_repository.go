package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// userStatsRepository は [repository.UserStatsRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type userStatsRepository struct{ db *sql.DB }

func NewUserStatsRepository(db *sql.DB) repository.UserStatsRepository {
	return &userStatsRepository{db: db}
}

// Compute は score_cards から提出数 / 平均スコアを集計する。
func (r *userStatsRepository) Compute(ctx context.Context, userID uint64) (*domain.UserStats, error) {
	stats := &domain.UserStats{UserID: userID}
	uid, ok := toInt64ID(userID)
	if !ok {
		return stats, nil // 存在し得ない user_id = 0 件（count=0 / avg=0）
	}
	row, err := sqlcgen.New(r.db).ComputeUserStats(ctx, uid)
	if err != nil {
		return nil, err
	}
	stats.TotalSessions = int(row.TotalSessions)
	stats.AverageScore = row.AverageScore
	return stats, nil
}
