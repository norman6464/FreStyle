package kb

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// ResolveWorkspaceUseCase は URL の slug と現在のユーザーから、操作対象のワークスペースを決める。
// ナレッジの HTTP 経路はすべてここを通ってテナントを確定させる
// （クライアントが送った workspace_id をそのまま信じる経路を作らない）。
//
// 所属していない slug も存在しない slug も、どちらも repository.ErrWorkspaceNotFound を返す。
// 呼び出し側で 403 と 404 を撃ち分けられるようにすると、slug（短く推測しやすい文字列）を
// 総当たりするだけでテナントの実在が分かってしまうため、区別自体をここで潰しておく。
type ResolveWorkspaceUseCase struct {
	workspaces  repository.KnowledgeBaseRepository
	permissions repository.KnowledgeBasePermissionRepository
	users       repository.UserRepository
}

func NewResolveWorkspaceUseCase(
	w repository.KnowledgeBaseRepository,
	p repository.KnowledgeBasePermissionRepository,
	u repository.UserRepository,
) *ResolveWorkspaceUseCase {
	return &ResolveWorkspaceUseCase{workspaces: w, permissions: p, users: u}
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
	// 停止中のワークスペースは無いものとして扱う。middleware は「叩いた人の所属」しか
	// 見ないので、そこを通り抜けた別のワークスペース（個人用や、principal として参加して
	// いる先）が停止されていても届いてしまう。ナレッジの全 HTTP 経路がこの解決を通るため、
	// ここで塞ぐ。存在を漏らさないよう、権限が無いときと同じ「見つからない」に畳む。
	if !ws.IsActive {
		return nil, repository.ErrWorkspaceNotFound
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
	companyWorkspaceID, err := userWorkspaceID(ctx, u.users, userID)
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

// DeleteWorkspaceUseCase はワークスペースを配下ごと消す。
//
// 誰が消せるか（ワークスペースの admin か）の判定はここではなく handler が
// CheckWorkspacePermissionUseCase で先に行う（認可は 1 か所・ほかの操作と同じ組み立て）。
// 「会社のワークスペースは消さない」という規則だけは repository（さらに SQL）が持つ。
// 認可と違って**誰であっても消してはいけない**ものなので、入口ではなく最も内側で守る。
type DeleteWorkspaceUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewDeleteWorkspaceUseCase(r repository.KnowledgeBaseRepository) *DeleteWorkspaceUseCase {
	return &DeleteWorkspaceUseCase{repo: r}
}

type DeleteWorkspaceInput struct {
	WorkspaceID string
}

func (u *DeleteWorkspaceUseCase) Execute(ctx context.Context, in DeleteWorkspaceInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	return u.repo.DeleteWorkspace(ctx, in.WorkspaceID)
}

// ErrInvalidWorkspaceSlug は slug が URL に出せる形（小文字英数字とハイフン）でないときに返す。
var ErrInvalidWorkspaceSlug = errors.New("invalid workspace slug")

// ErrInvalidSpaceKey は key が保存してよい形でないときに返す。
var ErrInvalidSpaceKey = errors.New("invalid space key")

// ErrInvalidSpaceVisibility は visibility が既知の値（workspace / private）でないときに返す。
var ErrInvalidSpaceVisibility = errors.New("invalid space visibility")

// ErrInvalidName は表示名が空、または列幅（200 文字）を超えるときに返す。
var ErrInvalidName = errors.New("invalid name")

// CreateWorkspaceUseCase はワークスペースを作り、作成者をその admin にする。
//
// 「作れるのは誰か」は認証済みのユーザー全員とする。新しく作るのは中身が空のテナントで、
// 既存のどのワークスペースへのアクセスも増えない（権限は principals / grants で閉じており、
// 別テナントの主体には届かない）。逆に既存のアプリ内ロール（company_admin 等）で
// 絞る案は採らない。ナレッジの権限は「特権ロールなら通る」という抜け道を
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
	// provisioner は private のスペース作成に使う（スペース + 作成者への grant を
	// 1 トランザクションで書く必要があり、単発の CreateSpace では表せない）。
	provisioner repository.WorkspaceProvisioner
}

func NewCreateSpaceUseCase(
	r repository.KnowledgeBaseRepository, p repository.WorkspaceProvisioner,
) *CreateSpaceUseCase {
	return &CreateSpaceUseCase{repo: r, provisioner: p}
}

type CreateSpaceInput struct {
	WorkspaceID string
	// Key は空でよい。空なら自動採番する（ワークスペースの slug と同じ方針）。
	Key  string
	Name string
	// Visibility は空なら 'workspace'（今までどおりの共有スペース）。
	Visibility domain.SpaceVisibility
	// CreatorUserID は作成者。Visibility が 'private' のときに要る
	//（作成者へ space_grant(admin) を張らないと、作った本人にも見えない）。
	CreatorUserID uint64
}

func (u *CreateSpaceUseCase) Execute(ctx context.Context, in CreateSpaceInput) (*domain.Space, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.Visibility == "" {
		in.Visibility = domain.SpaceVisibilityWorkspace
	}
	if !domain.ValidSpaceVisibility(in.Visibility) {
		return nil, ErrInvalidSpaceVisibility
	}
	if in.Visibility == domain.SpaceVisibilityPrivate && in.CreatorUserID == 0 {
		return nil, errors.New("creatorUserID is required for private space")
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
		space, err := u.createOnce(ctx, in)
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

// createOnce は 1 回分の作成。private は provisioner（grant とセット）、
// workspace は今までどおり repo の単発 INSERT。
func (u *CreateSpaceUseCase) createOnce(ctx context.Context, in CreateSpaceInput) (*domain.Space, error) {
	if in.Visibility == domain.SpaceVisibilityPrivate {
		return u.provisioner.ProvisionPrivateSpace(ctx, repository.PrivateSpaceProvisionInput{
			WorkspaceID:   in.WorkspaceID,
			Key:           in.Key,
			Name:          in.Name,
			CreatorUserID: in.CreatorUserID,
		})
	}
	space := &domain.Space{
		WorkspaceID: in.WorkspaceID,
		Key:         in.Key,
		Name:        in.Name,
		Visibility:  domain.SpaceVisibilityWorkspace,
	}
	if err := u.repo.CreateSpace(ctx, space); err != nil {
		return nil, err
	}
	return space, nil
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

// EnsurePersonalWorkspaceUseCase は、そのユーザーの個人ワークスペースが既にあれば返し、
// 無ければ作って返す。一意性は DB（uq_workspaces_personal_owner）が守るので、作成が
// repository.ErrPersonalWorkspaceAlreadyExists で競合したら引き直す（check-then-act はしない）。
type EnsurePersonalWorkspaceUseCase struct {
	workspaces  repository.KnowledgeBaseRepository
	provisioner repository.WorkspaceProvisioner
}

func NewEnsurePersonalWorkspaceUseCase(
	w repository.KnowledgeBaseRepository, p repository.WorkspaceProvisioner,
) *EnsurePersonalWorkspaceUseCase {
	return &EnsurePersonalWorkspaceUseCase{workspaces: w, provisioner: p}
}

type EnsurePersonalWorkspaceInput struct {
	UserID uint64
	// Name は新規作成のときだけ使う表示名（例: 利用者の氏名）。既に個人ワークスペースが
	// あるときは無視する（作成後に利用者が改名しているかもしれない値を上書きしない）。
	Name string
}

func (u *EnsurePersonalWorkspaceUseCase) Execute(
	ctx context.Context, in EnsurePersonalWorkspaceInput,
) (*domain.Workspace, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}

	// 大半のログイン（初回サインアップ以外）はここで終わる — 1 回の SELECT。
	if ws, err := u.workspaces.FindPersonalWorkspaceByOwner(ctx, in.UserID); err == nil {
		return ws, nil
	} else if !errors.Is(err, repository.ErrWorkspaceNotFound) {
		return nil, fmt.Errorf("find personal workspace: %w", err)
	}

	slug := generatedURLKey("w")
	ownerID := in.UserID
	for {
		created, err := u.provisioner.ProvisionWorkspace(ctx, repository.WorkspaceProvisionInput{
			Slug:                slug,
			Name:                in.Name,
			OwnerUserID:         in.UserID,
			PersonalOwnerUserID: &ownerID,
		})
		switch {
		case err == nil:
			return created, nil
		case errors.Is(err, repository.ErrWorkspaceSlugTaken):
			// 自動採番の衝突（48bit の乱数なので実際にはほぼ起きない）。引き直す。
			slug = generatedURLKey("w")
			continue
		case errors.Is(err, repository.ErrPersonalWorkspaceAlreadyExists):
			// この判定からこの INSERT までの間に、別のリクエスト（二重送信・同時実行）が
			// 先に作り終えていた。失敗として扱わず、その 1 つを引いて返す。
			ws, findErr := u.workspaces.FindPersonalWorkspaceByOwner(ctx, in.UserID)
			if findErr != nil {
				return nil, fmt.Errorf("find personal workspace after race: %w", findErr)
			}
			return ws, nil
		default:
			return nil, fmt.Errorf("provision personal workspace: %w", err)
		}
	}
}

// JoinCompanyWorkspaceUseCase は「その人の会社のワークスペース」へ自動で入れる。

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
