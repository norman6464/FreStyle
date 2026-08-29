package database

import (
	"context"
	"database/sql"
)

// ApplyCoreSchema はアプリケーション中核テーブル（users / roles / courses / exercises …）の
// スキーマを適用する（冪等）。ノート（[ApplyKnowledgeBaseSchema]）より先に呼ぶこと
// （権限モデルが users を参照するため）。
//
// DDL は CREATE TABLE / CREATE INDEX の IF NOT EXISTS だけで冪等になっており、1 文ずつ
// 順に実行する。CREATE TABLE IF NOT EXISTS は既に在るテーブルへ列を足さないので、
// 既存 DB への列追加は、カタログを見て足りないときだけ ALTER する DO ブロックで書く
// （schema.sql の冒頭に理由を書いてある）。
// 既に在る索引の CREATE INDEX は発行しない（理由は [applyEmbeddedSchema]）。
// 失敗したらエラーを返して起動を止める（スキーマが半端なまま listen を始めない）。
func ApplyCoreSchema(ctx context.Context, db *sql.DB) error {
	ddl, err := coreSchemaSection()
	if err != nil {
		return err
	}
	return applyEmbeddedSchema(ctx, db, "中核スキーマ", ddl)
}
