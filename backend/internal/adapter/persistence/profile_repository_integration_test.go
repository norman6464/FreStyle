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

// TestProfileRepository_Integration は sqlc 化した FindByUserID（round-trip / not-found）を実 Postgres で検証する。
func TestProfileRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewProfileRepository(db)
	ctx := context.Background()

	t.Run("FindByUserID は profile を返す", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "profiles")
		require.NoError(t, db.WithContext(ctx).Create(&domain.Profile{
			UserID: 7, Bio: "自己紹介", AvatarURL: "https://example.com/a.png", Status: "active",
		}).Error)

		got, err := repo.FindByUserID(ctx, 7)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, uint64(7), got.UserID)
		require.Equal(t, "自己紹介", got.Bio)
		require.Equal(t, "https://example.com/a.png", got.AvatarURL)
		require.Equal(t, "active", got.Status)
	})

	t.Run("未作成は (nil, nil)", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "profiles")
		got, err := repo.FindByUserID(ctx, 999)
		require.NoError(t, err)
		require.Nil(t, got)
	})
}
