package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ValidateInvitationTokenUseCase はマジックリンク受諾画面で token を検証する。
// 該当なし / 期限切れ / 既受諾 / canceled はすべて (nil, nil) を返す（メタ情報を漏らさない）。
// 成功時の ValidatedInvitation に email は含めない（token 漏洩時の被害局所化）。
type ValidateInvitationTokenUseCase struct {
	invitations repository.AdminInvitationRepository
	companies   repository.CompanyRepository
}

func NewValidateInvitationTokenUseCase(
	invitations repository.AdminInvitationRepository,
	companies repository.CompanyRepository,
) *ValidateInvitationTokenUseCase {
	return &ValidateInvitationTokenUseCase{invitations: invitations, companies: companies}
}

// ValidatedInvitation は受諾画面に表示する最低限の情報。
type ValidatedInvitation struct {
	Role        string
	DisplayName string
	CompanyID   uint64
	CompanyName string
}

func (u *ValidateInvitationTokenUseCase) Execute(ctx context.Context, token string) (*ValidatedInvitation, error) {
	if token == "" {
		return nil, nil
	}
	inv, err := u.invitations.FindPendingByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, nil
	}

	// company 取得に失敗しても招待自体は有効なので、CompanyName を空にして続行する。
	companyName := ""
	if u.companies != nil {
		if c, err := u.companies.FindByID(ctx, inv.CompanyID); err == nil && c != nil {
			companyName = c.Name
		}
	}

	return &ValidatedInvitation{
		Role:        normalizeInvitationRole(inv.Role),
		DisplayName: inv.DisplayName,
		CompanyID:   inv.CompanyID,
		CompanyName: companyName,
	}, nil
}

// normalizeInvitationRole は invitation の role を表示用に正規化する（想定外は trainee にフォールバック）。
func normalizeInvitationRole(role string) string {
	switch role {
	case domain.RoleCompanyAdmin, domain.RoleTrainee:
		return role
	default:
		return domain.RoleTrainee
	}
}
