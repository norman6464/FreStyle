package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// AdminInvitationRepository は invitations テーブルへのアクセスを提供する。
type AdminInvitationRepository interface {
	// ListAll は全社横断で招待を返す（SuperAdmin 用）。
	ListAll(ctx context.Context) ([]domain.AdminInvitation, error)
	// ListByCompanyID は CompanyAdmin が自社の招待のみを見る用。
	ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error)
	// FindPendingByEmail は同一 email の pending 招待の最新を返す（受諾フロー判定用）。
	FindPendingByEmail(ctx context.Context, email string) (*domain.AdminInvitation, error)
	// FindPendingByToken は token 一致 & pending & 未期限切れのみ返す（該当なしは nil, nil）。
	FindPendingByToken(ctx context.Context, token string) (*domain.AdminInvitation, error)
	Create(ctx context.Context, inv *domain.AdminInvitation) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
}
