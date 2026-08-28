package usecase

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrPrincipalKindMismatch は主体の種類が操作に合わないときに返す
// （グループでないものをグループとして扱おうとした等）。
var ErrPrincipalKindMismatch = errors.New("principal kind does not match the operation")

// kbGroupNameMaxLen は principals.name (varchar(200)) の上限。DB エラーの前に入口で弾く。
const kbGroupNameMaxLen = 200

// AddWorkspaceMemberUseCase はユーザーをワークスペースのメンバーにする。
// 所属は principals（kind='user'）の 1 行で表すので、この usecase はその行を作る（冪等）。
type AddWorkspaceMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewAddWorkspaceMemberUseCase(r repository.KnowledgeBasePermissionRepository) *AddWorkspaceMemberUseCase {
	return &AddWorkspaceMemberUseCase{repo: r}
}

type AddWorkspaceMemberInput struct {
	WorkspaceID string
	UserID      uint64
}

func (u *AddWorkspaceMemberUseCase) Execute(ctx context.Context, in AddWorkspaceMemberInput) (*domain.Principal, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	principal, err := u.repo.EnsureUserPrincipal(ctx, in.WorkspaceID, in.UserID)
	if err != nil {
		return nil, err
	}
	// 追加した瞬間から**全員が書ける**（ユーザー決定 2026-08-28）。
	// 既定を viewer にすると「入れたのに書けない」問い合わせが管理者に集まり、
	// 結局全員に editor を配って回ることになる。絞りたいスペース・ページは
	// 個別の grant / 例外で狭める（広い既定 + 狭い例外、の向きに揃える）。
	// **無いときだけ**与える（上書きしない）。追加は冪等で、既に admin の人へ
	// もう一度実行され得るため、上書きだと admin が editor に落ちる。
	if gerr := u.repo.GrantWorkspaceRoleIfAbsent(ctx, in.WorkspaceID, principal.ID, domain.GrantRoleEditor); gerr != nil {
		return nil, gerr
	}
	return principal, nil
}

// RemoveWorkspaceMemberUseCase はユーザーをワークスペースから外す。
// principal を消すと、その人に張られていた grant / restriction / グループ所属も
// FK の CASCADE で消える（権限だけが残らない）。
type RemoveWorkspaceMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRemoveWorkspaceMemberUseCase(r repository.KnowledgeBasePermissionRepository) *RemoveWorkspaceMemberUseCase {
	return &RemoveWorkspaceMemberUseCase{repo: r}
}

type RemoveWorkspaceMemberInput struct {
	WorkspaceID string
	UserID      uint64
}

func (u *RemoveWorkspaceMemberUseCase) Execute(ctx context.Context, in RemoveWorkspaceMemberInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return errors.New("userID is required")
	}
	principal, err := u.repo.FindUserPrincipal(ctx, in.WorkspaceID, in.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalNotFound) {
			return nil // 既に非メンバー（冪等）
		}
		return err
	}
	return u.repo.DeletePrincipal(ctx, in.WorkspaceID, principal.ID)
}

// CreatePrincipalGroupUseCase は権限をまとめて張るためのグループを作る。
// 名前はワークスペース内で一意（同名が 2 つあると権限を張る先を人が選べない）。
type CreatePrincipalGroupUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCreatePrincipalGroupUseCase(r repository.KnowledgeBasePermissionRepository) *CreatePrincipalGroupUseCase {
	return &CreatePrincipalGroupUseCase{repo: r}
}

type CreatePrincipalGroupInput struct {
	WorkspaceID string
	Name        string
}

func (u *CreatePrincipalGroupUseCase) Execute(ctx context.Context, in CreatePrincipalGroupInput) (*domain.Principal, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	if utf8.RuneCountInString(in.Name) > kbGroupNameMaxLen {
		return nil, errors.New("name is too long")
	}
	return u.repo.CreateGroupPrincipal(ctx, in.WorkspaceID, in.Name)
}

// AddGroupMemberUseCase はグループにユーザーを加える。
//
// 加える相手を主体 ID ではなくユーザー ID で受けるのは、グループの入れ子をこの入口から
// 作れないようにするため（DB 側も複合 FK で member を kind='user' に固定している）。
// 入れ子を許すと権限解決に再帰が要り、グループ同士の循環も防がなければならなくなる。
type AddGroupMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewAddGroupMemberUseCase(r repository.KnowledgeBasePermissionRepository) *AddGroupMemberUseCase {
	return &AddGroupMemberUseCase{repo: r}
}

type AddGroupMemberInput struct {
	WorkspaceID string
	// GroupPrincipalID は kind='group' の主体。
	GroupPrincipalID string
	// MemberUserID は加えるユーザー。メンバーでなければ主体が無いのでエラーになる。
	MemberUserID uint64
}

func (u *AddGroupMemberUseCase) Execute(ctx context.Context, in AddGroupMemberInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.GroupPrincipalID == "" {
		return errors.New("groupPrincipalID is required")
	}
	if in.MemberUserID == 0 {
		return errors.New("memberUserID is required")
	}
	group, err := u.repo.FindPrincipal(ctx, in.WorkspaceID, in.GroupPrincipalID)
	if err != nil {
		return err
	}
	if group.Kind != domain.PrincipalKindGroup {
		return ErrPrincipalKindMismatch
	}
	member, err := u.repo.FindUserPrincipal(ctx, in.WorkspaceID, in.MemberUserID)
	if err != nil {
		return err
	}
	return u.repo.AddGroupMember(ctx, in.WorkspaceID, group.ID, member.ID)
}

// RemoveGroupMemberUseCase はグループからユーザーを外す（冪等）。
type RemoveGroupMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewRemoveGroupMemberUseCase(r repository.KnowledgeBasePermissionRepository) *RemoveGroupMemberUseCase {
	return &RemoveGroupMemberUseCase{repo: r}
}

type RemoveGroupMemberInput struct {
	WorkspaceID      string
	GroupPrincipalID string
	MemberUserID     uint64
}

func (u *RemoveGroupMemberUseCase) Execute(ctx context.Context, in RemoveGroupMemberInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.GroupPrincipalID == "" {
		return errors.New("groupPrincipalID is required")
	}
	if in.MemberUserID == 0 {
		return errors.New("memberUserID is required")
	}
	member, err := u.repo.FindUserPrincipal(ctx, in.WorkspaceID, in.MemberUserID)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalNotFound) {
			return nil // 非メンバーはどのグループにも属していない（冪等）
		}
		return err
	}
	return u.repo.RemoveGroupMember(ctx, in.WorkspaceID, in.GroupPrincipalID, member.ID)
}

// EnsureSpaceEveryonePrincipalUseCase はスペースの「全員」を表す主体を用意する（冪等）。
// 「既定でチーム全員が編集できる」を 1 行の grant で表すための下ごしらえ。
type EnsureSpaceEveryonePrincipalUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewEnsureSpaceEveryonePrincipalUseCase(r repository.KnowledgeBasePermissionRepository) *EnsureSpaceEveryonePrincipalUseCase {
	return &EnsureSpaceEveryonePrincipalUseCase{repo: r}
}

type EnsureSpaceEveryonePrincipalInput struct {
	WorkspaceID string
	SpaceID     string
}

func (u *EnsureSpaceEveryonePrincipalUseCase) Execute(ctx context.Context, in EnsureSpaceEveryonePrincipalInput) (*domain.Principal, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	return u.repo.EnsureSpaceEveryonePrincipal(ctx, in.WorkspaceID, in.SpaceID)
}
