package database

import (
	"context"
	"database/sql"
	_ "embed"
)

// coreSchemaDDL はアプリケーション中核テーブルの DDL（実スキーマの正本）。
// バイナリに埋め込んで起動時に流すため、デプロイ物とスキーマ定義が必ず同じ版になる。
// 同じファイルが sqlc の型付け入力でもあるので、宣言と実体がずれない。
//
//go:embed schema/core.sql
var coreSchemaDDL string

// ApplyCoreSchema はアプリケーション中核テーブル（users / roles / courses / exercises …）の
// スキーマを適用する（冪等）。ノート（[ApplyKnowledgeBaseSchema]）より先に呼ぶこと
// （権限モデルが users を参照するため）。
//
// DDL は CREATE TABLE / CREATE INDEX の IF NOT EXISTS だけで冪等になっており、1 文ずつ
// 順に実行する。CREATE TABLE IF NOT EXISTS は既に在るテーブルへ列を足さないので、
// 既存 DB の列追加・型変更は migrations/000X_*.sql（明示 SQL）が担う。
// 既に在る索引の CREATE INDEX は発行しない（理由は [applyEmbeddedSchema]）。
// 失敗したらエラーを返して起動を止める（スキーマが半端なまま listen を始めない）。
func ApplyCoreSchema(ctx context.Context, db *sql.DB) error {
	return applyEmbeddedSchema(ctx, db, "中核スキーマ", coreSchemaDDL)
}
