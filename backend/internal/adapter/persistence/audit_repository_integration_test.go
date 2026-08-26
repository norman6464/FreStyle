//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuditRepository_Integration は Record の保存と ListRecent の新しい順 / limit を実 Postgres で検証する。
func TestAuditRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewAuditRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "audit_events")

	require.NoError(t, repo.Record(ctx, &domain.AuditEvent{
		ActorID: 1, ActorEmail: "a@x", ActorRole: string(domain.RoleSuperAdmin),
		Action: "PATCH /admin/companies/:id/active", TargetID: 1,
	}))
	require.NoError(t, repo.Record(ctx, &domain.AuditEvent{
		ActorID: 1, ActorEmail: "a@x", ActorRole: string(domain.RoleSuperAdmin),
		Action: "DELETE /admin/members/:userId", TargetID: 2,
	}))

	rows, err := repo.ListRecent(ctx, 10)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	// created_at DESC, id DESC のため後に入れた 2 件目が先頭。
	assert.Equal(t, "DELETE /admin/members/:userId", rows[0].Action)
	assert.Equal(t, uint64(2), rows[0].TargetID)
	assert.False(t, rows[0].CreatedAt.IsZero())

	// limit が効く。
	limited, err := repo.ListRecent(ctx, 1)
	require.NoError(t, err)
	require.Len(t, limited, 1)
}

// TestAuditRepository_Record_Integration は Record の契約
// （採番 id と created_at の書き戻し / 全列の保存 / 明示した created_at は保持）を固定する。
func TestAuditRepository_Record_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewAuditRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "audit_events")

	type auditRow struct {
		ID         uint64
		ActorID    uint64
		ActorEmail string
		ActorRole  string
		Action     string
		TargetID   uint64
		CreatedAt  time.Time
	}
	read := func(id uint64) auditRow {
		var row auditRow
		require.NoError(t, db.Raw(
			"SELECT id, actor_id, actor_email, actor_role, action, target_id, created_at FROM audit_events WHERE id = ?",
			id,
		).Scan(&row).Error)
		return row
	}

	t.Run("採番 id と created_at を書き戻し全列を保存する", func(t *testing.T) {
		e := &domain.AuditEvent{
			ActorID:    11,
			ActorEmail: "admin@example.com",
			ActorRole:  string(domain.RoleSuperAdmin),
			Action:     "PATCH /admin/companies/:id/active",
			TargetID:   22,
		}
		require.NoError(t, repo.Record(ctx, e))
		require.NotZero(t, e.ID, "採番された id が書き戻る")
		require.False(t, e.CreatedAt.IsZero(), "created_at が入る（DB 既定値は無い）")
		require.WithinDuration(t, time.Now(), e.CreatedAt, time.Minute)

		row := read(e.ID)
		require.Equal(t, uint64(11), row.ActorID)
		require.Equal(t, "admin@example.com", row.ActorEmail)
		require.Equal(t, string(domain.RoleSuperAdmin), row.ActorRole)
		require.Equal(t, "PATCH /admin/companies/:id/active", row.Action)
		require.Equal(t, uint64(22), row.TargetID)
		require.WithinDuration(t, e.CreatedAt, row.CreatedAt, time.Second)
	})

	t.Run("空文字 / 0 も欠損せずそのまま入る", func(t *testing.T) {
		e := &domain.AuditEvent{ActorID: 0, ActorEmail: "", ActorRole: "", Action: "", TargetID: 0}
		require.NoError(t, repo.Record(ctx, e))
		row := read(e.ID)
		require.Equal(t, uint64(0), row.ActorID)
		require.Equal(t, "", row.ActorEmail)
		require.Equal(t, "", row.ActorRole)
		require.Equal(t, "", row.Action)
		require.Equal(t, uint64(0), row.TargetID)
	})

	t.Run("明示した created_at は上書きされない", func(t *testing.T) {
		fixed := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
		e := &domain.AuditEvent{ActorID: 1, Action: "act", TargetID: 1, CreatedAt: fixed}
		require.NoError(t, repo.Record(ctx, e))
		row := read(e.ID)
		require.WithinDuration(t, fixed, row.CreatedAt, time.Second)
	})
}
