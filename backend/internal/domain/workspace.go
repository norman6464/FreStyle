package domain

import "time"

// Workspace はノートのテナント境界。配下の space / page / block はすべて workspace_id を持ち、
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
	Name      string    `json:"name"`
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
