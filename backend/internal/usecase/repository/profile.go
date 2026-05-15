package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ProfileRepository は profiles テーブル へ の アクセス を 提供 する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// profileRepository (GORM)。
type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (*domain.Profile, error)
	Upsert(ctx context.Context, p *domain.Profile) error
}
