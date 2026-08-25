package usecase

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrInvalidWorkspaceSlug は slug が URL に出せる形（小文字英数字とハイフン）でないときに返す。
var ErrInvalidWorkspaceSlug = errors.New("invalid workspace slug")

// ErrInvalidSpaceKey は key が保存してよい形でないときに返す。
var ErrInvalidSpaceKey = errors.New("invalid space key")

// ErrInvalidName は表示名が空、または列幅（200 文字）を超えるときに返す。
var ErrInvalidName = errors.New("invalid name")

// CreateWorkspaceUseCase はワークスペースを作り、作成者をその admin にする。
//
// 「作れるのは誰か」は認証済みのユーザー全員とする。新しく作るのは中身が空のテナントで、
// 既存のどのワークスペースへのアクセスも増えない（権限は principals / grants で閉じており、
// 別テナントの主体には届かない）。逆に既存のアプリ内ロール（company_admin 等）で
// 絞る案は採らない。ナレッジ基盤の権限は「特権ロールなら通る」という抜け道を
// 持たない設計で、作成だけをアプリ内ロールに結び付けると、権限の出どころが 2 系統になる。
//
// 作成者を admin にするのは repository（1 トランザクション）の責務。ここで
// 「作ってから権限を張る」と 2 手に分けると、片方だけ成功したときに
// 誰も入れないワークスペースが残る。
type CreateWorkspaceUseCase struct {
	provisioner repository.WorkspaceProvisioner
}

func NewCreateWorkspaceUseCase(p repository.WorkspaceProvisioner) *CreateWorkspaceUseCase {
	return &CreateWorkspaceUseCase{provisioner: p}
}

type CreateWorkspaceInput struct {
	Slug string
	Name string
	// OwnerUserID は作成者。この人が主体（kind='user'）になり admin の grant を受け取る。
	OwnerUserID uint64
}

func (u *CreateWorkspaceUseCase) Execute(ctx context.Context, in CreateWorkspaceInput) (*domain.Workspace, error) {
	if in.OwnerUserID == 0 {
		return nil, errors.New("ownerUserID is required")
	}
	if !domain.ValidWorkspaceSlug(in.Slug) {
		return nil, ErrInvalidWorkspaceSlug
	}
	if !validDisplayName(in.Name, domain.WorkspaceNameMaxLen) {
		return nil, ErrInvalidName
	}
	return u.provisioner.ProvisionWorkspace(ctx, repository.WorkspaceProvisionInput{
		Slug:        in.Slug,
		Name:        in.Name,
		OwnerUserID: in.OwnerUserID,
	})
}

// CreateSpaceUseCase はワークスペース配下にスペースを作る。
//
// 誰が作れるか（ワークスペースの実効権限）の判定はここではなく handler が
// CheckWorkspacePermissionUseCase で先に行う。ページ操作と同じ組み立て方に揃えている
// （認可は 1 か所、この usecase は「作る」だけを担う）。
type CreateSpaceUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewCreateSpaceUseCase(r repository.KnowledgeBaseRepository) *CreateSpaceUseCase {
	return &CreateSpaceUseCase{repo: r}
}

type CreateSpaceInput struct {
	WorkspaceID string
	Key         string
	Name        string
}

func (u *CreateSpaceUseCase) Execute(ctx context.Context, in CreateSpaceInput) (*domain.Space, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if !domain.ValidSpaceKey(in.Key) {
		return nil, ErrInvalidSpaceKey
	}
	if !validDisplayName(in.Name, domain.SpaceNameMaxLen) {
		return nil, ErrInvalidName
	}
	space := &domain.Space{WorkspaceID: in.WorkspaceID, Key: in.Key, Name: in.Name}
	if err := u.repo.CreateSpace(ctx, space); err != nil {
		return nil, err
	}
	return space, nil
}

// validDisplayName は表示名が空でなく列幅（文字数）に収まるかを返す。
// 列は varchar(n) で「文字数」の上限なので、バイト数ではなくルーン数で数える。
func validDisplayName(name string, maxLen int) bool {
	return name != "" && utf8.RuneCountInString(name) <= maxLen
}
