package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ValidateInvitationTokenUseCase はマジックリンク受諾画面で token を検証する。
//
// 公開エンドポイント (GET /api/v2/invitations/accept/:token) から呼ばれる。
// 該当なし / 期限切れ / 既受諾 / canceled はすべて (nil, nil) を返し、
// 呼び出し側で 404 を返す（メタ情報を漏らさない設計）。
//
// 成功時は ValidatedInvitation (DTO) を返す。email は含めない:
//   - 本人がメールから踏んでくる前提なので email は本人が知っている
//   - 万一 token が漏れた場合の被害局所化（招待先 email を覗かれない）
//
// 依存 port: [repository.AdminInvitationRepository] (token 検証) +
// [repository.CompanyRepository] (company 名 取得)。
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
		// 空 token は repository でも nil 返却するが、handler が空文字を渡してきた場合の
		// ガードを usecase 側にも置いて防御層を二重にする。
		return nil, nil
	}
	inv, err := u.invitations.FindPendingByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, nil
	}

	// 招待の company を引いて name を返す。FindByID が err を返した場合は
	// 招待そのものは有効なので company unknown でも 404 にせず、CompanyName を空で返す
	// （フロントは「招待先を取得できませんでした」とフォールバック表示する想定）。
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

// normalizeInvitationRole は invitation の role を表示用に正規化する。
// 想定外の値が入っていた場合は trainee にフォールバックして UI を壊さないようにする。
func normalizeInvitationRole(role string) string {
	switch role {
	case domain.RoleCompanyAdmin, domain.RoleTrainee:
		return role
	default:
		return domain.RoleTrainee
	}
}
