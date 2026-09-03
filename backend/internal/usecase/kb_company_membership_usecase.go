package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// JoinCompanyWorkspaceUseCase は「その人の会社のワークスペース」へ自動で入れる。

type JoinCompanyWorkspaceUseCase struct {
	permissions repository.KnowledgeBasePermissionRepository
	users       repository.UserRepository
}

func NewJoinCompanyWorkspaceUseCase(p repository.KnowledgeBasePermissionRepository, u repository.UserRepository) *JoinCompanyWorkspaceUseCase {
	return &JoinCompanyWorkspaceUseCase{permissions: p, users: u}
}

type JoinCompanyWorkspaceInput struct {
	UserID uint64
}

// userWorkspaceID はユーザーの所属ワークスペース ID を返す（users.workspace_id の直読み）。
func userWorkspaceID(ctx context.Context, users repository.UserRepository, userID uint64) (string, error) {
	u, err := users.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if u == nil || u.WorkspaceID == nil {
		return "", repository.ErrWorkspaceNotFound
	}
	return *u.WorkspaceID, nil
}

// Execute は会社のワークスペースへの所属を用意し、そのワークスペース ID を返す。
func (u *JoinCompanyWorkspaceUseCase) Execute(
	ctx context.Context, in JoinCompanyWorkspaceInput,
) (string, error) {
	if in.UserID == 0 {
		return "", errors.New("userID is required")
	}
	workspaceID, err := userWorkspaceID(ctx, u.users, in.UserID)
	if err != nil {
		return "", err
	}
	// 既に主体があるなら、この人の所属も役割も既に決まっている。何もしない。
	// ここで役割を足すと、取り消したはずの権限が次の読み取りで戻る。
	if _, err := u.permissions.FindUserPrincipal(ctx, workspaceID, in.UserID); err == nil {
		return workspaceID, nil
	} else if !errors.Is(err, repository.ErrPrincipalNotFound) {
		return "", err
	}

	principal, err := u.permissions.EnsureUserPrincipal(ctx, workspaceID, in.UserID)
	if err != nil {
		return "", err
	}
	// ここへ来るのは主体を新しく作ったときだけ。最初の役割を与える。
	if err := u.permissions.GrantWorkspaceRoleIfAbsent(
		ctx, workspaceID, principal.ID, domain.GrantRoleEditor,
	); err != nil {
		return "", err
	}
	return workspaceID, nil
}
