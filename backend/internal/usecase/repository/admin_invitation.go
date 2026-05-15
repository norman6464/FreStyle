package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// AdminInvitationRepository は admin_invitations テーブル へ の アクセス を 提供 する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// adminInvitationRepository (GORM)。 招待 受諾 / pending 検索 / 監査 用 status 更新
// の 一連 を 1 fat interface で 集約。
type AdminInvitationRepository interface {
	// ListAll は全社横断で招待を返す。SuperAdmin (運営) のダッシュボード用。
	ListAll(ctx context.Context) ([]domain.AdminInvitation, error)
	// ListByCompanyID は CompanyAdmin が自社の招待のみを見るための query。
	ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error)
	// FindPendingByEmail は招待受諾フローで「ログインしてきたユーザーが招待ユーザーか」
	// を判定するために使う。同一 email で複数 pending があれば最新を返す。
	FindPendingByEmail(ctx context.Context, email string) (*domain.AdminInvitation, error)
	// FindPendingByToken はマジックリンク受諾フロー用。token 一致 & status=pending & expires_at 未経過のみ返す。
	// 該当なしは (nil, nil) を返し、呼び出し側で「無効/期限切れ token」として扱う。
	FindPendingByToken(ctx context.Context, token string) (*domain.AdminInvitation, error)
	Create(ctx context.Context, inv *domain.AdminInvitation) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
}
