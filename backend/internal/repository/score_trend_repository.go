package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type ScoreTrendRepository interface {
	AggregateDaily(ctx context.Context, userID uint64, days int) ([]domain.ScoreTrendPoint, error)
}

type scoreTrendRepository struct{ db *gorm.DB }

func NewScoreTrendRepository(db *gorm.DB) ScoreTrendRepository { return &scoreTrendRepository{db: db} }

func (r *scoreTrendRepository) AggregateDaily(ctx context.Context, userID uint64, days int) ([]domain.ScoreTrendPoint, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT to_char(created_at, 'YYYY-MM-DD') AS d, AVG(overall_score) AS avg_score
		FROM score_cards
		WHERE user_id = ? AND created_at >= NOW() - INTERVAL '1 day' * ?
		GROUP BY d
		ORDER BY d
	`, userID, days).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []domain.ScoreTrendPoint
	for rows.Next() {
		var p domain.ScoreTrendPoint
		if err := rows.Scan(&p.Date, &p.OverallScore); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}
