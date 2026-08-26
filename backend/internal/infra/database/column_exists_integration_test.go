//go:build integration

package database

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

// TestColumnExists_SchemaScoped_Integration は columnExists が public スキーマだけを見ることを固定する。
//
// 本番の DB には Supabase 自身の auth スキーマが同居しており、そこにも users という表がある。
// スキーマを絞らないと、public.users から列を消した後も auth.users の同名列を拾って
// 「まだある」と誤判定し、列が無い前提の SQL を流して起動時マイグレーションが落ちる。
// 実際 auth.users には role 列があり、これは migrations/0021 が public.users から消す列と同名。
//
// columnExists は非公開なので同一パッケージに置く。testsupport は database を import するため、
// ここから使うと import 循環になる。そのため接続は自前で開く（スキーマ全体は要らず、
// 検証に必要な表だけを作る）。
func TestColumnExists_SchemaScoped_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL が未設定のためスキップ（DB を用意した実行では設定される）")
	}
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.Ping(), "結合テスト用 PostgreSQL に接続できません")

	ctx := context.Background()
	for _, stmt := range []string{
		`CREATE SCHEMA IF NOT EXISTS colcheck_other`,
		// 本番の auth.users を模して、別スキーマに同名の表と、public には無い列を作る。
		`CREATE TABLE IF NOT EXISTS colcheck_other.users (id bigint, only_in_other_schema text)`,
		`CREATE TABLE IF NOT EXISTS public.users (id bigint, email text)`,
	} {
		_, err := db.ExecContext(ctx, stmt)
		require.NoError(t, err, stmt)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS colcheck_other CASCADE`)
	})

	t.Run("別スキーマにしか無い列は無いと判定する", func(t *testing.T) {
		got, err := columnExists(ctx, db, "users", "only_in_other_schema")
		require.NoError(t, err)
		require.False(t, got, "public.users に無い列を、別スキーマの同名表から拾ってはいけない")
	})

	t.Run("public にある列は在ると判定する", func(t *testing.T) {
		got, err := columnExists(ctx, db, "users", "email")
		require.NoError(t, err)
		require.True(t, got)
	})

	t.Run("どこにも無い列は無いと判定する", func(t *testing.T) {
		got, err := columnExists(ctx, db, "users", "definitely_not_a_column")
		require.NoError(t, err)
		require.False(t, got)
	})
}
