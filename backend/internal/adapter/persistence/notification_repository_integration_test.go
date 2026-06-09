//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestNotificationRepository_Integration は sqlc 化した ListByUserID（created_at DESC・user 絞り）と
// CountUnread を実 Postgres で検証する。MarkRead で未読数が減ることも確認する。
func TestNotificationRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewNotificationRepository(db)
	ctx := context.Background()

	t.Run("ListByUserID は created_at DESC / CountUnread / MarkRead", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "notifications")

		older := time.Now().Add(-time.Hour)
		newer := time.Now()
		require.NoError(t, repo.Create(ctx, &domain.Notification{UserID: 7, Title: "古い", CreatedAt: older}))
		require.NoError(t, repo.Create(ctx, &domain.Notification{UserID: 7, Title: "新しい", CreatedAt: newer}))
		require.NoError(t, repo.Create(ctx, &domain.Notification{UserID: 99, Title: "他人"}))

		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.Equal(t, "新しい", rows[0].Title) // created_at DESC
		require.Equal(t, "古い", rows[1].Title)

		unread, err := repo.CountUnread(ctx, 7)
		require.NoError(t, err)
		require.Equal(t, int64(2), unread)

		require.NoError(t, repo.MarkRead(ctx, 7, rows[0].ID))
		unread, err = repo.CountUnread(ctx, 7)
		require.NoError(t, err)
		require.Equal(t, int64(1), unread)
	})
}
