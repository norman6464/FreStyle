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

// ErrInvalidRestrictionMode は既知でない例外の向きを指定したときに返す。
var ErrInvalidRestrictionMode = errors.New("invalid restriction mode")

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

// SetPageRestrictionUseCase はページ以下だけ既定を上書きする例外を設定する。
//
// allow を 1 つ足すと、そのページのそのケイパビリティは「載っている主体だけ」の限定公開に
// 切り替わる（domain.ResolvePagePermission の規則 3）。deny だけを足した場合は
// 名指しした主体だけが外れ、ほかの人の既定は変わらない。
type SetPageRestrictionUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewSetPageRestrictionUseCase(r repository.KnowledgeBasePermissionRepository) *SetPageRestrictionUseCase {
	return &SetPageRestrictionUseCase{repo: r}
}

type SetPageRestrictionInput struct {
	WorkspaceID string
	PageID      string
	PrincipalID string
	Capability  domain.Capability
	Mode        domain.RestrictionMode
}

func (u *SetPageRestrictionUseCase) Execute(ctx context.Context, in SetPageRestrictionInput) (*domain.PageRestriction, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	if in.PrincipalID == "" {
		return nil, errors.New("principalID is required")
	}
	if !in.Capability.Valid() {
		return nil, ErrInvalidCapability
	}
	if !in.Mode.Valid() {
		return nil, ErrInvalidRestrictionMode
	}
	if _, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.repo.UpsertPageRestriction(ctx, in.WorkspaceID, in.PageID, in.PrincipalID, in.Capability, in.Mode)
}

// ClearPageRestrictionUseCase はページの例外を解除する（冪等）。
// その段の最後の 1 行が消えると、解決はより遠い祖先の制限 → grant の既定へ戻る。
type ClearPageRestrictionUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewClearPageRestrictionUseCase(r repository.KnowledgeBasePermissionRepository) *ClearPageRestrictionUseCase {
	return &ClearPageRestrictionUseCase{repo: r}
}

type ClearPageRestrictionInput struct {
	WorkspaceID string
	PageID      string
	PrincipalID string
	Capability  domain.Capability
}

func (u *ClearPageRestrictionUseCase) Execute(ctx context.Context, in ClearPageRestrictionInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return errors.New("pageID is required")
	}
	if in.PrincipalID == "" {
		return errors.New("principalID is required")
	}
	if !in.Capability.Valid() {
		return ErrInvalidCapability
	}
	return u.repo.DeletePageRestriction(ctx, in.WorkspaceID, in.PageID, in.PrincipalID, in.Capability)
}
