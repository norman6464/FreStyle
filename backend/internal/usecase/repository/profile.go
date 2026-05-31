package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ProfileRepository は profiles テーブルへのアクセスを提供する。
type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (*domain.Profile, error)
	Upsert(ctx context.Context, p *domain.Profile) error
}
