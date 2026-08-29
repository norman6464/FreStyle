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
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewNotificationRepository(sqlDB)
	ctx := context.Background()

	t.Run("ListByUserID は created_at DESC / CountUnread / MarkRead", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "notifications")

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
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewNotificationRepository(sqlDB)
	ctx := context.Background()

	t.Run("複数件をまとめて作成できる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "notifications")

		ns := []domain.Notification{
			{UserID: 1, Type: "invitation", Title: "招待", Body: "会社から招待"},
			{UserID: 2, Type: "invitation", Title: "招待", Body: "会社から招待"},
			{UserID: 3, Type: "invitation", Title: "招待", Body: "会社から招待"},
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
		testsupport.TruncateAll(t, sqlDB, "notifications")

		// 宛先が 0 人のときに呼び出し側で分岐しなくて済むようにする。
		require.NoError(t, repo.CreateMany(ctx, nil))
		require.NoError(t, repo.CreateMany(ctx, []domain.Notification{}))

		rows, err := repo.ListByUserID(ctx, 1)
		require.NoError(t, err)
		require.Empty(t, rows)
	})
}

// TestNotificationRepository_CreateManyIssuesSingleInsert_Integration は、宛先が増えても
// 1 回の INSERT でまとめて書き込む契約を実 Postgres で固定する。
//
// sqlc 化後は GORM のロガーに乗らない（*sql.DB を直接叩く）ため、発行 SQL 数ではなく
// PostgreSQL のシステム列 xmin（挿入したトランザクション ID）で確かめる。1 回の INSERT なら
// 全行の xmin が一致し、宛先ごとの個別 INSERT へ退行すると xmin が宛先数だけ分かれる。
func TestNotificationRepository_CreateManyIssuesSingleInsert_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	testsupport.TruncateAll(t, sqlDB, "notifications")
	repo := persistence.NewNotificationRepository(sqlDB)

	ns := make([]domain.Notification, 0, 10)
	for i := uint64(1); i <= 10; i++ {
		ns = append(ns, domain.Notification{
			UserID: i, Type: "invitation", Title: "招待", Body: "会社から招待",
		})
	}
	require.NoError(t, repo.CreateMany(context.Background(), ns))

	var distinctTx int64
	require.NoError(t, sqlDB.QueryRow("SELECT COUNT(DISTINCT xmin::text) FROM notifications").Scan(&distinctTx))
	require.Equal(t, int64(1), distinctTx, "宛先 10 件でも 1 トランザクション（1 回の INSERT）でまとめて書く")
}

// TestNotificationRepository_MarkAllRead_Integration は MarkAllRead が対象 user の未読だけを
// 既読化し、他 user に触れないことを実 Postgres で固定する。
func TestNotificationRepository_MarkAllRead_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewNotificationRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "notifications")

	require.NoError(t, repo.Create(ctx, &domain.Notification{UserID: 5, Title: "a"}))
	require.NoError(t, repo.Create(ctx, &domain.Notification{UserID: 5, Title: "b"}))
	require.NoError(t, repo.Create(ctx, &domain.Notification{UserID: 6, Title: "c"}))

	require.NoError(t, repo.MarkAllRead(ctx, 5))

	u5, err := repo.CountUnread(ctx, 5)
	require.NoError(t, err)
	require.Equal(t, int64(0), u5, "user 5 の未読は 0 になる")

	u6, err := repo.CountUnread(ctx, 6)
	require.NoError(t, err)
	require.Equal(t, int64(1), u6, "他 user の通知には触れない")
}
