//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuditRepository_Integration は Record の保存と ListRecent の新しい順 / limit を実 Postgres で検証する。
func TestAuditRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewAuditRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "audit_events")

	require.NoError(t, repo.Record(ctx, &domain.AuditEvent{
		ActorID: 1, ActorEmail: "a@x", ActorRole: domain.RoleSuperAdmin,
		Action: "PATCH /admin/companies/:id/active", TargetID: 1,
	}))
	require.NoError(t, repo.Record(ctx, &domain.AuditEvent{
		ActorID: 1, ActorEmail: "a@x", ActorRole: domain.RoleSuperAdmin,
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
