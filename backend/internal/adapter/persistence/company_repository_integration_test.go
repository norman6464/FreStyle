//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestCompanyRepository_Integration は sqlc 化した ListAll（name 昇順）/ FindByID（round-trip /
// not-found で ErrRecordNotFound）と、生 SQL の UpdateAiChatEnabled を実 Postgres で検証する。
func TestCompanyRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewCompanyRepository(db)
	ctx := context.Background()

	t.Run("ListAll は name 昇順 / FindByID で round-trip", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "companies")
		// bool の zero値(false)は GORM が省略し DEFAULT true に化けるため、false は UpdateAiChatEnabled で明示する。
		require.NoError(t, db.WithContext(ctx).Create(&domain.Company{ID: 10, Name: "B社"}).Error)
		require.NoError(t, db.WithContext(ctx).Create(&domain.Company{ID: 20, Name: "A社"}).Error)
		require.NoError(t, repo.UpdateAiChatEnabled(ctx, 20, false))

		all, err := repo.ListAll(ctx)
		require.NoError(t, err)
		require.Len(t, all, 2)
		require.Equal(t, "A社", all[0].Name) // name ASC
		require.Equal(t, "B社", all[1].Name)

		// 既定は true（DEFAULT）。
		b, err := repo.FindByID(ctx, 10)
		require.NoError(t, err)
		require.True(t, b.AiChatEnabledForTrainees)

		got, err := repo.FindByID(ctx, 20)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "A社", got.Name)
		require.False(t, got.AiChatEnabledForTrainees) // UpdateAiChatEnabled(false) が効いている
	})

	t.Run("FindByID は not-found で ErrRecordNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "companies")
		_, err := repo.FindByID(ctx, 999)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("UpdateAiChatEnabled で切替できる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "companies")
		require.NoError(t, db.WithContext(ctx).Create(&domain.Company{ID: 30, Name: "C社", AiChatEnabledForTrainees: true}).Error)

		require.NoError(t, repo.UpdateAiChatEnabled(ctx, 30, false))

		got, err := repo.FindByID(ctx, 30)
		require.NoError(t, err)
		require.False(t, got.AiChatEnabledForTrainees)
	})
}
