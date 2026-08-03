//go:build integration

package persistence_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

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

// countingLogger は実行された SQL のうち INSERT の回数を数える GORM ロガー。
// 「まとめて 1 回で書き込む」ことを、保存結果ではなく実際に発行された SQL で確かめる。
type countingLogger struct {
	gormlogger.Interface
	mu     sync.Mutex
	insert int
}

func (l *countingLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (string, int64),
	err error,
) {
	sql, _ := fc()
	if strings.Contains(strings.ToUpper(sql), "INSERT INTO") {
		l.mu.Lock()
		l.insert++
		l.mu.Unlock()
	}
	l.Interface.Trace(ctx, begin, fc, err)
}

// TestNotificationRepository_CreateManyIssuesSingleInsert_Integration は、宛先が増えても
// 発行される INSERT が 1 回であることを実 Postgres で検証する（FRESTYLE-17）。
//
// 保存結果の件数だけを見ていると、GORM の CreateBatchSize が有効な環境で複数回に
// 分割されても気づけない。実際に流れた SQL を数えて契約を固定する。
func TestNotificationRepository_CreateManyIssuesSingleInsert_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	testsupport.TruncateAll(t, db, "notifications")

	counter := &countingLogger{Interface: db.Logger}
	repo := persistence.NewNotificationRepository(db.Session(&gorm.Session{Logger: counter}))

	ns := make([]domain.Notification, 0, 10)
	for i := uint64(1); i <= 10; i++ {
		ns = append(ns, domain.Notification{
			UserID: i, Type: "company_application", Title: "申請", Body: "A 社から申請",
		})
	}
	require.NoError(t, repo.CreateMany(context.Background(), ns))

	require.Equal(t, 1, counter.insert, "宛先 10 件でも INSERT は 1 回であること")
}
