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
// 後段が前段の作ったテーブルを見つけられずエラーになる
// （権限モデルは users を、テナント橋渡しは workspaces を参照する）。
// 2 回目は既に全部揃った状態からの再実行で、こちらが本番の通常起動に相当する。
func TestMigrate_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := t.Context()
	resetPublicSchema(t, db)

	require.NoError(t, database.Migrate(ctx, db), "1 回目（まっさらな DB への適用）")
	require.NoError(t, database.Migrate(ctx, db), "2 回目（本番の通常起動に相当する再適用）")

	t.Run("中核テーブルが揃っている", func(t *testing.T) {
		for _, table := range []string{
			"roles", "users", "user_oidc_identities", "companies", "company_applications",
			"courses", "course_chapters", "master_exercises", "master_exercise_examples",
			"company_exercises", "exercise_submissions", "notes",
			"notifications", "invitations", "audit_events",
			"rich_documents",
		} {
			require.True(t, tableExists(t, db, table), "中核テーブル %s が無い", table)
		}
	})

	t.Run("ノートと権限モデルが揃っている", func(t *testing.T) {
		for _, table := range []string{
			"workspaces", "spaces", "pages", "blocks", "page_paths", "page_snapshots",
			"principals", "principal_members", "workspace_grants", "space_grants",
			"page_restrictions", "page_allow_lists", "share_links",
		} {
			require.True(t, tableExists(t, db, table), "ノートのテーブル %s が無い", table)
		}
	})

	t.Run("テナント橋渡しの列と FK が張られている", func(t *testing.T) {
		// ここが通るのは、ノート（workspaces）より後にテナント橋渡しを流したときだけ。
		require.True(t, columnExists(t, db, "companies", "workspace_id"))
		require.True(t, columnExists(t, db, "users", "workspace_id"))
		require.True(t, constraintExists(t, db, "companies", "fk_companies_workspace"))
		require.True(t, constraintExists(t, db, "users", "fk_users_workspace"))
		require.True(t, indexExists(t, db, "uq_companies_workspace_id"))
	})

	t.Run("バックフィル後の制約が張られている", func(t *testing.T) {
		require.True(t, constraintExists(t, db, "roles", "ck_roles_name_not_empty"))
		require.True(t, constraintExists(t, db, "users", "fk_users_role"))
		require.True(t, constraintExists(t, db, "user_oidc_identities", "fk_user_oidc_identities_user"))
		require.True(t, constraintExists(t, db, "user_oidc_identities", "ck_user_oidc_identities_not_empty"))
		require.True(t, constraintExists(t, db, "rich_documents", "fk_rich_documents_owner"))
		require.True(t, constraintExists(t, db, "rich_documents", "ck_rich_documents_doc"))
		require.True(t, constraintExists(t, db, "rich_documents", "ck_rich_documents_title_len"))
		require.True(t, indexExists(t, db, "uq_users_email_active"))
	})

	t.Run("Migrate はデータを書かない（roles / companies を seed しない）", func(t *testing.T) {
		var roles, companies int64
		require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM roles`).Scan(&roles))
		require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM companies`).Scan(&companies))
		require.EqualValues(t, 0, roles)
		require.EqualValues(t, 0, companies)
	})

	t.Run("会社があれば、ワークスペースへのバックフィルが 1 回だけ効く", func(t *testing.T) {
		var companyID int64
		require.NoError(t, db.QueryRowContext(ctx,
			`INSERT INTO companies (name, created_at, updated_at) VALUES ('検証用の会社', NOW(), NOW()) RETURNING id`,
		).Scan(&companyID))
		require.NoError(t, database.Migrate(ctx, db))

		var workspaces int64
		require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM workspaces`).Scan(&workspaces))
		require.EqualValues(t, 1, workspaces)

		var slug string
		require.NoError(t, db.QueryRowContext(
			ctx,
			`SELECT w.slug FROM workspaces w JOIN companies c ON c.workspace_id = w.id WHERE c.id = $1`, companyID,
		).Scan(&slug))
		require.Regexp(t, `^ws-[0-9a-f]{32}$`, slug)

		require.NoError(t, database.Migrate(ctx, db), "2 回目")
		require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM workspaces`).Scan(&workspaces))
		require.EqualValues(t, 1, workspaces, "2 回流してもワークスペースは増えない")
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
