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
// 既に役割があれば触らない。admin を editor へ格下げしない。
//
// # プライベートスペースには届かない
//
// ここで与えるのはワークスペース全体の grant で、visibility='private' のスペースには
// 届かない（事実を集める側のクエリがふるう）。会社の全員が入っても、
// プライベートは付与された人にだけ見える。
type JoinCompanyWorkspaceUseCase struct {
	permissions repository.KnowledgeBasePermissionRepository
}

func NewJoinCompanyWorkspaceUseCase(p repository.KnowledgeBasePermissionRepository) *JoinCompanyWorkspaceUseCase {
	return &JoinCompanyWorkspaceUseCase{permissions: p}
}

type JoinCompanyWorkspaceInput struct {
	UserID uint64
}

// Execute は会社のワークスペースへの所属を用意し、そのワークスペース ID を返す。
// 会社に属していないユーザーは ErrWorkspaceNotFound（入れる先が無い）。
func (u *JoinCompanyWorkspaceUseCase) Execute(
	ctx context.Context, in JoinCompanyWorkspaceInput,
) (string, error) {
	if in.UserID == 0 {
		return "", errors.New("userID is required")
	}
	workspaceID, err := u.permissions.FindUserCompanyWorkspaceID(ctx, in.UserID)
	if err != nil {
		return "", err
	}
	principal, err := u.permissions.EnsureUserPrincipal(ctx, workspaceID, in.UserID)
	if err != nil {
		return "", err
	}
	// 既定の役割は「無いときだけ」与える。既にある役割は触らない
	// （admin を editor へ落とさない・個別に絞った設定を毎回踏み潰さない）。
	if err := u.permissions.GrantWorkspaceRoleIfAbsent(
		ctx, workspaceID, principal.ID, domain.GrantRoleEditor,
	); err != nil {
		return "", err
	}
	return workspaceID, nil
}
