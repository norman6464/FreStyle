package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// userStatsRepository は [repository.UserStatsRepository] の GORM 実装。
type userStatsRepository struct{ db *gorm.DB }

func NewUserStatsRepository(db *gorm.DB) repository.UserStatsRepository {
	return &userStatsRepository{db: db}
}

// Compute は score_cards から提出数 / 平均スコアを集計する。
func (r *userStatsRepository) Compute(ctx context.Context, userID uint64) (*domain.UserStats, error) {
	stats := &domain.UserStats{UserID: userID}
	row := r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*), COALESCE(AVG(overall_score),0) FROM score_cards WHERE user_id = ?",
		userID,
	).Row()
	var avg float64
	if err := row.Scan(&stats.TotalSessions, &avg); err != nil {
		return nil, err
	}
	stats.AverageScore = avg
	return stats, nil
}
