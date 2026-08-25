package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ResolveWorkspaceUseCase は URL の slug と現在のユーザーから、操作対象のワークスペースを決める。
// ナレッジ基盤の HTTP 経路はすべてここを通ってテナントを確定させる
// （クライアントが送った workspace_id をそのまま信じる経路を作らない）。
//
// 所属していない slug も存在しない slug も、どちらも repository.ErrWorkspaceNotFound を返す。
// 呼び出し側で 403 と 404 を撃ち分けられるようにすると、slug（短く推測しやすい文字列）を
// 総当たりするだけでテナントの実在が分かってしまうため、区別自体をここで潰しておく。
type ResolveWorkspaceUseCase struct {
	workspaces  repository.KnowledgeBaseRepository
	permissions repository.KnowledgeBasePermissionRepository
}

func NewResolveWorkspaceUseCase(
	w repository.KnowledgeBaseRepository,
	p repository.KnowledgeBasePermissionRepository,
) *ResolveWorkspaceUseCase {
	return &ResolveWorkspaceUseCase{workspaces: w, permissions: p}
}

type ResolveWorkspaceInput struct {
	// Slug は URL に出るワークスペースの識別子。
	Slug string
	// UserID は現在ログインしているユーザー（users.id）。
	UserID uint64
}

func (u *ResolveWorkspaceUseCase) Execute(ctx context.Context, in ResolveWorkspaceInput) (*domain.Workspace, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.Slug == "" {
		return nil, repository.ErrWorkspaceNotFound
	}
	ws, err := u.workspaces.FindWorkspaceBySlug(ctx, in.Slug)
	if err != nil {
		return nil, err
	}
	// 所属の正本は principals（kind='user'）の行の有無。専用のメンバーシップ表は持たない。
	member, err := u.permissions.IsWorkspaceMember(ctx, ws.ID, in.UserID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, repository.ErrWorkspaceNotFound
	}
	return ws, nil
}
