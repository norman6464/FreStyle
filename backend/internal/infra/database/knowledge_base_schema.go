package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

// knowledgeBaseSchemaDDL はナレッジ基盤の骨格 6 テーブルの DDL（実スキーマの正本）。
// バイナリに埋め込んで起動時に流すため、デプロイ物とスキーマ定義が必ず同じ版になる。
//
//go:embed schema/knowledge_base.sql
var knowledgeBaseSchemaDDL string

// knowledgeBasePermissionSchemaDDL は権限モデル 6 テーブルの DDL（実スキーマの正本）。
// 骨格の DDL より後に流す必要がある（spaces / pages を参照するため）。
//
//go:embed schema/knowledge_base_permissions.sql
var knowledgeBasePermissionSchemaDDL string

// ApplyKnowledgeBaseSchema はナレッジ基盤（workspaces / spaces / pages / blocks /
// page_paths / page_snapshots と、権限モデルの principals / principal_members /
// workspace_grants / space_grants / page_restrictions / share_links）のスキーマを適用する（冪等）。
//
// このテーブル群は GORM を一切通さない。AutoMigrate は複合 FK / CHECK / 部分 UNIQUE /
// コレーション指定を表現できず、構造体タグと明示 SQL に定義が二重化するため、
// schema/knowledge_base*.sql を唯一の正本として素の *sql.DB に流す。
// 接続プールは GORM と共有する（db.DB() で取り出したものを渡す）。
//
// 2 つの DDL は 1 つのトランザクションで順に流す。権限側は骨格側の spaces / pages と
// AutoMigrate が作る users を参照するため、順序（AutoMigrate → 骨格 → 権限）を崩せない。
//
// DDL は CREATE ... IF NOT EXISTS だけで冪等になっており、複数文をまとめて 1 度に実行する。
// 失敗したらエラーを返して起動を止める（スキーマが半端なまま listen を始めない）。
func ApplyKnowledgeBaseSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ナレッジ基盤スキーマ: トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op

	// pgbouncer（transaction pooler）前提のため、セッションロックではなく
	// トランザクションロック（コミットで自動解放）を使う。
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, migrateAdvisoryLockKey); err != nil {
		return fmt.Errorf("ナレッジ基盤スキーマ: advisory lock の取得に失敗: %w", err)
	}
	if _, err := tx.ExecContext(ctx, knowledgeBaseSchemaDDL); err != nil {
		return fmt.Errorf("ナレッジ基盤スキーマ: DDL の適用に失敗: %w", err)
	}
	if _, err := tx.ExecContext(ctx, knowledgeBasePermissionSchemaDDL); err != nil {
		return fmt.Errorf("ナレッジ基盤スキーマ: 権限モデル DDL の適用に失敗: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ナレッジ基盤スキーマ: コミットに失敗: %w", err)
	}
	return nil
}
