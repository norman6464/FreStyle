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
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
