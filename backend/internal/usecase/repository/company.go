package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CompanyRepository は companies テーブルへのアクセスを提供する。
type CompanyRepository interface {
	ListAll(ctx context.Context) ([]domain.Company, error)
	FindByID(ctx context.Context, id uint64) (*domain.Company, error)
	// FindByWorkspaceID は対応ワークスペースから会社を引く（1:1。該当なしは nil, nil）。
	FindByWorkspaceID(ctx context.Context, workspaceID string) (*domain.Company, error)
	// UpdateActive は会社アカウントの有効/無効を更新する（false で無効化）。
	UpdateActive(ctx context.Context, companyID uint64, active bool) error
}
