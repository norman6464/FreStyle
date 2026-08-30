package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// JoinCompanyWorkspaceUseCase は「その人の会社のワークスペース」へ自動で入れる。
//
// # なぜ要るのか
//
// 会社ごとのワークスペースは起動時のバックフィルが用意し、users.workspace_id へ写している。
// 一方ノートの所属は principals（kind='user'）の行が唯一の表現で、その行は
// ワークスペースを自分で作った人と、明示的に追加された人にしか無い。結果として
// **同じ会社の他のメンバーは、会社のワークスペースを開いても非メンバー扱い**になり、
// 一覧にも出ず URL を叩いても 404 になっていた。
//
// チームスペース（visibility='workspace'）は「会社の中なら誰でも見られる入れ物」なので、
// 会社に属していること自体を所属の根拠にする。ここで principals の行と既定の役割を
// 冪等に用意し、以降は既存の権限解決がそのまま動く（所属の表現は増やさない）。
//
// # 与える役割
//
// editor（読むだけの人を作らない — .claude/CLAUDE.md の「広い既定 + 狭い例外」）。
//
// **役割を与えるのは、主体をこのとき新しく作った場合だけ。** 既に主体がある人には
// 一切触らない。触ると「admin が誰かの役割を取り消したのに、その人が一覧を開いた
// 瞬間に editor へ戻る」という形で権限管理が効かなくなる（役割の取り消しは
// workspace_grants の行を消すだけで、主体の行は残るため）。
//
// # プライベートスペースには届かない
//
// ここで与えるのはワークスペース全体の grant で、visibility='private' のスペースには
// 届かない（事実を集める側のクエリがふるう）。会社の全員が入っても、
// プライベートは付与された人にだけ見える。
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
// 会社に属さない・存在しないユーザーはいずれも ErrWorkspaceNotFound
// （呼び出し側の分岐を増やさない。どちらも「自動で入れる先が無い」で同じ扱いになる）。
// JoinCompanyWorkspaceUseCase と ResolveWorkspaceUseCase.joinCompany の両方から使う。
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
// 会社に属していないユーザーは ErrWorkspaceNotFound（入れる先が無い）。
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
