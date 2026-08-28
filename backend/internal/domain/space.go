package domain

import "time"

// Space はワークスペース内のページの束（部門・プロジェクト単位の入れ物）。
// Key はワークスペース内で一意な短い識別子で、URL とページの所属表示に使う。
type Space struct {
	ID string `json:"id"`
	// WorkspaceID はテナント境界。pages からの複合 FK の参照先にもなる。
	WorkspaceID string `json:"workspaceId"`
	// Key はワークスペース内で一意な短い識別子（例: "eng"）。
	Key string `json:"key"`
	// Name は表示名。
	Name string `json:"name"`
	// Visibility はワークスペース既定の grant が届くか。値は SpaceVisibility* が正。
	Visibility SpaceVisibility `json:"visibility"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// SpaceVisibility はスペースの見え方の区分。
//
// 'workspace' はワークスペース全体の grant（メンバー既定の役割）が届く通常のスペース。
// 'private' には**スペース単位の付与（space_grants）だけ**が届き、ワークスペース全体の
// grant も「そのスペースの全員（space_all）」宛ての grant も効かない — space_all まで
// 塞ぐのは、その 1 行で全メンバーに開いてしまい「プライベート」の意味が壊れるため。
// 「プライベートかどうか」を grant の構成から導出しない（viewer を 1 人足しただけで
// 区分が飛び移らない）ための明示の印。
type SpaceVisibility string

const (
	SpaceVisibilityWorkspace SpaceVisibility = "workspace"
	SpaceVisibilityPrivate   SpaceVisibility = "private"
)

// ValidSpaceVisibility は保存してよい visibility の値かを返す。
func ValidSpaceVisibility(v SpaceVisibility) bool {
	return v == SpaceVisibilityWorkspace || v == SpaceVisibilityPrivate
}

// SpaceKeyMaxLen / SpaceNameMaxLen は spaces の列幅（varchar(64) / varchar(200)）。
const (
	SpaceKeyMaxLen  = 64
	SpaceNameMaxLen = 200
)

// ValidSpaceKey はスペースの key として保存してよい形かを返す。
// 形は workspaces.slug と同じ（どちらも人が打つ短い識別子で、揺れを持ち込まない）。
func ValidSpaceKey(key string) bool {
	return validURLKey(key, SpaceKeyMaxLen)
}
