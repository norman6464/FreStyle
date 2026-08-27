//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestCompanyRepository_Integration は sqlc 化した ListAll（name 昇順）/ FindByID（round-trip /
// not-found で domain.ErrNotFound）を実 Postgres で検証する。
func TestCompanyRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCompanyRepository(sqlDB)
	ctx := context.Background()

	t.Run("ListAll は name 昇順 / FindByID で round-trip", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "companies")
		for _, c := range []struct {
			id   uint64
			name string
		}{{10, "B社"}, {20, "A社"}} {
			_, err := sqlDB.ExecContext(ctx,
				`INSERT INTO companies (id, name, created_at, updated_at) VALUES ($1, $2, now(), now())`,
				c.id, c.name)
			require.NoError(t, err)
		}

		all, err := repo.ListAll(ctx)
		require.NoError(t, err)
		require.Len(t, all, 2)
		require.Equal(t, "A社", all[0].Name) // name ASC
		require.Equal(t, "B社", all[1].Name)

		got, err := repo.FindByID(ctx, 20)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "A社", got.Name)
	})

	t.Run("FindByID は not-found で domain.ErrNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "companies")
		_, err := repo.FindByID(ctx, 999)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}
