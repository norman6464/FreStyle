//go:build integration

// Package database_test: users.id を指す外部キー（段 1）の削除時の挙動そのものを検証する。
// schema_apply_integration_test.go は「制約が存在するか」までしか見ないので、ここでは
// 実際に行を消してみて、持ち物（CASCADE）と記録（RESTRICT）が宣言どおりに動くことを確かめる。
package database_test

import (
	"testing"

	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestUsersForeignKey_Integration は users.id への外部キー方針（段 1。schema.hcl の users
// テーブル直後のコメント参照）を、代表列 1 つずつで実地に確かめる。
func TestUsersForeignKey_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := t.Context()

	insertUser := func(t *testing.T) int64 {
		t.Helper()
		var id int64
		require.NoError(t, db.QueryRowContext(ctx,
			`INSERT INTO users (id, email, name, created_at, updated_at)
			 VALUES ((SELECT COALESCE(MAX(id), 0) + 1 FROM users), $1, 'fk-test', now(), now())
			 RETURNING id`,
			"fk-test+"+t.Name()+"@example.test").Scan(&id))
		return id
	}

	t.Run("持ち物CASCADE_notificationsは本人の削除で一緒に消える", func(t *testing.T) {
		userID := insertUser(t)
		_, err := db.ExecContext(ctx,
			`INSERT INTO notifications (user_id, type, title, body, is_read, created_at)
			 VALUES ($1, 'test', 't', '', false, now())`, userID)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
		require.NoError(t, err, "持ち物しか無いユーザーの削除は通るはず（CASCADE）")

		var remaining int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT count(*) FROM notifications WHERE user_id = $1`, userID).Scan(&remaining))
		require.Zero(t, remaining, "notifications も一緒に消えているはず")
	})

	t.Run("記録RESTRICT_pagesの作成者が残っていると本人は消せない", func(t *testing.T) {
		userID := insertUser(t)
		var wsID, spaceID string
		require.NoError(t, db.QueryRowContext(
			ctx,
			`INSERT INTO workspaces (id, slug, name) VALUES (gen_random_uuid(), 'fk-test', 'fk-test') RETURNING id`,
		).Scan(&wsID))
		require.NoError(t, db.QueryRowContext(ctx,
			`INSERT INTO spaces (id, workspace_id, "key", name) VALUES (gen_random_uuid(), $1, 'fk', 'fk-test') RETURNING id`,
			wsID).Scan(&spaceID))
		_, err := db.ExecContext(ctx,
			`INSERT INTO pages (id, workspace_id, space_id, "position", title, created_by_user_id)
			 VALUES (gen_random_uuid(), $1, $2, 'a', 't', $3)`, wsID, spaceID, userID)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
		require.Error(t, err, "作成記録が残っている限りユーザーは消せないはず（RESTRICT）")
		require.Contains(t, err.Error(), "fk_pages_created_by")

		// 後始末: pages → workspaces（CASCADE）→ user の順で消す。
		_, err = db.ExecContext(ctx, `DELETE FROM workspaces WHERE id = $1`, wsID)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
		require.NoError(t, err)
	})

	// 本番のような非空テーブルへ安全に FK を追加する手順（NOT VALID で追加 → 既存行を
	// スキャンしない VALIDATE CONSTRAINT で確かめる）そのものを、わざと孤児行を 1 つ作って
	// 再現する。ADD 時点では孤児があっても通り、VALIDATE で初めて検出されることを確かめる。
	t.Run("記録RESTRICT_NOT_VALIDで追加した直後は孤児があっても通りVALIDATEで検出される", func(t *testing.T) {
		_, err := db.ExecContext(ctx, `ALTER TABLE notifications DROP CONSTRAINT fk_notifications_user`)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = db.ExecContext(ctx, `DELETE FROM notifications WHERE user_id = 999999999`)
			_, _ = db.ExecContext(ctx,
				`ALTER TABLE notifications ADD CONSTRAINT fk_notifications_user
				   FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE NO ACTION ON DELETE CASCADE`)
		})

		_, err = db.ExecContext(ctx,
			`INSERT INTO notifications (user_id, type, title, body, is_read, created_at)
			 VALUES (999999999, 'test', 'orphan', '', false, now())`)
		require.NoError(t, err, "FK を外した状態なので孤児行はそのまま入る")

		_, err = db.ExecContext(ctx,
			`ALTER TABLE notifications ADD CONSTRAINT fk_notifications_user
			   FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE NO ACTION ON DELETE CASCADE NOT VALID`)
		require.NoError(t, err, "NOT VALID は既存行を検査しないので孤児があっても追加できる")

		_, err = db.ExecContext(ctx, `ALTER TABLE notifications VALIDATE CONSTRAINT fk_notifications_user`)
		require.Error(t, err, "孤児行がある限り VALIDATE は落ちるはず")
		require.Contains(t, err.Error(), "fk_notifications_user")

		_, err = db.ExecContext(ctx, `DELETE FROM notifications WHERE user_id = 999999999`)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `ALTER TABLE notifications VALIDATE CONSTRAINT fk_notifications_user`)
		require.NoError(t, err, "孤児を片付ければ VALIDATE は通るはず")
	})
}
