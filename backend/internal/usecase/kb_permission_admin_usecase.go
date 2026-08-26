package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CanRemoveWorkspaceAdminUseCase は「この相手からワークスペースの admin を外しても、
// admin が 1 人以上残るか」に答える。権限を減らす操作の前に呼び、false なら断る。
//
// # なぜこの問いが要るのか（「最後の admin」を剥がせなくする理由）
//
// ナレッジ基盤の権限は principals / grants / restrictions だけで閉じており、
// 「アプリの super_admin なら通る」という抜け道を意図的に持たない（domain/grant.go）。
// その裏返しとして、ワークスペースの admin が 0 人になった瞬間、そのワークスペースの
// 権限を変えられる人は API のどこにも居なくなる。スペースを増やすことも、
// 誰かに権限を戻すこともできず、復旧手段は DB を直接触ることだけになる。
//
// 逆に「最後の 1 人は自分を外せない」で詰まる場面は、先に別の誰かへ admin を渡せば
// 必ず解ける。取り返しがつかない側（0 人）を禁じ、手数が 1 つ増えるだけの側を許す。
//
// # 何を数えるか
//
// 数えるのは kind='user' の主体が持つ admin だけ。グループ宛ての admin を数に入れると、
// メンバーが 1 人も居ないグループが「最後の admin」として残り、結局誰も権限を
// 変えられないワークスペースが同じようにできてしまう（grant の行からはグループの
// 中身が分からない）。その分だけ判定は厳しくなるが、余計に断られるのは
// 「グループ経由の admin しか居ないのに、ユーザー宛ての admin を外そうとした」場合だけで、
// 誰か 1 人に admin を張れば必ず通る。安全側に外れる。
//
// # 競合について（既知の限界）
//
// この確認と実際の取り消しは別のトランザクションなので、2 人の admin をほぼ同時に
// 外す要求が両方この検査を通り抜ける余地は残る。単一の SQL で守るには取り消しの
// DELETE 自体に「他に admin が居ること」を条件として書く必要があり、それは
// repository の設計変更になる。ここで塞ぐのは日常的な誤操作
// （1 人しか居ないと分かっている状態で外す）で、競合は未対応。
type CanRemoveWorkspaceAdminUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCanRemoveWorkspaceAdminUseCase(r repository.KnowledgeBasePermissionRepository) *CanRemoveWorkspaceAdminUseCase {
	return &CanRemoveWorkspaceAdminUseCase{repo: r}
}

// CanRemoveWorkspaceAdminInput は対象を主体 ID かユーザー ID のどちらかで指す。
// grant の取り消しは主体 ID を、メンバーの削除はユーザー ID を持っているため両方を受ける
// （メンバーを消すと principal ごと消え、その主体の grant も CASCADE で消えるので、
// 「grant を外す」と同じ影響がある）。
type CanRemoveWorkspaceAdminInput struct {
	WorkspaceID string
	// PrincipalID は admin を外す相手の主体。空なら UserID から引き直す。
	PrincipalID string
	// UserID は admin を外す相手をユーザーで指すときに使う。
	UserID uint64
}

func (u *CanRemoveWorkspaceAdminUseCase) Execute(ctx context.Context, in CanRemoveWorkspaceAdminInput) (bool, error) {
	if in.WorkspaceID == "" {
		return false, errors.New("workspaceID is required")
	}
	if in.PrincipalID == "" && in.UserID == 0 {
		return false, errors.New("principalID or userID is required")
	}

	target := in.PrincipalID
	if target == "" {
		principal, err := u.repo.FindUserPrincipal(ctx, in.WorkspaceID, in.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrPrincipalNotFound) {
				// 非メンバーは grant を 1 つも持たない。外しても admin は減らない。
				return true, nil
			}
			return false, err
		}
		target = principal.ID
	}

	grants, err := u.repo.ListWorkspaceGrants(ctx, in.WorkspaceID)
	if err != nil {
		return false, err
	}
	targetIsAdmin := false
	others := make([]string, 0, len(grants))
	for _, g := range grants {
		if g.Role != domain.GrantRoleAdmin {
			continue
		}
		if g.PrincipalID == target {
			targetIsAdmin = true
			continue
		}
		others = append(others, g.PrincipalID)
	}
	if !targetIsAdmin {
		// 元から admin ではない相手なので、この操作で admin は 1 人も減らない。
		return true, nil
	}

	// 残る admin のうち、実際に人が入っていると確実に言えるもの（kind='user'）を探す。
	// 1 人でも見つかればそこで打ち切る（admin は多くないが、全件引く必要もない）。
	for _, principalID := range others {
		p, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, principalID)
		if err != nil {
			if errors.Is(err, repository.ErrPrincipalNotFound) {
				// grant を読んでから主体を引くまでの間に消えた。数に入れない（安全側）。
				continue
			}
			return false, err
		}
		if p.Kind == domain.PrincipalKindUser {
			return true, nil
		}
	}
	return false, nil
}
