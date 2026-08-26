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

// TestSyncSeededSequences_Integration は、明示 id で seed した表の採番列が実 max に揃い、
// 新規作成が主キー衝突で失敗しないことを実 PostgreSQL で固定する。
//
// 本番はこの同期が無いために companies_id_seq が初期値のままで、会社を新規作成すると
// 初回だけ 23505 で落ちる状態だった。同じ状態を再現してから Migrate を流す。
func TestSyncSeededSequences_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()

	// roles.id は採番列ではない（固定マスタなので既定値なし）ため対象外。
	for _, table := range []string{"companies"} {
		t.Run(table+" は seed 後に採番列が実 max へ揃う", func(t *testing.T) {
			// 本番と同じ「採番列が初期値のまま」を再現する。
			_, err := db.ExecContext(ctx,
				`SELECT setval(pg_get_serial_sequence($1, 'id'), 1, false)`, table)
			require.NoError(t, err)

			require.NoError(t, database.Migrate(ctx, db))

			var maxID, nextVal int64
			require.NoError(t, db.QueryRowContext(ctx,
				`SELECT COALESCE(max(id), 0) FROM `+table).Scan(&maxID))
			require.NoError(t, db.QueryRowContext(ctx,
				`SELECT nextval(pg_get_serial_sequence($1, 'id'))`, table).Scan(&nextVal))

			require.Greater(t, nextVal, maxID,
				"次の採番が既存 id を超えていないと、新規作成が主キー衝突で落ちる")
		})
	}

	t.Run("採番列が初期値のままでも会社を新規作成できる", func(t *testing.T) {
		_, err := db.ExecContext(ctx,
			`SELECT setval(pg_get_serial_sequence('companies', 'id'), 1, false)`)
		require.NoError(t, err)

		require.NoError(t, database.Migrate(ctx, db))

		// 採番に任せた INSERT が 1 回目から通ること（本番はここが 23505 で落ちていた）。
		var newID int64
		err = db.QueryRowContext(ctx,
			`INSERT INTO companies (name, created_at, updated_at)
			 VALUES ('採番確認用', NOW(), NOW()) RETURNING id`).Scan(&newID)
		require.NoError(t, err, "初回の新規作成が主キー衝突で落ちてはいけない")
		require.Greater(t, newID, int64(1))

		_, err = db.ExecContext(ctx, `DELETE FROM companies WHERE id = $1`, newID)
		require.NoError(t, err)
	})
}

var _ = sql.ErrNoRows
