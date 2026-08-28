package usecase

import (
	"context"
	"encoding/hex"
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"

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
	// Slug は空でよい。空なら自動採番する — URL に使う名前は利用者に決めさせない
	// （ユーザー決定 2026-08-28。人が付けた名前は衝突・改名の欲求・情報の漏れを生む）。
	Slug string
	Name string
	// OwnerUserID は作成者。この人が主体（kind='user'）になり admin の grant を受け取る。
	OwnerUserID uint64
}

func (u *CreateWorkspaceUseCase) Execute(ctx context.Context, in CreateWorkspaceInput) (*domain.Workspace, error) {
	if in.OwnerUserID == 0 {
		return nil, errors.New("ownerUserID is required")
	}
	autoSlug := in.Slug == ""
	if autoSlug {
		in.Slug = generatedURLKey("w")
	}
	if !domain.ValidWorkspaceSlug(in.Slug) {
		return nil, ErrInvalidWorkspaceSlug
	}
	if !validDisplayName(in.Name, domain.WorkspaceNameMaxLen) {
		return nil, ErrInvalidName
	}
	for {
		w, err := u.provisioner.ProvisionWorkspace(ctx, repository.WorkspaceProvisionInput{
			Slug:        in.Slug,
			Name:        in.Name,
			OwnerUserID: in.OwnerUserID,
		})
		// 自動採番が衝突したら引き直す（48bit の乱数なので実際にはほぼ起きないが、
		// 起きたときに利用者へ 409 を見せる理由が無い）。人が指定した slug の 409 はそのまま返す。
		if autoSlug && errors.Is(err, repository.ErrWorkspaceSlugTaken) {
			in.Slug = generatedURLKey("w")
			continue
		}
		return w, err
	}
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
	// Key は空でよい。空なら自動採番する（ワークスペースの slug と同じ方針）。
	Key  string
	Name string
}

func (u *CreateSpaceUseCase) Execute(ctx context.Context, in CreateSpaceInput) (*domain.Space, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	autoKey := in.Key == ""
	if autoKey {
		in.Key = generatedURLKey("s")
	}
	if !domain.ValidSpaceKey(in.Key) {
		return nil, ErrInvalidSpaceKey
	}
	if !validDisplayName(in.Name, domain.SpaceNameMaxLen) {
		return nil, ErrInvalidName
	}
	for {
		space := &domain.Space{WorkspaceID: in.WorkspaceID, Key: in.Key, Name: in.Name}
		err := u.repo.CreateSpace(ctx, space)
		// 自動採番の衝突は引き直す（ワークスペースの slug と同じ方針）。
		if autoKey && errors.Is(err, repository.ErrSpaceKeyTaken) {
			in.Key = generatedURLKey("s")
			continue
		}
		if err != nil {
			return nil, err
		}
		return space, nil
	}
}

// RenameSpaceUseCase はスペースの表示名だけを変える。
//
// key は変えない。key は URL とスペース識別の一部で、変えると共有済みの場所が全部外れる。
// 表示名は人が読むための欄なので自由に変えてよい — この非対称が、2 つを別の欄に
// 分けている理由そのもの。
//
// 誰が変えられるか（スペースの実効権限）の判定は handler が CheckSpacePermissionUseCase で
// 先に行う（CreateSpace と同じ分担）。
type RenameSpaceUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewRenameSpaceUseCase(r repository.KnowledgeBaseRepository) *RenameSpaceUseCase {
	return &RenameSpaceUseCase{repo: r}
}

type RenameSpaceInput struct {
	WorkspaceID string
	SpaceID     string
	Name        string
}

func (u *RenameSpaceUseCase) Execute(ctx context.Context, in RenameSpaceInput) (*domain.Space, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	if !validDisplayName(in.Name, domain.SpaceNameMaxLen) {
		return nil, ErrInvalidName
	}
	if err := u.repo.UpdateSpaceName(ctx, in.WorkspaceID, in.SpaceID, in.Name); err != nil {
		return nil, err
	}
	// 更新後の姿を読み直して返す（updated_at は DB の now() が入るため、書いた値では作れない）。
	return u.repo.FindSpace(ctx, in.WorkspaceID, in.SpaceID)
}

// validDisplayName は表示名が空でなく列幅（文字数）に収まるかを返す。
// 列は varchar(n) で「文字数」の上限なので、バイト数ではなくルーン数で数える。
func validDisplayName(name string, maxLen int) bool {
	return name != "" && utf8.RuneCountInString(name) <= maxLen
}

// generatedURLKey は slug / key の自動採番。UUID の先頭 12 桁（16 進）を使う。
// 短い連番にしないのは、URL の識別子から作成順・総数が読めてしまうため。
// 12 桁（48 ビット）なら 1 ワークスペースの規模で衝突は事実上起きず、
// 万一衝突しても一意制約が 409 で止める（黙って上書きにはならない）。
func generatedURLKey(prefix string) string {
	id := uuid.New()
	return prefix + "-" + hex.EncodeToString(id[:6])
}
