package repository

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrPrincipalNotFound は対象の主体が存在しない（または別ワークスペースのもの）ときに返す。
var ErrPrincipalNotFound = errors.New("principal not found")

// ErrShareLinkNotFound は対象の共有リンクが存在しないときに返す。
// トークンが違う場合もこれを返す（存在の有無自体を漏らさない）。
var ErrShareLinkNotFound = errors.New("share link not found")

// ErrUserNotFound は主体を作ろうとしたユーザーが users に存在しないときに返す。
//
// principals.user_id は users への FK なので、実在しないユーザー ID で
// EnsureUserPrincipal を呼ぶと制約違反になる。それをそのまま上へ流すと
// 「ユーザー ID を間違えた」という入力の誤りが 500 になり、呼び出し側は
// DB 障害と区別できない（再試行すべきだと誤解する）。
var ErrUserNotFound = errors.New("user not found")

// ErrPrincipalGroupNameTaken はグループ名が同じワークスペースで使用済みのときに返す。
// 名前はワークスペース内で一意（uq_principals_group_name）で、同名が 2 つあると
// 権限を張る先を人が選べなくなる。
var ErrPrincipalGroupNameTaken = errors.New("principal group name is already taken")

// PageWithViewFacts は 1 ページと、そのページを閲覧できるかを決める事実の組。
// ListSpacePageViewFacts が返す（ふるい落としは domain.ResolvePageView が行う）。
//
// 事実が閲覧専用の型なのは、一覧のクエリが閲覧の列しか集めないため。
// 1 ページ解決と同じ domain.PagePermissionFacts に載せると、編集の例外を 1 つも
// 見ないまま CanEdit が既定で返る（domain.PageViewFacts の説明を参照）。
type PageWithViewFacts struct {
	Page  domain.Page
	Facts domain.PageViewFacts
}

// PageWithPermissionFacts は 1 ページの ID と、その実効権限を決める事実の組。
// ListSubtreePagePermissionFacts が返す（判定は domain.ResolvePagePermission が行う）。
//
// 閲覧専用の PageWithViewFacts と分かれているのは、集めた事実が違うため。
// 編集可否は閲覧可否を含むので、編集を問う経路は閲覧と編集の両方の事実を集める。
// ページ本体を持たないのは、この型を使う経路（サブツリー一括操作の入口検査）が
// 可否だけを必要とし、見えないページの中身を呼び出し側へ渡す必要が無いため。
type PageWithPermissionFacts struct {
	PageID string
	Facts  domain.PagePermissionFacts
}

// ShareLinkWrite は共有リンクの発行に渡す値。
//
// ID と PrincipalID を持たないのは、どちらも採番が repository の責務のため
// （主体の作成と共有リンクの作成は同じトランザクションで行う）。
type ShareLinkWrite struct {
	WorkspaceID string
	PageID      string
	// Capability はリンク経由でできることの既定。
	Capability domain.Capability
	// TokenHash は共有 URL に載るトークンの SHA-256（32 バイト）。平文は渡さない。
	TokenHash []byte
	// PasswordHash はパスワードの bcrypt ハッシュ。nil ならパスワード無し。
	PasswordHash *string
	// ExpiresAt は有効期限。nil なら無期限。
	ExpiresAt *time.Time
	// CreatedByUserID は発行者。
	CreatedByUserID uint64
}

// KnowledgeBasePermissionRepository はナレッジ基盤の権限モデル（principals /
// principal_members / workspace_grants / space_grants / page_restrictions / share_links）への
// アクセスを提供する。
//
// KnowledgeBaseRepository（ページとブロック）と分けているのは、境界が違うため。
// あちらをひとつの fat interface にまとめている理由は「ページ作成 = pages + page_paths」
// のように複数テーブルを 1 トランザクションで書く操作が中心だからで、権限の書き込みは
// それらと同じトランザクションに入らない（権限を張る操作とページを書く操作は別の要求）。
// 同じ interface に足すと、ページだけを扱う実装や fake が権限のメソッドまで
// 実装しなければならなくなり、境界の意味も薄れる。
//
// 読み取りは pages / page_paths をまたぐ（実効権限の解決に closure が要る）が、
// 境界を決めるのは書き込みのトランザクション単位なので問題にしない。
type KnowledgeBasePermissionRepository interface {
	// EnsureUserPrincipal はユーザーの主体を作る（既にあればそれを返す）。
	// この行があること自体がワークスペース所属を意味する。
	EnsureUserPrincipal(ctx context.Context, workspaceID string, userID uint64) (*domain.Principal, error)
	// EnsureSpaceEveryonePrincipal はスペースの「全員」を表す主体を作る（既にあればそれを返す）。
	EnsureSpaceEveryonePrincipal(ctx context.Context, workspaceID, spaceID string) (*domain.Principal, error)
	// CreateGroupPrincipal はグループの主体を作る。名前はワークスペース内で一意。
	CreateGroupPrincipal(ctx context.Context, workspaceID, name string) (*domain.Principal, error)
	// FindPrincipal は主体を 1 件引く。無い・別ワークスペースなら ErrPrincipalNotFound。
	FindPrincipal(ctx context.Context, workspaceID, principalID string) (*domain.Principal, error)
	// FindUserPrincipal はユーザーの主体を引く。無ければ ErrPrincipalNotFound（= 非メンバー）。
	FindUserPrincipal(ctx context.Context, workspaceID string, userID uint64) (*domain.Principal, error)
	// DeletePrincipal は主体を消す。紐づく grant / restriction / グループ所属も
	// FK の CASCADE で消える。対象が無ければ ErrPrincipalNotFound。
	//
	// 許可リストに載っていた主体でも、その段が許可リスト制であること自体は消えない
	// （印は page_allow_lists が持ち、主体を参照しない）。載っていた人が居なくなった段は
	// 「誰も載っていない許可リスト」＝ 誰にも見えない状態になる。閉じる側へ倒すのは、
	// オフボーディング 1 回で祖先の限定公開が第三者に開くのを避けるため。
	DeletePrincipal(ctx context.Context, workspaceID, principalID string) error
	// IsWorkspaceMember はユーザーがワークスペースのメンバーかを返す。
	IsWorkspaceMember(ctx context.Context, workspaceID string, userID uint64) (bool, error)
	// ListMemberWorkspaces はそのユーザーが所属するワークスペースを返す（slug 順）。
	// 所属は principals（kind='user'）の行が唯一の表現なので、その JOIN がそのまま答えになる。
	// ナレッジ基盤で唯一テナントを跨いで読むメソッド（どのテナントに入れるかを答える口）で、
	// 絞り込みは user_id だけが行う。
	ListMemberWorkspaces(ctx context.Context, userID uint64) ([]domain.Workspace, error)

	// AddGroupMember はグループに主体を所属させる（冪等）。member 側は kind='user' でなければ
	// DB の複合 FK が弾く（グループの入れ子を作らせない）。
	AddGroupMember(ctx context.Context, workspaceID, groupPrincipalID, memberPrincipalID string) error
	// RemoveGroupMember はグループから主体を外す。存在しなければ何もしない（冪等）。
	RemoveGroupMember(ctx context.Context, workspaceID, groupPrincipalID, memberPrincipalID string) error

	// UpsertWorkspaceGrant はワークスペース全体での既定の役割を与える（同じ主体には 1 行だけ）。
	UpsertWorkspaceGrant(ctx context.Context, workspaceID, principalID string, role domain.GrantRole) (*domain.WorkspaceGrant, error)
	// DeleteWorkspaceGrant はワークスペース全体での既定の役割を剥がす（冪等）。
	DeleteWorkspaceGrant(ctx context.Context, workspaceID, principalID string) error
	// ListWorkspaceGrants はワークスペースの grant 一覧を返す。
	ListWorkspaceGrants(ctx context.Context, workspaceID string) ([]domain.WorkspaceGrant, error)

	// UpsertSpaceGrant はスペースでの既定の役割を与える（同じ主体には 1 行だけ）。
	UpsertSpaceGrant(ctx context.Context, workspaceID, spaceID, principalID string, role domain.GrantRole) (*domain.SpaceGrant, error)
	// DeleteSpaceGrant はスペースでの既定の役割を剥がす（冪等）。
	DeleteSpaceGrant(ctx context.Context, workspaceID, spaceID, principalID string) error
	// ListSpaceGrants はスペースの grant 一覧を返す。
	ListSpaceGrants(ctx context.Context, workspaceID, spaceID string) ([]domain.SpaceGrant, error)

	// UpsertPageRestriction はページの例外を設定する（同じ (ページ, 主体, ケイパビリティ) は 1 行）。
	// allow を張ると、そのページのそのケイパビリティは許可リスト制になる（印も同じ
	// トランザクションで立つ）。逆に既存の allow 行を deny へ書き換えたときは、その段に
	// allow 行が 1 行も残らなければ印も同じトランザクションで畳む。つまり印を畳むのは
	// DeletePageRestriction だけではない。
	UpsertPageRestriction(ctx context.Context, workspaceID, pageID, principalID string, capability domain.Capability, mode domain.RestrictionMode) (*domain.PageRestriction, error)
	// DeletePageRestriction はページの例外を解除する（冪等）。消したのが最後の allow 行なら
	// 許可リスト制も畳み、解決はより遠い祖先 → grant の既定へ戻る。
	// deny 行の解除では許可リスト制を畳まない（無関係な 1 行で限定公開が解けないように）。
	DeletePageRestriction(ctx context.Context, workspaceID, pageID, principalID string, capability domain.Capability) error
	// ListPageRestrictions はそのページ自身に張られた例外の一覧を返す（継承分は含まない）。
	ListPageRestrictions(ctx context.Context, workspaceID, pageID string) ([]domain.PageRestriction, error)
	// ListPageAllowListCapabilities はそのページ自身が許可リスト制になっている
	// ケイパビリティを返す。載っていた主体が消えて allow 行が 0 行になった段は
	// ListPageRestrictions には現れないため、権限設定を見せるときは両方を読む
	// （でないと「制限なし」に見えるページが実際には誰にも見えない、という説明できない
	// 食い違いになる）。
	ListPageAllowListCapabilities(ctx context.Context, workspaceID, pageID string) ([]domain.Capability, error)

	// CreateShareLink は共有リンクを発行する。kind='share_link' の主体の採番と作成も
	// 同じトランザクションで行う（主体だけが残る／リンクだけが残る状態を作らない）。
	CreateShareLink(ctx context.Context, in ShareLinkWrite) (*domain.ShareLink, error)
	// RevokeShareLink は共有リンクを失効させる。既に失効済みなら何もしない（冪等）。
	// 対象が無い・別ワークスペースなら ErrShareLinkNotFound。
	RevokeShareLink(ctx context.Context, workspaceID, shareLinkID string) error
	// FindShareLinkByTokenHash はトークンの SHA-256 から共有リンクを引く。
	// 期限切れ・失効も含めて返す（判定は usecase 側）。無ければ ErrShareLinkNotFound。
	FindShareLinkByTokenHash(ctx context.Context, tokenHash []byte) (*domain.ShareLink, error)
	// ListPageShareLinks はページに発行された共有リンクの一覧を返す（失効済みも含む）。
	ListPageShareLinks(ctx context.Context, workspaceID, pageID string) ([]domain.ShareLink, error)

	// PagePermissionFactsForUser はログイン済みユーザーとして、1 ページの実効権限を決める
	// 事実を 1 回のクエリで集める。判定は domain.ResolvePagePermission が行う。
	// ページが無い・別ワークスペースなら ErrPageNotFound。
	PagePermissionFactsForUser(ctx context.Context, workspaceID, pageID string, userID uint64) (*domain.PagePermissionFacts, error)
	// PagePermissionFactsForPrincipal は共有リンクの来訪者（kind='share_link' の主体）として
	// 同じ事実を集める。既定（リンクの capability）は呼び出し側が facts に載せる。
	PagePermissionFactsForPrincipal(ctx context.Context, workspaceID, pageID, principalID string) (*domain.PagePermissionFacts, error)
	// ListSpacePageViewFacts はスペース配下の現役ページ全件と、その閲覧の事実を
	// 1 回のクエリで返す（ページごとに問い合わせない）。編集の事実は集めないので、
	// 編集可否をここから出さないこと（返す型がそれを表している）。
	ListSpacePageViewFacts(ctx context.Context, workspaceID, spaceID string, userID uint64) ([]PageWithViewFacts, error)
	// SpacePermissionFactsForUser はページを介さず、スペース 1 つの実効権限を決める事実を集める。
	// 判定は domain.ResolveScopePermission が行う。スペースが無い・別ワークスペースなら
	// ErrSpaceNotFound。
	//
	// 返すのは「その入れ物に届いている役割の集合」だけで、どれを採るかの規則は持たない。
	// ページ単位の例外（page_restrictions）も見ない — スペースには例外の層が無いため。
	// したがってこの口の答えを「そのスペースのあるページを編集してよいか」に使ってはいけない
	// （ページに張られた deny を見ておらず、必ず緩い側へ倒れる）。使ってよいのは
	// 対象がまだ存在しない操作（スペース直下へのページ作成）だけ。
	//
	// スペースの実在をここで確かめるのは、確かめないと fail-open になるため。
	// workspace_grants は配下の全スペースに届くので、別ワークスペースのスペース ID を
	// 渡されても「自分のワークスペースでの役割」がそのまま返り、他テナントのスペースに対して
	// editor と答えてしまう。
	SpacePermissionFactsForUser(ctx context.Context, workspaceID, spaceID string, userID uint64) (*domain.ScopeFacts, error)
	// WorkspacePermissionFactsForUser はワークスペースそのものに対する実効権限を決める事実を集める。
	// スペースを作る操作のように、どのスペースにも属さない判定に使う。
	//
	// 実在を確かめないのは、ワークスペースが無ければ grant も 1 行も無く、
	// 役割 0 個 ＝ 何もできない（fail-closed）に自然と倒れるため。
	// 呼び出し側は middleware が slug から解決したワークスペースを渡す。
	WorkspacePermissionFactsForUser(ctx context.Context, workspaceID string, userID uint64) (*domain.ScopeFacts, error)
	// ListSubtreePagePermissionFacts はサブツリー（対象ページ自身 + 全子孫）の各ページと、
	// その実効権限を決める事実を 1 回のクエリで返す。アーカイブ済みのページも含む。
	// ページとその子孫をまとめて書き換える操作が、根 1 枚の権限だけで通らないようにするための口。
	// ページが無い・別ワークスペースなら空スライス（呼び出し側が先に根の権限を確かめている）。
	ListSubtreePagePermissionFacts(ctx context.Context, workspaceID, pageID string, userID uint64) ([]PageWithPermissionFacts, error)
}
