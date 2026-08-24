package domain

import "time"

// Space はワークスペース内のページの束（部門・プロジェクト単位の入れ物）。
// Key はワークスペース内で一意な短い識別子で、URL とページの所属表示に使う。
type Space struct {
	ID string `gorm:"type:uuid;primaryKey" json:"id"`
	// WorkspaceID はテナント境界。pages からの複合 FK の参照先にもなる。
	WorkspaceID string `gorm:"column:workspace_id;type:uuid;not null;index" json:"workspaceId"`
	// Key はワークスペース内で一意な短い識別子（例: "eng"）。
	Key string `gorm:"column:key;type:varchar(64);not null" json:"key"`
	// Name は表示名。
	Name      string    `gorm:"column:name;type:varchar(200);not null" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName は GORM のテーブル名を固定する。
func (Space) TableName() string { return "spaces" }
