package domain

import "time"

// Workspace はナレッジ基盤のテナント境界。配下の space / page / block はすべて
// workspace_id を持ち、複合 FK で「別テナントの行を親にできない」ことを DB 側で保証する
// （スキーマの正本は infra/database/schema/knowledge_base.sql）。
//
// ナレッジ基盤の型は GORM を通さない（AutoMigrate の対象外・GORM タグを持たない）。
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
