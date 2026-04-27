package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type UserStatsRepository interface {
	Compute(ctx context.Context, userID uint64) (*domain.UserStats, error)
}

type userStatsRepository struct{ db *gorm.DB }

func NewUserStatsRepository(db *gorm.DB) UserStatsRepository { return &userStatsRepository{db: db} }

// Compute は ai_chat_sessions と score_cards から集計。
// SQL は最小限に留め、追加のメトリクスは後続 PR で拡張する。
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
