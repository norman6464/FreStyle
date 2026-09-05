//go:build integration

// Package database_test は起動時に適用するスキーマ（database.ApplySchema）そのものを
// 本物の PostgreSQL に対して検証する。
//
// testsupport.OpenTestDB は ApplySchema を呼ぶだけで、テーブルの並びや制約の詳細までは
// 見ない。ここで実際に張られた表・制約・索引を突き合わせ、schema.hcl の変更が
// schema.gen.sql と食い違ったまま出荷される事故を防ぐ。
package database_test

import (
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestApplySchema_Integration はまっさらな DB へ ApplySchema を流し、期待する形になることを固定する。
func TestApplySchema_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := t.Context()
	resetPublicSchema(t, db)

	require.NoError(t, database.ApplySchema(ctx, db))

	t.Run("中核テーブルが揃っている", func(t *testing.T) {
		for _, table := range []string{
			"users", "user_oidc_identities",
			"master_exercises", "master_exercise_examples",
			"exercise_submissions", "notes",
			"notifications",
			"rich_documents",
		} {
			require.True(t, tableExists(t, db, table), "中核テーブル %s が無い", table)
		}
	})

	t.Run("roles マスタは作られない", func(t *testing.T) {
		// アプリ全体のロール（かつての users.role）は撤去済み。参照先マスタも作らない。
		require.False(t, tableExists(t, db, "roles"), "roles テーブルが残っている")
	})

	t.Run("退役済みのテナント移行期テーブルは作られない", func(t *testing.T) {
		// companies / company_applications / company_exercises はテナントの正本が
		// workspaces へ完全移行済みのレガシー（退役 PR で撤去）。
		for _, table := range []string{"companies", "company_applications", "company_exercises"} {
			require.False(t, tableExists(t, db, table), "退役済みのテーブル %s が残っている", table)
		}
	})

	t.Run("ノートと権限モデルが揃っている", func(t *testing.T) {
		for _, table := range []string{
			"workspaces", "spaces", "pages", "blocks", "page_paths", "page_snapshots",
			"principals", "principal_members", "workspace_grants", "space_grants",
			"page_grants", "share_links",
		} {
			require.True(t, tableExists(t, db, table), "ノートのテーブル %s が無い", table)
		}
	})

	t.Run("権限を打ち消す置き場は作られない", func(t *testing.T) {
		// 権限は 3 段の付与（workspace / space / page）を足し合わせ、届いた中で
		// 最も強い役割で決まる。下の段が上の段を弱める仕組みは持たないので、その
		// 置き場だったテーブルが DDL に戻っていないことを見る。
		for _, table := range []string{"page_restrictions", "page_allow_lists"} {
			require.False(t, tableExists(t, db, table),
				"使わないテーブル %s が作られている（狭める側の仕組みは持たない）", table)
		}
	})

	t.Run("循環参照の FK が張られている", func(t *testing.T) {
		// users ⇄ workspaces の循環依存（旧 DO ブロック）。Atlas は FK をまとめて末尾の
		// ALTER で張るため、宣言側は普通の foreign_key ブロックのままで済む。
		require.True(t, columnExists(t, db, "users", "workspace_id"))
		require.True(t, constraintExists(t, db, "users", "fk_users_workspace"))
	})

	t.Run("役割・識別子まわりの制約が張られている", func(t *testing.T) {
		require.True(t, constraintExists(t, db, "user_oidc_identities", "fk_user_oidc_identities_user"))
		require.True(t, constraintExists(t, db, "user_oidc_identities", "ck_user_oidc_identities_not_empty"))
		require.True(t, constraintExists(t, db, "rich_documents", "fk_rich_documents_owner"))
		require.True(t, constraintExists(t, db, "rich_documents", "ck_rich_documents_doc"))
		require.True(t, constraintExists(t, db, "rich_documents", "ck_rich_documents_title_len"))
		require.True(t, indexExists(t, db, "uq_users_email_active"))
	})
}

// TestApplySchema_二重呼び出しは何もしない_Integration は、同じ PostgreSQL を複数のテスト
// バイナリが順に共有する結合テストの実運用を固定する。schema.gen.sql は CREATE 文だけで
// IF NOT EXISTS を持たないため、素で 2 回 Exec すると必ず失敗する。ApplySchema は
// users テーブルの有無で「初めてか」を判定して 2 回目以降を静かに no-op にする
// （database.go の hasCoreSchema を参照）。
func TestApplySchema_二重呼び出しは何もしない_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t) // 内部で ApplySchema が 1 回済んでいる
	ctx := t.Context()

	require.NoError(t, database.ApplySchema(ctx, db), "2 回目の呼び出しは失敗してはいけない")
	require.True(t, tableExists(t, db, "users"))
}

// resetPublicSchema は public schema を作り直して、まっさらな DB を再現する。
// 結合テストは serializeIntegration で 1 テスト関数ずつ直列に走るので、他のテストと
// 衝突しない（このテストが最後にスキーマを作り直して返す）。
func resetPublicSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(t.Context(), `DROP SCHEMA public CASCADE`)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), `CREATE SCHEMA public`)
	require.NoError(t, err)
}

func tableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()
	var n int64
	require.NoError(t, db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM information_schema.tables
		  WHERE table_schema = current_schema() AND table_name = $1`, table).Scan(&n))
	return n > 0
}

func columnExists(t *testing.T, db *sql.DB, table, column string) bool {
	t.Helper()
	var n int64
	require.NoError(t, db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM information_schema.columns
		  WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2`,
		table, column).Scan(&n))
	return n > 0
}

func constraintExists(t *testing.T, db *sql.DB, table, name string) bool {
	t.Helper()
	var n int64
	require.NoError(t, db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM pg_constraint WHERE conname = $1 AND conrelid = $2::regclass`,
		name, table).Scan(&n))
	return n > 0
}

func indexExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int64
	require.NoError(t, db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM pg_indexes
		  WHERE schemaname = current_schema() AND indexname = $1`, name).Scan(&n))
	return n > 0
}
