package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrInvalidGrantRole は既知でない役割を指定したときに返す。
var ErrInvalidGrantRole = errors.New("invalid grant role")

// ErrInvalidCapability は既知でないケイパビリティを指定したときに返す。
var ErrInvalidCapability = errors.New("invalid capability")

// GrantWorkspaceRoleUseCase はワークスペース全体での既定の役割を主体に与える。
// 配下の全スペースに効くので、テナント全体の管理者はここで 1 行張れば足りる。
type GrantWorkspaceRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewGrantWorkspaceRoleUseCase(r repository.KnowledgeBasePermissionRepository) *GrantWorkspaceRoleUseCase {
	return &GrantWorkspaceRoleUseCase{repo: r}
}

type GrantWorkspaceRoleInput struct {
	WorkspaceID string
	PrincipalID string
	Role        domain.GrantRole
}

func (u *GrantWorkspaceRoleUseCase) Execute(ctx context.Context, in GrantWorkspaceRoleInput) (*domain.WorkspaceGrant, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PrincipalID == "" {
		return nil, errors.New("principalID is required")
	}
	if !in.Role.Valid() {
		return nil, ErrInvalidGrantRole
	}
	// 主体の実在とテナントの一致は DB の複合 FK でも守られるが、先に引いて
	// 「別ワークスペースの ID を渡した」を FK 違反ではなく not found として返す。
	if _, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.repo.UpsertWorkspaceGrant(ctx, in.WorkspaceID, in.PrincipalID, in.Role)
}

// RevokeWorkspaceRoleUseCase はワークスペース全体での既定の役割を剥がす（冪等）。
type RevokeWorkspaceRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRevokeWorkspaceRoleUseCase(r repository.KnowledgeBasePermissionRepository) *RevokeWorkspaceRoleUseCase {
	return &RevokeWorkspaceRoleUseCase{repo: r}
}

type RevokeWorkspaceRoleInput struct {
	WorkspaceID string
	PrincipalID string
}

func (u *RevokeWorkspaceRoleUseCase) Execute(ctx context.Context, in RevokeWorkspaceRoleInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.PrincipalID == "" {
		return errors.New("principalID is required")
	}
	return u.repo.DeleteWorkspaceGrant(ctx, in.WorkspaceID, in.PrincipalID)
}

// GrantSpaceRoleUseCase はスペースでの既定の役割を主体に与える。
type GrantSpaceRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewGrantSpaceRoleUseCase(r repository.KnowledgeBasePermissionRepository) *GrantSpaceRoleUseCase {
	return &GrantSpaceRoleUseCase{repo: r}
}

type GrantSpaceRoleInput struct {
	WorkspaceID string
	SpaceID     string
	PrincipalID string
	Role        domain.GrantRole
}

func (u *GrantSpaceRoleUseCase) Execute(ctx context.Context, in GrantSpaceRoleInput) (*domain.SpaceGrant, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	if in.PrincipalID == "" {
		return nil, errors.New("principalID is required")
	}
	if !in.Role.Valid() {
		return nil, ErrInvalidGrantRole
	}
	if _, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.repo.UpsertSpaceGrant(ctx, in.WorkspaceID, in.SpaceID, in.PrincipalID, in.Role)
}

// RevokeSpaceRoleUseCase はスペースでの既定の役割を剥がす（冪等）。
type RevokeSpaceRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRevokeSpaceRoleUseCase(r repository.KnowledgeBasePermissionRepository) *RevokeSpaceRoleUseCase {
	return &RevokeSpaceRoleUseCase{repo: r}
}

type RevokeSpaceRoleInput struct {
	WorkspaceID string
	SpaceID     string
	PrincipalID string
}

func (u *RevokeSpaceRoleUseCase) Execute(ctx context.Context, in RevokeSpaceRoleInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return errors.New("spaceID is required")
	}
	if in.PrincipalID == "" {
		return errors.New("principalID is required")
	}
	return u.repo.DeleteSpaceGrant(ctx, in.WorkspaceID, in.SpaceID, in.PrincipalID)
}

// GrantPageRoleUseCase はページでの既定の役割を主体に与える。
//
// 既定の 3 段目（ワークスペース → スペース → ページ）で、このページとその子孫に効く。
// 合成は上の 2 段と同じで、複数の経路から届いた役割のうち最も強いものが実効になる。
//
// **これで誰かを弱めることはできない。** 上位で editor を得ている相手にここで viewer を
// 張っても editor のままで、下げたつもりが効かない。付与はどこまでも足し算だけで、
// 打ち消す層は持たない（domain.GrantRole.Rank と domain.PagePermissionFacts に規則と理由がある）。
// 狭めたい内容は private のスペースへ置く。
type GrantPageRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewGrantPageRoleUseCase(r repository.KnowledgeBasePermissionRepository) *GrantPageRoleUseCase {
	return &GrantPageRoleUseCase{repo: r}
}

type GrantPageRoleInput struct {
	WorkspaceID string
	PageID      string
	PrincipalID string
	Role        domain.GrantRole
}

func (u *GrantPageRoleUseCase) Execute(ctx context.Context, in GrantPageRoleInput) (*domain.PageGrant, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	if in.PrincipalID == "" {
		return nil, errors.New("principalID is required")
	}
	if !in.Role.Valid() {
		return nil, ErrInvalidGrantRole
	}
	if _, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.repo.UpsertPageGrant(ctx, in.WorkspaceID, in.PageID, in.PrincipalID, in.Role)
}

// RevokePageRoleUseCase はページでの既定の役割を剥がす（冪等）。
//
// 消えるのはこの段で足した分だけで、ワークスペース / スペース / 祖先のページから
// 届いている役割はそのまま残る。**「このページだけ見せない」は書けない** —
// 狭めたい内容は private のスペースへ置く。
//
// 「最後の admin」の検査は要らない。守っているのはワークスペースの admin が 0 人に
// なることで、ページの grant を全部消してもワークスペースの admin は配下の全ページに届く
// （RevokeSpaceRoleUseCase と同じ理由）。
type RevokePageRoleUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRevokePageRoleUseCase(r repository.KnowledgeBasePermissionRepository) *RevokePageRoleUseCase {
	return &RevokePageRoleUseCase{repo: r}
}

type RevokePageRoleInput struct {
	WorkspaceID string
	PageID      string
	PrincipalID string
}

func (u *RevokePageRoleUseCase) Execute(ctx context.Context, in RevokePageRoleInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return errors.New("pageID is required")
	}
	if in.PrincipalID == "" {
		return errors.New("principalID is required")
	}
	return u.repo.DeletePageGrant(ctx, in.WorkspaceID, in.PageID, in.PrincipalID)
}

// ListPageGrantsUseCase はそのページ自身に張られた既定の役割の一覧を返す。
//
// **返るのは「このページを見られる人の一覧」ではない。** この段で足した行だけで、
// 上の段や祖先のページから届いている相手は含まれない。空で返ってきても
// 「誰も見られない」ではなく「この段では何も足していない」の意味になる。
// 呼び出し側（画面）はそれが分かる見せ方をすること。
type ListPageGrantsUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListPageGrantsUseCase(r repository.KnowledgeBasePermissionRepository) *ListPageGrantsUseCase {
	return &ListPageGrantsUseCase{repo: r}
}

type ListPageGrantsInput struct {
	WorkspaceID string
	PageID      string
}

func (u *ListPageGrantsUseCase) Execute(ctx context.Context, in ListPageGrantsInput) ([]domain.PageGrant, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	return u.repo.ListPageGrants(ctx, in.WorkspaceID, in.PageID)
}

// ListGrantablePrincipalsUseCase は権限を張れる相手を表示名つきで返す。
//
// 返るのはワークスペース全体の主体で、ページでは絞らない。ページ単位の付与も
// 相手はワークスペースの主体だからで、ここで絞る意味が無い（絞ると
// 「同じ人に張れるはずなのに一覧に出ない」というずれが生まれる）。
//
// 呼べる範囲は handler 側の gate が決める。この一覧を使うのは「そのページの権限を
// 変えられる人」なので、認可もページ単位で掛ける。
type ListGrantablePrincipalsUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListGrantablePrincipalsUseCase(r repository.KnowledgeBasePermissionRepository) *ListGrantablePrincipalsUseCase {
	return &ListGrantablePrincipalsUseCase{repo: r}
}

type ListGrantablePrincipalsInput struct {
	WorkspaceID string
}

func (u *ListGrantablePrincipalsUseCase) Execute(
	ctx context.Context, in ListGrantablePrincipalsInput,
) ([]domain.GrantablePrincipal, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	return u.repo.ListGrantablePrincipals(ctx, in.WorkspaceID)
}
