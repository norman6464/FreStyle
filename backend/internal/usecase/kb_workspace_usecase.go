package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ResolveWorkspaceUseCase は URL の slug と現在のユーザーから、操作対象のワークスペースを決める。
// ノートの HTTP 経路はすべてここを通ってテナントを確定させる
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
		// 会社のワークスペースなら、まだ principals の行が無いだけなので入れる。
		// URL を直に開いた人も一覧を経ずにここへ来るため、判定の直前で用意する
		// （用意した事実は principals に書くので、所属の表現は 1 つのまま）。
		joined, jerr := u.joinCompany(ctx, ws.ID, in.UserID)
		if jerr != nil {
			return nil, jerr
		}
		if !joined {
			return nil, repository.ErrWorkspaceNotFound
		}
	}
	return ws, nil
}

// joinCompany は「そのワークスペースがこの人の会社のものなら」所属を用意する。
// 会社が違う・会社に属していないなら false（呼び出し側は 404 に倒す）。
func (u *ResolveWorkspaceUseCase) joinCompany(
	ctx context.Context, workspaceID string, userID uint64,
) (bool, error) {
	companyWorkspaceID, err := u.permissions.FindUserCompanyWorkspaceID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return false, nil
		}
		return false, err
	}
	if companyWorkspaceID != workspaceID {
		return false, nil
	}
	// ここに来るのは IsWorkspaceMember が false のときだけなので、主体はまだ無い。
	// 主体を作り、最初の役割を与える。**既にある人には触らない**という規則は
	// JoinCompanyWorkspaceUseCase と同じ（取り消した権限を読み取りで戻さない）。
	principal, err := u.permissions.EnsureUserPrincipal(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}
	if err := u.permissions.GrantWorkspaceRoleIfAbsent(
		ctx, workspaceID, principal.ID, domain.GrantRoleEditor,
	); err != nil {
		return false, err
	}
	return true, nil
}
