package repository

import (
	"context"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// ProfileRepository は profiles テーブルへのアクセスを提供する。
type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (*domain.Profile, error)
	Upsert(ctx context.Context, p *domain.Profile) error
	// UpdateStatus は一言ステータス（絵文字・テキスト・失効時刻）だけを upsert する
	// （段 14。PUT /me/status 用）。bio / avatar_url には触れない。
	UpdateStatus(ctx context.Context, userID uint64, emoji, text string, expiresAt *time.Time) (*domain.Profile, error)
}
