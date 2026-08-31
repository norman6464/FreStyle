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
	workspaces  repository.WorkspaceActivationReader
}

func NewValidateInvitationTokenUseCase(
	invitations repository.AdminInvitationRepository,
	workspaces repository.WorkspaceActivationReader,
) *ValidateInvitationTokenUseCase {
	return &ValidateInvitationTokenUseCase{invitations: invitations, workspaces: workspaces}
}

// ValidatedInvitation は受諾画面に表示する最低限の情報。
type ValidatedInvitation struct {
	Role domain.RoleName
	Name string
	// WorkspaceName は招待元の表示名。招待された人が「どこに招かれたのか」を判断する唯一の手掛かり。
	WorkspaceName string
	WorkspaceID   *string
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

	// ワークスペース取得に失敗しても招待自体は有効なので、名前を空にして続行する。
	workspaceName := ""
	if u.workspaces != nil && inv.WorkspaceID != nil {
		if w, err := u.workspaces.FindWorkspaceByID(ctx, *inv.WorkspaceID); err == nil && w != nil {
			workspaceName = w.Name
		}
	}

	return &ValidatedInvitation{
		Role:          normalizeInvitationRole(inv.Role),
		Name:          inv.Name,
		WorkspaceName: workspaceName,
		WorkspaceID:   inv.WorkspaceID,
	}, nil
}

// normalizeInvitationRole は invitation の role を表示用に正規化する（想定外は trainee にフォールバック）。
func normalizeInvitationRole(role domain.RoleName) domain.RoleName {
	switch role {
	case domain.RoleCompanyAdmin, domain.RoleTrainee:
		return role
	default:
		return domain.RoleTrainee
	}
}
