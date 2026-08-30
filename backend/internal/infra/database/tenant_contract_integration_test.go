//go:build integration

package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// legacySchemaSQL は company_id だけを持ち workspace_id をまだ持たない「移行前の本番 DB」。
// Migrate() がこの状態から所属を失わずに移行できることを固定するための下ごしらえで、
// 実際の本番 DB もこの形（workspace_id 列が無い）から始まる。
const legacySchemaSQL = `
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;

CREATE TABLE companies (
  id bigserial PRIMARY KEY, name text NOT NULL DEFAULT '',
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE roles (
  id integer PRIMARY KEY, name text NOT NULL, description text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO roles (id, name) VALUES (1,'super_admin'),(2,'company_admin'),(3,'trainee');
CREATE TABLE users (
  id bigserial PRIMARY KEY, email text NOT NULL DEFAULT '', password_hash text,
  name text NOT NULL DEFAULT '', company_id bigint, role_id integer NOT NULL DEFAULT 3,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE TABLE courses (
  id bigserial PRIMARY KEY, company_id bigint NOT NULL, created_by_user_id bigint NOT NULL,
  title text NOT NULL DEFAULT '', description text NOT NULL DEFAULT '',
  category text NOT NULL DEFAULT '', language varchar(50) NOT NULL DEFAULT '',
  sort_order bigint NOT NULL DEFAULT 100, is_published boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE course_chapters (
  id bigserial PRIMARY KEY, company_id bigint NOT NULL, course_id bigint NOT NULL,
  created_by_user_id bigint NOT NULL, title text NOT NULL DEFAULT '',
  doc jsonb, revision bigint NOT NULL DEFAULT 1, schema_version bigint NOT NULL DEFAULT 1,
  sort_order bigint NOT NULL DEFAULT 100, is_published boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE company_exercises (
  id bigserial PRIMARY KEY, company_id bigint NOT NULL, language varchar(32) NOT NULL,
  title varchar(200) NOT NULL, description text NOT NULL, starter_code text NOT NULL,
  hint_text text, expected_output text, difficulty smallint NOT NULL DEFAULT 1,
  is_published boolean NOT NULL DEFAULT false, chapter_id bigint, created_by bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE TABLE rich_documents (
  id uuid PRIMARY KEY, owner_id bigint NOT NULL, company_id bigint,
  kind text NOT NULL, title text NOT NULL, is_public boolean NOT NULL DEFAULT false,
  schema_version bigint NOT NULL DEFAULT 1, doc jsonb NOT NULL,
  revision bigint NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE TABLE invitations (
  id bigserial PRIMARY KEY, company_id bigint NOT NULL, email text NOT NULL DEFAULT '',
  role text NOT NULL DEFAULT '', name text NOT NULL DEFAULT '', status text NOT NULL DEFAULT '',
  token varchar(64), expires_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO companies (id, name) VALUES (1, '会社 A'), (2, '会社 B');
INSERT INTO users (id, email, name, company_id, role_id) VALUES
  (1, 'a@example.test', 'A 社の人', 1, 3),
  (2, 'b@example.test', 'B 社の人', 2, 2),
  (3, 'root@example.test', '運営', NULL, 1);
INSERT INTO courses (id, company_id, created_by_user_id, title) VALUES
  (1, 1, 1, 'A 社のコース'),
  (2, 2, 2, 'B 社のコース');
INSERT INTO course_chapters (id, company_id, course_id, created_by_user_id, title) VALUES
  (1, 1, 1, 1, 'A 社の章'),
  (2, 2, 2, 2, 'B 社の章');
INSERT INTO company_exercises (id, company_id, language, title, description, starter_code, created_by) VALUES
  (1, 1, 'go', 'A 社の演習', '説明', 'package main', 1),
  (2, 2, 'go', 'B 社の演習', '説明', 'package main', 2);
INSERT INTO rich_documents (id, owner_id, company_id, kind, title, doc) VALUES
  ('0198a000-0000-7000-8000-00000000d001', 1, 1, 'note', 'A 社のメモ', '{"type":"doc","content":[]}'),
  ('0198a000-0000-7000-8000-00000000d002', 2, 2, 'note', 'B 社のメモ', '{"type":"doc","content":[]}'),
  -- 会社を持たない文書。workspace_id も NULL のまま残る（既定のテナントへ流し込まない）。
  ('0198a000-0000-7000-8000-00000000d003', 3, NULL, 'note', '運営のメモ', '{"type":"doc","content":[]}');
INSERT INTO invitations (id, company_id, email, role, status) VALUES
  (1, 1, 'invitee@example.test', 'trainee', 'pending');
`

// TestMigrate_移行前のDBから所属を失わずにcompany_idを撤去する_Integration は、
// company_id しか持たない既存 DB へ Migrate() を流したとき、所属が workspace_id へ
// 移ってから company_id が落ちることを固定する。
//
// このチケットで落とすのは業務テーブル 4 つ（courses / course_chapters /
// company_exercises / rich_documents）。users / invitations はまだ読み書きされているので
// 残り、毎起動の移送で workspace_id を埋め続ける（それも下で確かめる）。
//
// なぜこのテストが要るか（実際に踏んだ壊れ方）:
//
//	company_id の DROP を schema.sql（= Migrate の最初に流れる ApplyCoreSchema）へ置くと、
//	バックフィルより先に列が消える。対応関係を失った状態でバックフィルが走っても
//	復元できず、既存ユーザー・コース・招待がすべて所属不明になる。撤去は必ず
//	「workspaces を用意 → company_id から移送 → company_id を落とす」の順で行う。
func TestMigrate_移行前のDBから所属を失わずにcompany_idを撤去する_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, legacySchemaSQL)
	require.NoError(t, err, "移行前スキーマの用意に失敗")

	require.NoError(t, database.Migrate(ctx, db), "移行前の DB からの Migrate が失敗した")

	// 業務テーブルからは company_id が消えている（このチケットの撤去対象）。
	for _, table := range []string{"courses", "course_chapters", "company_exercises", "rich_documents"} {
		require.False(t, hasCompanyID(t, db, table), "%s.company_id が残っている", table)
	}
	// users / invitations はまだ読み書きされているので残る（後続のチケットで落とす）。
	for _, table := range []string{"users", "invitations"} {
		require.True(t, hasCompanyID(t, db, table), "%s.company_id を落とすのは後続のチケット", table)
	}

	// 所属していた行は workspace_id を引き継いでいる。会社ごとに別のワークスペースであること
	// まで見る（全行が同じ既定テナントへ流し込まれていないことの担保）。
	wsA := userWorkspace(t, db, 1)
	wsB := userWorkspace(t, db, 2)
	require.True(t, wsA.Valid, "A 社のユーザーが所属を失っている")
	require.True(t, wsB.Valid, "B 社のユーザーが所属を失っている")
	require.NotEqual(t, wsA.String, wsB.String, "別会社のユーザーが同じワークスペースに入っている")

	// 未所属（運営管理者）は NULL のまま。既定のテナントへ流し込まない。
	require.False(t, userWorkspace(t, db, 3).Valid, "未所属のユーザーに所属が付いている")

	// 撤去対象の 4 表すべてが同じ移送を受ける（所有会社のワークスペースを引き継ぐ）。
	// ここを courses だけにすると、他の 3 表は移送も DROP も通らないまま
	// 「一度も存在しなかった列が無い」ことを確かめるだけの検証になる。
	for _, q := range []struct {
		query string
		want  string
	}{
		{`SELECT workspace_id FROM courses WHERE id = 1`, wsA.String},
		{`SELECT workspace_id FROM courses WHERE id = 2`, wsB.String},
		{`SELECT workspace_id FROM course_chapters WHERE id = 1`, wsA.String},
		{`SELECT workspace_id FROM course_chapters WHERE id = 2`, wsB.String},
		{`SELECT workspace_id FROM company_exercises WHERE id = 1`, wsA.String},
		{`SELECT workspace_id FROM company_exercises WHERE id = 2`, wsB.String},
		{`SELECT workspace_id FROM rich_documents WHERE id = '0198a000-0000-7000-8000-00000000d001'`, wsA.String},
		{`SELECT workspace_id FROM rich_documents WHERE id = '0198a000-0000-7000-8000-00000000d002'`, wsB.String},
	} {
		require.Equal(t, q.want, tableWorkspace(t, db, q.query), q.query)
	}
	// 会社を持たない文書は NULL のまま（未所属を既定のテナントへ流し込まない）。
	var loneDoc sql.NullString
	require.NoError(t, db.QueryRow(
		`SELECT workspace_id::text FROM rich_documents WHERE id = '0198a000-0000-7000-8000-00000000d003'`,
	).Scan(&loneDoc))
	require.False(t, loneDoc.Valid, "会社を持たない文書に所属が付いている")

	// invitations は company_id が残るので、毎起動の移送で workspace_id が埋まる。
	require.Equal(t, wsA.String, tableWorkspace(t, db, `SELECT workspace_id FROM invitations WHERE id = 1`))

	// 2 回目の Migrate も通る（company_id はもう無いので移送も DROP も no-op）。
	require.NoError(t, database.Migrate(ctx, db), "2 回目の Migrate が失敗した（冪等でない）")
	require.Equal(t, wsA.String, userWorkspace(t, db, 1).String, "2 回目の Migrate で所属が変わった")
}

// hasCompanyID は指定テーブルに company_id 列が在るかを返す。
func hasCompanyID(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, db.QueryRow(
		`SELECT EXISTS (
		   SELECT 1 FROM information_schema.columns
		    WHERE table_schema = current_schema()
		      AND table_name = $1 AND column_name = 'company_id'
		 )`, table,
	).Scan(&exists))
	return exists
}

func userWorkspace(t *testing.T, db *sql.DB, userID int64) sql.NullString {
	t.Helper()
	var ws sql.NullString
	require.NoError(t, db.QueryRow(`SELECT workspace_id::text FROM users WHERE id = $1`, userID).Scan(&ws))
	return ws
}

func tableWorkspace(t *testing.T, db *sql.DB, query string) string {
	t.Helper()
	var ws sql.NullString
	require.NoError(t, db.QueryRow(query).Scan(&ws))
	require.True(t, ws.Valid, "workspace_id が NULL のまま: %s", query)
	return ws.String
}

// TestDropCompanyIDColumns_移送漏れがあれば撤去せず止まる_Integration は、不可逆な DDL の
// 直前に置いた検査が実際に効くことを固定する。
//
// 移送の SQL は companies.workspace_id が埋まっている会社しか対象にしない。埋まっていない
// 会社が残ったまま company_id を落とすと対応関係が失われる（復元できない）。順序を守っていても
// 起きる壊れ方なので、検査そのものが動くことを見る。
func TestDropCompanyIDColumns_移送漏れがあれば撤去せず止まる_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, legacySchemaSQL)
	require.NoError(t, err, "移行前スキーマの用意に失敗")

	// 会社 → ワークスペースの紐付けだけ先に用意し、そのうち 1 社をわざと未紐付けへ戻す
	// （部分失敗・手動投入でこの状態になり得る）。
	require.NoError(t, database.ApplyCoreSchema(ctx, db))
	require.NoError(t, database.ApplyKnowledgeBaseSchema(ctx, db))
	require.NoError(t, database.BackfillWorkspacesFromCompanies(ctx, db))
	require.NoError(t, database.MigrateWorkspaceIDsFromCompanyID(ctx, db))
	_, err = db.ExecContext(ctx, `UPDATE companies SET workspace_id = NULL WHERE id = 2`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `UPDATE courses SET workspace_id = NULL WHERE company_id = 2`)
	require.NoError(t, err)

	// このテストは共有 DB をわざと壊れた状態にするので、後続のテストへ波及させないよう
	// 必ず migrate 済みの形へ戻してから抜ける（結合テストは 1 台の PostgreSQL を共有する）。
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		require.NoError(t, database.BackfillWorkspacesFromCompanies(cleanupCtx, db))
		require.NoError(t, database.MigrateWorkspaceIDsFromCompanyID(cleanupCtx, db))
		require.NoError(t, database.DropCompanyIDColumns(cleanupCtx, db))
	})

	// 撤去は止まる。
	err = database.DropCompanyIDColumns(ctx, db)
	require.Error(t, err, "移送漏れがあるのに company_id を落とした")
	require.Contains(t, err.Error(), "courses")

	// 列は残っているので、会社の紐付けを直せば移送からやり直せる。
	require.True(t, hasCompanyID(t, db, "courses"), "中断したのに列が落ちている")
}
