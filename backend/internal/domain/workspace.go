package domain

import "time"

// Workspace はナレッジのテナント境界。配下の space / page / block はすべて workspace_id を持ち、
// 複合 FK で「別テナントの行を親にできない」ことを DB 側で保証する。
//
// スキーマの正本は infra/database/schema/knowledge_base.sql。
// 永続化は sqlc 生成コードから詰め替える。段 1-b で repository が付くまで参照元は無い。
//
// ID は推測不能な UUID。採番は repository 層で UUIDv7 を振る（段 1-b で追加）。
type Workspace struct {
	ID string `json:"id"`
	// Slug は URL に出る短い識別子（テナント内ではなくグローバルに一意）。
	Slug string `json:"slug"`
	// Name は表示名。
	Name string `json:"name"`
	// IsActive はテナントが利用可能か。false にすると、このワークスペースに所属する
	// 全員が API を使えなくなる（middleware が入口で弾く）。停止の唯一の表現。
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MemberWorkspace はワークスペースと、そのユーザーから見た実効権限の組。
// 一覧 API は削除アイコンの出し分けに要る CanManage だけを添えて返す
// （フロントは canManage を見ずに全員へ削除操作を出しており、押しても 403 になる問題があった）。
type MemberWorkspace struct {
	Workspace
	// CanManage は自分がこのワークスペースの admin か（DeleteWorkspace が要求する権限と同じ）。
	CanManage bool `json:"canManage"`
}

// WorkspaceMember はワークスペースに属する人 1 人。
//
// GrantablePrincipal とは用途が別で、こちらは**人だけ**を返す。あちらは権限を張る相手なので
// グループやスペース全員のような人でない主体も含み、閲覧にページの管理権限を要求する。
// 担当の表示名を出すことと、発言で人を名指すことは、権限を変えられない人にも要るので
// 所属していれば読める口が別に要る。
//
// PrincipalID と UserID の両方を持つのは、指す先が用途で違うため。担当は主体（principals）に
// 割り当てるので PrincipalID、発言中の名指しは users を指すので UserID を使う。
type WorkspaceMember struct {
	PrincipalID string `json:"principalId"`
	UserID      uint64 `json:"userId"`
	// Name は表示名。空文字のことがある（登録時に名前を持たない発行者があるため）。
	Name string `json:"name"`
	// AvatarURL / StatusMessage は profiles 由来（段 5）。設定していなければ空文字。
	AvatarURL     string `json:"avatarUrl"`
	StatusMessage string `json:"status"`
}

// AdminWorkspaceMember はメンバー管理画面（段 7）向けの 1 人。WorkspaceMember と違い、
// 停止中のアカウントも含み（復帰させる操作の対象になるため）、現在のワークスペース全体の
// 役割も一緒に返す。
type AdminWorkspaceMember struct {
	PrincipalID string `json:"principalId"`
	UserID      uint64 `json:"userId"`
	Name        string `json:"name"`
	// AccountStatus はアカウント全体の状態（active / suspended）。ここに deactivated は
	// 現れない — 退会は全ワークスペースを退出してから起きるため、この一覧に載る時点で
	// 対象外になっている。
	AccountStatus UserStatus `json:"accountStatus"`
	AvatarURL     string     `json:"avatarUrl"`
	StatusMessage string     `json:"statusMessage"`
	// Role はワークスペース全体の既定役割。nil は「ワークスペース全体には役割を持たない
	// （スペース/ページ単位の grant だけで見えている）」ことを表す。
	Role *GrantRole `json:"role,omitempty"`
}

// SpaceMember はスペースに届いている権限を人に解決した 1 行（段 9）。同じ人が複数経路
// （本人・所属グループ・スペース全員・ワークスペース全体）から役割を得ることがあり、
// 採用するのは最も強い役割（GrantRole.Rank 参照。ここでも弱める規則は無い）。
type SpaceMember struct {
	UserID    uint64    `json:"userId"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatarUrl"`
	Role      GrantRole `json:"role"`
	// Via はその役割がどの経路で届いたか。
	//   "direct"    本人への直接付与
	//   "group"     所属グループ、またはスペース全員（space_all）経由
	//   "workspace" ワークスペース全体の grant からの継承
	// 同じ人が複数の経路で同じ強さの役割を得ている場合は direct > group > workspace の
	// 優先度で選ぶ（最も具体的な理由を示すため）。
	Via string `json:"via"`
}

// WorkspaceSlugMaxLen / WorkspaceNameMaxLen は workspaces の列幅（varchar(64) / varchar(200)）。
// DB の CHECK / 列幅と同じ値を入口でも見て、桁あふれを 500 ではなく 400 で返せるようにする。
const (
	WorkspaceSlugMaxLen = 64
	WorkspaceNameMaxLen = 200
)

// ValidWorkspaceSlug は URL に出せる slug かを返す。
//
// 小文字英数字とハイフンだけに絞り、先頭と末尾は英数字に限る。大文字や記号を許すと
// 「同じに見えて別のワークスペース」（Acme と acme）が作れてしまい、URL を見ても
// どちらのテナントか判断できなくなる。長さの上限は DB の CHECK と同じ。
func ValidWorkspaceSlug(slug string) bool {
	return validURLKey(slug, WorkspaceSlugMaxLen)
}

// validURLKey は URL に出る識別子（workspaces.slug / spaces.key）の共通の形。
// 空でなく、[a-z0-9-] だけからなり、先頭・末尾がハイフンでないこと。
func validURLKey(s string, maxLen int) bool {
	if s == "" || len(s) > maxLen {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '-':
			// 先頭・末尾のハイフンは見た目の差が分かりにくく、区切りとして意味を持たない。
			if i == 0 || i == len(s)-1 {
				return false
			}
		default:
			return false
		}
	}
	return true
}
