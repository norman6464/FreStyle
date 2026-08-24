package domain

import "time"

// Workspace はナレッジ基盤のテナント境界。配下の space / page / block はすべて
// workspace_id を持ち、複合 FK で「別テナントの行を親にできない」ことを DB 側で保証する
// （制約は infra/database の ApplyKnowledgeBaseConstraints が張る）。
//
// ID は推測不能な UUID。採番は repository 層で UUIDv7 を振る（段 1-b で追加）。
// domain は標準ライブラリ + GORM tag のみに依存させるため、ここには採番ヘルパを置かない。
type Workspace struct {
	ID string `gorm:"type:uuid;primaryKey" json:"id"`
	// Slug は URL に出る短い識別子（テナント内ではなくグローバルに一意）。
	Slug string `gorm:"column:slug;type:varchar(64);not null" json:"slug"`
	// Name は表示名。
	Name      string    `gorm:"column:name;type:varchar(200);not null" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName は GORM のテーブル名を固定する。
func (Workspace) TableName() string { return "workspaces" }
