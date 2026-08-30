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
INSERT INTO invitations (id, company_id, email, role, status) VALUES
  (1, 1, 'invitee@example.test', 'trainee', 'pending');
`

// TestMigrate_移行前のDBから所属を失わずにcompany_idを撤去する_Integration は、
// company_id しか持たない既存 DB へ Migrate() を流したとき、所属が workspace_id へ
// 移ってから company_id が落ちることを固定する。
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

	// company_id は 6 表すべてから消えている（撤去そのもの）。
	var remaining int
	require.NoError(t, db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM information_schema.columns
		  WHERE table_schema = current_schema() AND column_name = 'company_id'`,
	).Scan(&remaining))
	require.Zero(t, remaining, "company_id 列が残っている")

	// 所属していた行は workspace_id を引き継いでいる。会社ごとに別のワークスペースであること
	// まで見る（全行が同じ既定テナントへ流し込まれていないことの担保）。
	wsA := userWorkspace(t, db, 1)
	wsB := userWorkspace(t, db, 2)
	require.True(t, wsA.Valid, "A 社のユーザーが所属を失っている")
	require.True(t, wsB.Valid, "B 社のユーザーが所属を失っている")
	require.NotEqual(t, wsA.String, wsB.String, "別会社のユーザーが同じワークスペースに入っている")

	// 未所属（運営管理者）は NULL のまま。既定のテナントへ流し込まない。
	require.False(t, userWorkspace(t, db, 3).Valid, "未所属のユーザーに所属が付いている")

	// 業務テーブルも同じ移送を受ける。コースは所有会社のワークスペースを引き継ぐ。
	require.Equal(t, wsA.String, tableWorkspace(t, db, `SELECT workspace_id FROM courses WHERE id = 1`))
	require.Equal(t, wsB.String, tableWorkspace(t, db, `SELECT workspace_id FROM courses WHERE id = 2`))
	require.Equal(t, wsA.String, tableWorkspace(t, db, `SELECT workspace_id FROM invitations WHERE id = 1`))

	// 2 回目の Migrate も通る（company_id はもう無いので移送も DROP も no-op）。
	require.NoError(t, database.Migrate(ctx, db), "2 回目の Migrate が失敗した（冪等でない）")
	require.Equal(t, wsA.String, userWorkspace(t, db, 1).String, "2 回目の Migrate で所属が変わった")
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
