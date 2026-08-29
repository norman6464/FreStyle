package database

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSplitSQLStatements は DDL の文分割が、コメントや引用符の内側のセミコロンで
// 切らないことを固定する。ここを取り違えると起動時に壊れた SQL を流すことになる。
func TestSplitSQLStatements(t *testing.T) {
	tests := []struct {
		name string
		ddl  string
		want []string
	}{
		{
			name: "空文字は 0 文",
			ddl:  "",
			want: nil,
		},
		{
			name: "コメントだけの本文は 0 文",
			ddl:  "-- 説明\n/* まとめ */\n",
			want: nil,
		},
		{
			name: "末尾のセミコロンが無い文も拾う",
			ddl:  "CREATE TABLE a (id int)",
			want: []string{"CREATE TABLE a (id int)"},
		},
		{
			name: "文の直前のコメントは文にぶら下げる",
			ddl:  "-- 表 a\nCREATE TABLE a (id int);\nCREATE TABLE b (id int);",
			want: []string{"-- 表 a\nCREATE TABLE a (id int)", "CREATE TABLE b (id int)"},
		},
		{
			name: "行コメント内のセミコロンでは切らない",
			ddl:  "CREATE TABLE a ( -- ; ここでは切らない\n  id int\n);",
			want: []string{"CREATE TABLE a ( -- ; ここでは切らない\n  id int\n)"},
		},
		{
			name: "入れ子のブロックコメント内のセミコロンでは切らない",
			ddl:  "/* 外 /* 内 ; */ ; */ CREATE TABLE a (id int);",
			want: []string{"/* 外 /* 内 ; */ ; */ CREATE TABLE a (id int)"},
		},
		{
			name: "文字列リテラル内のセミコロンでは切らない",
			ddl:  "CREATE TABLE a (s text DEFAULT ';');",
			want: []string{"CREATE TABLE a (s text DEFAULT ';')"},
		},
		{
			name: "二重化したクォートを閉じ引用と誤認しない",
			ddl:  "CREATE TABLE a (s text DEFAULT 'it''s ; ok');",
			want: []string{"CREATE TABLE a (s text DEFAULT 'it''s ; ok')"},
		},
		{
			name: "引用付き識別子内のセミコロンでは切らない",
			ddl:  `CREATE TABLE "a;b" (id int);`,
			want: []string{`CREATE TABLE "a;b" (id int)`},
		},
		{
			name: "dollar quote の中身では切らない",
			ddl:  "DO $$ BEGIN CREATE INDEX i ON a (id); END $$;\nCREATE TABLE b (id int);",
			want: []string{"DO $$ BEGIN CREATE INDEX i ON a (id); END $$", "CREATE TABLE b (id int)"},
		},
		{
			name: "タグ付き dollar quote の中身でも切らない",
			ddl:  "DO $tag$ SELECT 1; $tag$;",
			want: []string{"DO $tag$ SELECT 1; $tag$"},
		},
		{
			name: "位置パラメータの $1 は dollar quote ではない",
			ddl:  "SELECT pg_advisory_xact_lock($1); SELECT 2;",
			want: []string{"SELECT pg_advisory_xact_lock($1)", "SELECT 2"},
		},
		{
			name: "空の文は捨てる",
			ddl:  "CREATE TABLE a (id int);;\n;\n",
			want: []string{"CREATE TABLE a (id int)"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, splitSQLStatements(tt.ddl))
		})
	}
}

// TestCreateIndexName は CREATE INDEX 文だけを索引名付きで見分けられることを固定する。
// ここで取りこぼすと、既存の索引に対しても CREATE INDEX を流してしまい
// テーブルの ShareLock を握る（見分けすぎると必要な索引を作り損ねる）。
func TestCreateIndexName(t *testing.T) {
	tests := []struct {
		name string
		stmt string
		want string
	}{
		{"通常の索引", "CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes (user_id)", "idx_notes_user_id"},
		{"一意索引", "CREATE UNIQUE INDEX IF NOT EXISTS uq_a ON a (b)", "uq_a"},
		{"IF NOT EXISTS 無し", "CREATE INDEX idx_a ON a (b)", "idx_a"},
		{"CONCURRENTLY 付き", "CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_a ON a (b)", "uq_a"},
		{"大文字小文字と改行が混ざっても拾う", "create\n  Unique  Index\n  If Not Exists  Idx_A ON a (b)", "idx_a"},
		{"引用付きの名前は畳まない", `CREATE INDEX IF NOT EXISTS "Idx_A" ON a (b)`, "Idx_A"},
		{"直前のコメントを跨いで拾う", "-- 索引\n/* 二重 */\nCREATE INDEX idx_a ON a (b)", "idx_a"},
		{"CREATE TABLE は対象外", "CREATE TABLE IF NOT EXISTS a (id int)", ""},
		{"ALTER TABLE は対象外", "ALTER TABLE a ADD COLUMN b int", ""},
		{"DO ブロックの中の CREATE INDEX は対象外", "DO $$ BEGIN CREATE INDEX idx_a ON a (b); END $$", ""},
		{"文字列に CREATE INDEX が出てきても対象外", "INSERT INTO t VALUES ('CREATE INDEX idx_a ON a (b)')", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, createIndexName(tt.stmt))
		})
	}
}

// TestSplitSQLStatements_埋め込みスキーマ は正本の DDL がそのまま分割できること
// （分割結果に空文が無く、索引名を持つ文が実際に取り出せること）を確かめる。
func TestSplitSQLStatements_埋め込みスキーマ(t *testing.T) {
	core, err := coreSchemaSection()
	require.NoError(t, err)
	note, err := noteSchemaSection()
	require.NoError(t, err)

	for _, ddl := range []struct {
		name string
		body string
	}{
		// 1 ファイルだが適用は 2 回に分かれるので、切り出した節ごとに確かめる。
		{"中核", core},
		{"ノート", note},
		{"全体", schemaDDL},
	} {
		t.Run(ddl.name, func(t *testing.T) {
			stmts := splitSQLStatements(ddl.body)
			require.NotEmpty(t, stmts)

			indexes := 0
			for _, stmt := range stmts {
				require.NotEmpty(t, stripLeadingComments(stmt), "コメントだけの断片が文として残っている")
				if createIndexName(stmt) != "" {
					indexes++
				}
			}
			require.NotZero(t, indexes, "CREATE INDEX を 1 本も認識できていない")
		})
	}
}

// TestSummarizeStatement はエラーメッセージ用の要約が 1 行に収まることを固定する。
func TestSummarizeStatement(t *testing.T) {
	require.Equal(t, "CREATE TABLE a (", summarizeStatement("-- 説明\nCREATE TABLE a (\n  id int\n)"))
	require.Equal(t, "", summarizeStatement("-- 説明だけ"))

	long := "CREATE INDEX IF NOT EXISTS idx_" + strings.Repeat("x", 200) + " ON a (b)"
	require.Len(t, []rune(summarizeStatement(long)), 81) // 80 文字 + 省略記号
}
