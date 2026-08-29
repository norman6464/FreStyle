package database

import (
	"context"
	"database/sql"
	_ "embed"
)

// knowledgeBaseSchemaDDL はノートの骨格 6 テーブルの DDL（実スキーマの正本）。
// バイナリに埋め込んで起動時に流すため、デプロイ物とスキーマ定義が必ず同じ版になる。
//
//go:embed schema/knowledge_base.sql
var knowledgeBaseSchemaDDL string

// knowledgeBasePermissionSchemaDDL は権限モデル 7 テーブルの DDL（実スキーマの正本）。
// 骨格の DDL より後に流す必要がある（spaces / pages を参照するため）。
//
//go:embed schema/knowledge_base_permissions.sql
var knowledgeBasePermissionSchemaDDL string

// ApplyKnowledgeBaseSchema はノート（workspaces / spaces / pages / blocks /
// page_paths / page_snapshots と、権限モデルの principals / principal_members /
// workspace_grants / space_grants / page_restrictions / page_allow_lists / share_links）の
// スキーマを適用する（冪等）。
//
// 2 つの DDL は 1 つのトランザクションで順に流す。権限側は骨格側の spaces / pages と
// 中核スキーマの users を参照するため、順序（中核 → 骨格 → 権限）を崩せない。
//
// DDL は CREATE ... IF NOT EXISTS だけで冪等になっており、1 文ずつ順に実行する。
// 既に在る索引の CREATE INDEX は発行しない（理由は [applyEmbeddedSchema]）。
// 失敗したらエラーを返して起動を止める（スキーマが半端なまま listen を始めない）。
func ApplyKnowledgeBaseSchema(ctx context.Context, db *sql.DB) error {
	return applyEmbeddedSchema(
		ctx, db, "ノートスキーマ",
		knowledgeBaseSchemaDDL, knowledgeBasePermissionSchemaDDL,
	)
}
