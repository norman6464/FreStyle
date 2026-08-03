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

// TestNotificationRepository_CreateMany_Integration は一括作成を実 Postgres で検証する。
// 宛先が増えるたびに DB との往復が増えないよう 1 回の INSERT にまとめている（FRESTYLE-17）。
func TestNotificationRepository_CreateMany_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewNotificationRepository(db)
	ctx := context.Background()

	t.Run("複数件をまとめて作成できる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "notifications")

		ns := []domain.Notification{
			{UserID: 1, Type: "company_application", Title: "申請", Body: "A 社から申請"},
			{UserID: 2, Type: "company_application", Title: "申請", Body: "A 社から申請"},
			{UserID: 3, Type: "company_application", Title: "申請", Body: "A 社から申請"},
		}
		require.NoError(t, repo.CreateMany(ctx, ns))

		// 宛先ごとに 1 件ずつ、本文まで保存されていること。
		for _, want := range ns {
			rows, err := repo.ListByUserID(ctx, want.UserID)
			require.NoError(t, err)
			require.Len(t, rows, 1)
			require.Equal(t, want.Title, rows[0].Title)
			require.Equal(t, want.Body, rows[0].Body)
			require.Equal(t, want.Type, rows[0].Type)
			require.False(t, rows[0].IsRead)
			require.NotZero(t, rows[0].ID, "ID が採番されること")
		}
	})

	t.Run("空スライスは何もせず成功する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "notifications")

		// 宛先が 0 人のときに呼び出し側で分岐しなくて済むようにする。
		require.NoError(t, repo.CreateMany(ctx, nil))
		require.NoError(t, repo.CreateMany(ctx, []domain.Notification{}))

		rows, err := repo.ListByUserID(ctx, 1)
		require.NoError(t, err)
		require.Empty(t, rows)
	})
}
