package usecase

import (
	"context"
	"fmt"

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

	// 招待先が決まっているなら、その名前は必ず引けるはず。invitations.workspace_id には
	// FK（fk_invitations_workspace）があるので行は必ず在る。引けないのは不整合であって
	// 「名前が無い招待」ではないので、握りつぶさず返す（handler が 500 に写す）。
	// 名前を空にして 200 で通すと、どこに招かれたのか分からないまま受諾させることになる。
	workspaceName := ""
	if inv.WorkspaceID != nil {
		w, err := u.workspaces.FindWorkspaceByID(ctx, *inv.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("find workspace: %w", err)
		}
		if w == nil {
			return nil, fmt.Errorf("招待 %d が指すワークスペース %s が見つかりません", inv.ID, *inv.WorkspaceID)
		}
		workspaceName = w.Name
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
