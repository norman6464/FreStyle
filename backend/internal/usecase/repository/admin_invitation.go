package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// AdminInvitationRepository は invitations テーブルへのアクセスを提供する。
type AdminInvitationRepository interface {
	// ListAll は全社横断で招待を返す（SuperAdmin 用）。
	ListAll(ctx context.Context) ([]domain.AdminInvitation, error)
	// ListByWorkspaceID はワークスペース単位の招待一覧を返す（CompanyAdmin の自社一覧と、
	// SuperAdmin が任意の会社を指定するときの両方で使う）。
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
