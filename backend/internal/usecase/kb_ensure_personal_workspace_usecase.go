package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// EnsurePersonalWorkspaceUseCase は、そのユーザーの個人ワークスペースが既にあれば返し、
// 無ければ作って返す。一意性は DB（uq_workspaces_personal_owner）が守るので、作成が
// repository.ErrPersonalWorkspaceAlreadyExists で競合したら引き直す（check-then-act はしない）。
type EnsurePersonalWorkspaceUseCase struct {
	workspaces  repository.KnowledgeBaseRepository
	provisioner repository.WorkspaceProvisioner
}

func NewEnsurePersonalWorkspaceUseCase(
	w repository.KnowledgeBaseRepository, p repository.WorkspaceProvisioner,
) *EnsurePersonalWorkspaceUseCase {
	return &EnsurePersonalWorkspaceUseCase{workspaces: w, provisioner: p}
}

type EnsurePersonalWorkspaceInput struct {
	UserID uint64
	// Name は新規作成のときだけ使う表示名（例: 利用者の氏名）。既に個人ワークスペースが
	// あるときは無視する（作成後に利用者が改名しているかもしれない値を上書きしない）。
	Name string
}

func (u *EnsurePersonalWorkspaceUseCase) Execute(
	ctx context.Context, in EnsurePersonalWorkspaceInput,
) (*domain.Workspace, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}

	// 大半のログイン（初回サインアップ以外）はここで終わる — 1 回の SELECT。
	if ws, err := u.workspaces.FindPersonalWorkspaceByOwner(ctx, in.UserID); err == nil {
		return ws, nil
	} else if !errors.Is(err, repository.ErrWorkspaceNotFound) {
		return nil, fmt.Errorf("find personal workspace: %w", err)
	}

	slug := generatedURLKey("w")
	ownerID := in.UserID
	for {
		created, err := u.provisioner.ProvisionWorkspace(ctx, repository.WorkspaceProvisionInput{
			Slug:                slug,
			Name:                in.Name,
			OwnerUserID:         in.UserID,
			PersonalOwnerUserID: &ownerID,
		})
		switch {
		case err == nil:
			return created, nil
		case errors.Is(err, repository.ErrWorkspaceSlugTaken):
			// 自動採番の衝突（48bit の乱数なので実際にはほぼ起きない）。引き直す。
			slug = generatedURLKey("w")
			continue
		case errors.Is(err, repository.ErrPersonalWorkspaceAlreadyExists):
			// この判定からこの INSERT までの間に、別のリクエスト（二重送信・同時実行）が
			// 先に作り終えていた。失敗として扱わず、その 1 つを引いて返す。
			ws, findErr := u.workspaces.FindPersonalWorkspaceByOwner(ctx, in.UserID)
			if findErr != nil {
				return nil, fmt.Errorf("find personal workspace after race: %w", findErr)
			}
			return ws, nil
		default:
			return nil, fmt.Errorf("provision personal workspace: %w", err)
		}
	}
}
