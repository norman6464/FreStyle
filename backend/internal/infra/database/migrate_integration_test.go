//go:build integration

// Package database_test は起動時マイグレーション（database.Migrate）そのものを
// 本物の PostgreSQL に対して検証する。
//
// testsupport.OpenTestDB は Migrate を呼ばず、Apply*Schema / Seed* を自前の順序で並べ直して
// いる。そちらだけを回していると Migrate の適用順を壊しても結合テストは緑のまま通り、
// 実起動だけが relation does not exist で落ちる。ここで Migrate を丸ごと流して、適用順と
// 冪等性をまとめて固定する。
package database_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestMigrate_Integration は空の DB へ Migrate を 2 回続けて流し、適用順・冪等性を固定する。
//
// 1 回目はまっさらな schema から始めるので、段の順序を入れ替えると
// 後段が前段の作ったテーブルを見つけられずエラーになる（権限モデルは users を参照する）。
// 2 回目は既に全部揃った状態からの再実行で、こちらが本番の通常起動に相当する。
func TestMigrate_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := t.Context()
	resetPublicSchema(t, db)

	require.NoError(t, database.Migrate(ctx, db), "1 回目（まっさらな DB への適用）")
	require.NoError(t, database.Migrate(ctx, db), "2 回目（本番の通常起動に相当する再適用）")

	t.Run("中核テーブルが揃っている", func(t *testing.T) {
		for _, table := range []string{
			"users", "user_oidc_identities",
			"courses", "course_chapters", "master_exercises", "master_exercise_examples",
			"exercise_submissions", "notes",
			"notifications", "invitations", "audit_events",
			"rich_documents",
		} {
			require.True(t, tableExists(t, db, table), "中核テーブル %s が無い", table)
		}
	})

	t.Run("roles マスタは作られない", func(t *testing.T) {
		// ロールは 3 つで固定され、名前も値も domain.RoleName に直接書いてある
		// （実体は「表」ではなく「コンパイル時の定数」だった）。users.role が値を直接持つ
		// ようになったので、参照先マスタは撤去した。
		require.False(t, tableExists(t, db, "roles"), "roles テーブルが残っている")
	})

	t.Run("退役済みのテナント移行期テーブルは作られない", func(t *testing.T) {
		// companies / company_applications / company_exercises はテナントの正本が
		// workspaces へ完全移行済みのレガシー（本項の撤去 PR で退役）。
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

	t.Run("権限を打ち消す置き場は新しい DB に作られない", func(t *testing.T) {
		// 権限は 3 段の付与（workspace / space / page）を足し合わせ、届いた中で
		// 最も強い役割で決まる。下の段が上の段を弱める仕組みは持たないので、その
		// 置き場だったテーブルが DDL に戻っていないことを見る。
		//
		// 「無いこと」を確かめるのは、うっかり書き戻しても誰も気づかないため。
		// 表があれば読む側がそれを見に行く実装を足せてしまい、規則が 2 つに割れる。
		//
		// **見られるのは新しい DB だけ。** 起動時 DDL に DROP は書いていない（毎回の起動で
		// ロックを取る操作を増やさないため）ので、既にこの表を持っている DB では残り続ける。
		// そちらは 1 回きりの移行 SQL で落とす。順序は「アプリを先に出す → そのあと落とす」で、
		// 新しいアプリはこの表を一切参照しないので、残っていても壊れない。
		for _, table := range []string{"page_restrictions", "page_allow_lists"} {
			require.False(t, tableExists(t, db, table),
				"使わないテーブル %s が作られている（狭める側の仕組みは持たない）", table)
		}
	})

	t.Run("workspace_id 列と FK が張られている", func(t *testing.T) {
		// ここが通るのは、ノート（workspaces）より後に節Ⅰの DO ブロックを流したときだけ。
		require.True(t, columnExists(t, db, "users", "workspace_id"))
		require.True(t, constraintExists(t, db, "users", "fk_users_workspace"))
	})

	t.Run("バックフィル後の制約が張られている", func(t *testing.T) {
		require.True(t, constraintExists(t, db, "users", "ck_users_role"))
		require.True(t, constraintExists(t, db, "user_oidc_identities", "fk_user_oidc_identities_user"))
		require.True(t, constraintExists(t, db, "user_oidc_identities", "ck_user_oidc_identities_not_empty"))
		require.True(t, constraintExists(t, db, "rich_documents", "fk_rich_documents_owner"))
		require.True(t, constraintExists(t, db, "rich_documents", "ck_rich_documents_doc"))
		require.True(t, constraintExists(t, db, "rich_documents", "ck_rich_documents_title_len"))
		require.True(t, indexExists(t, db, "uq_users_email_active"))
	})
}

// TestMigrate_索引が揃っていれば書き込みを止めない_Integration は、本番の通常起動に相当する
// 再適用が中核テーブルのロックを取らないことを固定する。
//
// notes への書き込みトランザクションを開いたまま Migrate を流す。CREATE INDEX を素で
// 発行していると、既存索引でスキップされる場合でも ShareLock を要求して
// RowExclusiveLock と衝突し、Migrate が止まる（そして後続のライターがその後ろに積み上がる）。
func TestMigrate_索引が揃っていれば書き込みを止めない_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := t.Context()
	testsupport.TruncateAll(t, db, "notes")

	// 索引が全部揃った「本番相当」の状態にする。
	require.NoError(t, database.Migrate(ctx, db))

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO notes (user_id, title, content, created_at, updated_at)
		 VALUES (1, 'ロック保持', '', now(), now())`)
	require.NoError(t, err)

	// ロックを取りに行くと lock_timeout の再試行に入り、10 秒では終わらない。
	migrateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	require.NoError(t, database.Migrate(migrateCtx, db),
		"notes への書き込み中でも起動時マイグレーションは通り抜けられなければならない")

	require.NoError(t, tx.Rollback())
}

// resetPublicSchema は public schema を作り直して、まっさらな DB を再現する。
// 結合テストは serializeIntegration で 1 テスト関数ずつ直列に走るので、他のテストと
// 衝突しない（このテストの Migrate が最後にスキーマを作り直して返す）。
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
