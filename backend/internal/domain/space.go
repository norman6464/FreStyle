package domain

import "time"

// Space はワークスペース内のページの束（部門・プロジェクト単位の入れ物）。
// Key はワークスペース内で一意な短い識別子で、URL とページの所属表示に使う。
//
// Workspace と同じくナレッジ基盤の型なので GORM を通さない（段 1-b で repository が付くまで参照元は無い）。
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
