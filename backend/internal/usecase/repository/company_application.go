package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CompanyApplicationRepository は company_applications テーブルへのアクセスを提供する。
type CompanyApplicationRepository interface {
	Create(ctx context.Context, app *domain.CompanyApplication) error
	// ListAll は全申請を新しい順で返す（super_admin の一覧用）。
	ListAll(ctx context.Context) ([]domain.CompanyApplication, error)
	UpdateStatus(ctx context.Context, id uint64, status string) error
}
