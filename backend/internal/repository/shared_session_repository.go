package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type SharedSessionRepository interface {
	ListPublic(ctx context.Context, limit int) ([]domain.SharedSession, error)
	Create(ctx context.Context, s *domain.SharedSession) error
}

type sharedSessionRepository struct{ db *gorm.DB }

func NewSharedSessionRepository(db *gorm.DB) SharedSessionRepository {
	return &sharedSessionRepository{db: db}
}

func (r *sharedSessionRepository) ListPublic(ctx context.Context, limit int) ([]domain.SharedSession, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows []domain.SharedSession
	err := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *sharedSessionRepository) Create(ctx context.Context, s *domain.SharedSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}
