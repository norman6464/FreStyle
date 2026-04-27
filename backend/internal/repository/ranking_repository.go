package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type RankingRepository interface {
	TopByAverageScore(ctx context.Context, limit int) ([]domain.RankingEntry, error)
}

type rankingRepository struct{ db *gorm.DB }

func NewRankingRepository(db *gorm.DB) RankingRepository { return &rankingRepository{db: db} }

func (r *rankingRepository) TopByAverageScore(ctx context.Context, limit int) ([]domain.RankingEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT u.id, u.display_name, COALESCE(AVG(sc.overall_score), 0) AS avg_score
		FROM users u
		LEFT JOIN score_cards sc ON sc.user_id = u.id
		WHERE u.deleted_at IS NULL
		GROUP BY u.id, u.display_name
		ORDER BY avg_score DESC
		LIMIT ?
	`, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.RankingEntry
	rank := 1
	for rows.Next() {
		var e domain.RankingEntry
		if err := rows.Scan(&e.UserID, &e.DisplayName, &e.AverageScore); err != nil {
			return nil, err
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}
	return entries, nil
}
