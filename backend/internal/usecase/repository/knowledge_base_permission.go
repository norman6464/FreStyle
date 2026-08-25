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

// PageViewFacts は 1 ページと、そのページを閲覧できるかを決める事実の組。
// ListSpacePageViewFacts が返す（ふるい落としは domain.ResolvePagePermission が行う）。
type PageViewFacts struct {
	Page  domain.Page
	Facts domain.PagePermissionFacts
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
	DeletePrincipal(ctx context.Context, workspaceID, principalID string) error
	// IsWorkspaceMember はユーザーがワークスペースのメンバーかを返す。
	IsWorkspaceMember(ctx context.Context, workspaceID string, userID uint64) (bool, error)

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
	UpsertPageRestriction(ctx context.Context, workspaceID, pageID, principalID string, capability domain.Capability, mode domain.RestrictionMode) (*domain.PageRestriction, error)
	// DeletePageRestriction はページの例外を解除する（冪等）。最後の 1 行が消えると
	// その段の制限が無くなり、解決はより遠い祖先 → grant の既定へ戻る。
	DeletePageRestriction(ctx context.Context, workspaceID, pageID, principalID string, capability domain.Capability) error
	// ListPageRestrictions はそのページ自身に張られた例外の一覧を返す（継承分は含まない）。
	ListPageRestrictions(ctx context.Context, workspaceID, pageID string) ([]domain.PageRestriction, error)

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
	// 1 回のクエリで返す（ページごとに問い合わせない）。
	ListSpacePageViewFacts(ctx context.Context, workspaceID, spaceID string, userID uint64) ([]PageViewFacts, error)
}
