package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// AdminInvitationRepository は invitations テーブルへのアクセスを提供する。
type AdminInvitationRepository interface {
	// ListAll は全社横断で招待を返す（SuperAdmin 用）。
	ListAll(ctx context.Context) ([]domain.AdminInvitation, error)
	// ListByCompanyID は SuperAdmin が ?companyId= で任意の会社を指定するときに使う。
	ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error)
	// ListByWorkspaceID は CompanyAdmin が自社の招待のみを見る用。
	// FRESTYLE-401（段4横展開）: CompanyAdmin 経路だけを company_id 直読みから
	// workspace_id 経由へ切り替え済み。ListByCompanyID とは呼び分けで、リネームではない
	// （ListByCompanyID は SuperAdmin の任意会社指定に引き続き使うため残している）。
	ListByWorkspaceID(ctx context.Context, workspaceID string) ([]domain.AdminInvitation, error)
	// FindPendingByEmail は同一 email の pending 招待の最新を返す（受諾フロー判定用）。
	FindPendingByEmail(ctx context.Context, email string) (*domain.AdminInvitation, error)
	// FindPendingByToken は token 一致 & pending & 未期限切れのみ返す（該当なしは nil, nil）。
	FindPendingByToken(ctx context.Context, token string) (*domain.AdminInvitation, error)
	// FindByID は ID 一致の招待を返す（該当なしは nil, nil）。会社スコープの認可判定に使う。
	FindByID(ctx context.Context, id uint64) (*domain.AdminInvitation, error)
	Create(ctx context.Context, inv *domain.AdminInvitation) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
}
