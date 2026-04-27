package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type FavoritePhraseRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.FavoritePhrase, error)
	Create(ctx context.Context, p *domain.FavoritePhrase) error
	Delete(ctx context.Context, id uint64) error
}

type favoritePhraseRepository struct{ db *gorm.DB }

func NewFavoritePhraseRepository(db *gorm.DB) FavoritePhraseRepository {
	return &favoritePhraseRepository{db: db}
}

func (r *favoritePhraseRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.FavoritePhrase, error) {
	var rows []domain.FavoritePhrase
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *favoritePhraseRepository) Create(ctx context.Context, p *domain.FavoritePhrase) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *favoritePhraseRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.FavoritePhrase{}, id).Error
}
