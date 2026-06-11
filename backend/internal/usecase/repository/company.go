package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CompanyRepository は companies テーブルへのアクセスを提供する。
type CompanyRepository interface {
	ListAll(ctx context.Context) ([]domain.Company, error)
	FindByID(ctx context.Context, id uint64) (*domain.Company, error)
	// UpdateAiChatEnabled は trainee への AI チャット許可フラグを更新する。
	UpdateAiChatEnabled(ctx context.Context, companyID uint64, enabled bool) error
	// UpdateActive は会社アカウントの有効/無効を更新する（false で無効化）。
	UpdateActive(ctx context.Context, companyID uint64, active bool) error
}
