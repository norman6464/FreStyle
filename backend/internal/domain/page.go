package domain

import "time"

// Page はノートの 1 ページ。ページ同士は ParentID で木構造をなす（無限入れ子）。
//
// 兄弟の並び順は整数の連番ではなく分数インデックス（internal/pkg/fracindex が採番する文字列キー）で
// 持つ。1 行動かすたびに後続を振り直す UPDATE を避けるため。順序の比較は「同じ親の中」でのみ意味を持つ。
// DB 側は position 列を COLLATE "C" に固定し、Go のバイト比較と ORDER BY を一致させる。
type Page struct {
	ID string `json:"id"`
	// WorkspaceID はテナント境界。space / 親ページとの複合 FK に使い、テナント越えの親子を DB が弾く。
	WorkspaceID string `json:"workspaceId"`
	// SpaceID は所属スペース。(workspace_id, space_id) の複合 FK で spaces を参照する。
	SpaceID string `json:"spaceId"`
	// ParentID は親ページ。NULL はスペース直下（ルート）を意味する。
	ParentID *string `json:"parentId,omitempty"`
	// Position は兄弟内の並び順を表す分数インデックス（fracindex.Between で採番する）。
	Position string `json:"position"`
	// Title は一覧・検索・パンくずに使うページ名。
	Title string `json:"title"`
	// CreatedByUserID は作成者（users.id）。
	CreatedByUserID uint64 `json:"createdByUserId"`
	// ArchivedAt はアーカイブ日時。NULL が現役。物理削除ではなくアーカイブで隠す運用のため、
	// 一意制約（同じ親の中で position が重複しない）はアーカイブ済みを除外した部分ユニークで張る。
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
