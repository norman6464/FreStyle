//go:build integration

// Package database_test は users.is_active / deleted_at を users.status 1 列へまとめる
// 移行（段 3）そのものを検証する。
//
// Atlas の宣言的 schema apply は「まだ何も無い空の DB」への最終形の適用と、
// 「実 DB との構造差分（列・制約・索引）の計算」はできるが、backfill（既存行の値から
// 新しい列の値を計算する）は行わない。列追加のデフォルト値は全行一律にしか効かないため、
// is_active=false / deleted_at IS NOT NULL の行が混じった DB へ素の schema apply を
// かけると、それらの行が黙って status='active'（デフォルト値）になってしまう。
//
// 本番は 2026-09-11 時点で 5 行とも is_active=true・deleted_at NULL（確認済み。
// supabase db query --linked）なので、今回の本番適用は素の schema apply で安全に足りる。
// このテストは「まっさらな DB のテストでは通ってしまう」（段 3 チケットが名指しした
// リスク）に備えて、移行前の姿（3 通りの組み合わせ）を手動で再現し、正しい手順で
// backfill してから列を落とせば意図どおりの status になること、backfill を端折ると
// ck_users_status_deleted_at が検知して止めることの両方を固定する。
package database_test

import (
	"testing"
	"time"

	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestUserStatusMigration_Integration は is_active/deleted_at → status の移行手順を、
// 「移行前の姿」の users 相当テーブルに対して実際に流して確かめる。
func TestUserStatusMigration_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := t.Context()

	// legacy_users は「段 3 適用前の users」を最小限で再現したテーブル。実テーブルと
	// 名前を分け、他の結合テストの TruncateAll / FK と無関係に単独で作って消せるようにする。
	setupLegacyTable := func(t *testing.T) {
		t.Helper()
		_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS legacy_users`)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `
			CREATE TABLE legacy_users (
				id bigserial PRIMARY KEY,
				email text NOT NULL DEFAULT '',
				is_active boolean NOT NULL DEFAULT true,
				deleted_at timestamptz NULL
			)`)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS legacy_users`)
		})
	}

	// seedThreeCombinations は意味のある 3 通りの組み合わせ（有効・停止・退会）を入れる。
	// 「退会済みだが有効」という 4 つ目の組み合わせは、そもそも誰も作っていなかった
	// 想定なので入れない。
	seedThreeCombinations := func(t *testing.T) (activeID, suspendedID, deactivatedID int64) {
		t.Helper()
		require.NoError(t, db.QueryRowContext(
			ctx,
			`INSERT INTO legacy_users (email, is_active, deleted_at) VALUES ('active@example.test', true, NULL) RETURNING id`,
		).Scan(&activeID))
		require.NoError(t, db.QueryRowContext(
			ctx,
			`INSERT INTO legacy_users (email, is_active, deleted_at) VALUES ('suspended@example.test', false, NULL) RETURNING id`,
		).Scan(&suspendedID))
		require.NoError(t, db.QueryRowContext(
			ctx,
			`INSERT INTO legacy_users (email, is_active, deleted_at) VALUES ('deactivated@example.test', false, now()) RETURNING id`,
		).Scan(&deactivatedID))
		return
	}

	t.Run("正しい手順でbackfillすると3通りの状態が意図どおりstatusに写る", func(t *testing.T) {
		setupLegacyTable(t)
		activeID, suspendedID, deactivatedID := seedThreeCombinations(t)

		// 拡張: NULL 許可で status を足す（この時点では NOT NULL 制約も CHECK も無い）。
		_, err := db.ExecContext(ctx, `ALTER TABLE legacy_users ADD COLUMN status text`)
		require.NoError(t, err)

		// 移送: 既存 2 列の値から status を計算する。deleted_at の有無を優先し、
		// 次に is_active を見る（「退会済みだが有効」という組み合わせは無い前提だが、
		// 万一あっても退会を優先するほうが安全側に倒れる）。
		_, err = db.ExecContext(ctx, `
			UPDATE legacy_users SET status = CASE
				WHEN deleted_at IS NOT NULL THEN 'deactivated'
				WHEN NOT is_active THEN 'suspended'
				ELSE 'active'
			END`)
		require.NoError(t, err)

		// 縮小: NOT NULL 化・CHECK 制約・is_active の削除。制約は schema.hcl の
		// ck_users_status / ck_users_status_deleted_at と同じ定義にする。
		for _, stmt := range []string{
			`ALTER TABLE legacy_users ALTER COLUMN status SET NOT NULL`,
			`ALTER TABLE legacy_users ALTER COLUMN status SET DEFAULT 'active'`,
			`ALTER TABLE legacy_users ADD CONSTRAINT ck_users_status
				CHECK (status = ANY (ARRAY['active'::text, 'suspended'::text, 'deactivated'::text]))`,
			`ALTER TABLE legacy_users ADD CONSTRAINT ck_users_status_deleted_at
				CHECK ((status = 'deactivated'::text) = (deleted_at IS NOT NULL))`,
			`ALTER TABLE legacy_users DROP COLUMN is_active`,
		} {
			_, err := db.ExecContext(ctx, stmt)
			require.NoError(t, err, "手順: %s", stmt)
		}

		statusOf := func(id int64) string {
			var s string
			require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM legacy_users WHERE id = $1`, id).Scan(&s))
			return s
		}
		require.Equal(t, "active", statusOf(activeID))
		require.Equal(t, "suspended", statusOf(suspendedID))
		require.Equal(t, "deactivated", statusOf(deactivatedID))
	})

	t.Run("backfillを端折るとck_users_status_deleted_atが追加時に検知する", func(t *testing.T) {
		setupLegacyTable(t)
		_, _, deactivatedID := seedThreeCombinations(t)

		// わざと不完全な backfill: deleted_at を見ず is_active だけで決める
		// （典型的な書き間違い。退会済み行が deleted_at を持ったまま status='suspended' になる）。
		_, err := db.ExecContext(ctx, `ALTER TABLE legacy_users ADD COLUMN status text`)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `
			UPDATE legacy_users SET status = CASE WHEN is_active THEN 'active' ELSE 'suspended' END`)
		require.NoError(t, err)

		// ここまでは通る（status 単体の CHECK は満たす）。
		_, err = db.ExecContext(ctx, `
			ALTER TABLE legacy_users ADD CONSTRAINT ck_users_status
				CHECK (status = ANY (ARRAY['active'::text, 'suspended'::text, 'deactivated'::text]))`)
		require.NoError(t, err)

		// deleted_at との整合 CHECK で不整合（status='suspended' なのに deleted_at が入っている）
		// が検知され、追加そのものが失敗する。
		_, err = db.ExecContext(ctx, `
			ALTER TABLE legacy_users ADD CONSTRAINT ck_users_status_deleted_at
				CHECK ((status = 'deactivated'::text) = (deleted_at IS NOT NULL))`)
		require.ErrorContains(t, err, "ck_users_status_deleted_at")

		// 実際に退会済み行が壊れたままであることも確認しておく（検知の理由を裏付ける）。
		var status string
		require.NoError(t, db.QueryRowContext(ctx, `SELECT status FROM legacy_users WHERE id = $1`, deactivatedID).Scan(&status))
		require.Equal(t, "suspended", status, "backfill を端折ると退会済み行が誤って suspended になる")
	})

	// ここからは移行後の実テーブルに対して、CHECK 制約が実際に不正値を拒否することを確かめる。
	insertUser := func(t *testing.T, email, status string, deletedAt *time.Time) error {
		t.Helper()
		_, err := db.ExecContext(
			ctx,
			`INSERT INTO users (email, name, status, deleted_at, created_at, updated_at)
			 VALUES ($1, $1, $2, $3, now(), now())`,
			email, status, deletedAt,
		)
		return err
	}
	now := time.Now()

	t.Run("ck_users_statusは3値以外を拒否する", func(t *testing.T) {
		require.ErrorContains(t, insertUser(t, "bogus-status@example.test", "banned", nil), "ck_users_status")
	})

	t.Run("ck_users_status_deleted_atはactiveなのにdeleted_atがある行を拒否する", func(t *testing.T) {
		err := insertUser(t, "active-but-deleted@example.test", "active", &now)
		require.ErrorContains(t, err, "ck_users_status_deleted_at")
	})

	t.Run("ck_users_status_deleted_atはdeactivatedなのにdeleted_atが無い行を拒否する", func(t *testing.T) {
		err := insertUser(t, "deactivated-but-not-deleted@example.test", "deactivated", nil)
		require.ErrorContains(t, err, "ck_users_status_deleted_at")
	})

	t.Run("整合する組み合わせは通る", func(t *testing.T) {
		require.NoError(t, insertUser(t, "consistent-active@example.test", "active", nil))
		require.NoError(t, insertUser(t, "consistent-suspended@example.test", "suspended", nil))
		require.NoError(t, insertUser(t, "consistent-deactivated@example.test", "deactivated", &now))
	})
}
