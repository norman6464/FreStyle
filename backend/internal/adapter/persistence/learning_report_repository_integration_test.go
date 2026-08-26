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

// TestLearningReportRepository_Integration は Create の採番・created_at 自動補完・全フィールド往復と
// ListByUserID の user 絞り込みを実 Postgres で固定する（並び順は list_order_stability が担当）。
func TestLearningReportRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewLearningReportRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "learning_reports")

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 7)

	t.Run("Create は id と created_at を書き戻す（created_at 未設定なら now 補完）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "learning_reports")
		r := &domain.LearningReport{
			UserID: 7, PeriodFrom: from, PeriodTo: to,
			Status: domain.LearningReportStatusPending,
		}
		before := time.Now().UTC().Add(-time.Second)
		require.NoError(t, repo.Create(ctx, r))
		require.NotZero(t, r.ID, "採番 ID が書き戻る")
		require.False(t, r.CreatedAt.IsZero(), "created_at が補完される")
		require.WithinRange(t, r.CreatedAt.UTC(), before, time.Now().UTC().Add(time.Second))

		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, r.ID, rows[0].ID)
		require.Equal(t, uint64(7), rows[0].UserID)
		require.Equal(t, domain.LearningReportStatusPending, rows[0].Status)
		require.Equal(t, "", rows[0].S3Key, "未設定の s3_key は空文字")
		require.Equal(t, from.Unix(), rows[0].PeriodFrom.Unix())
		require.Equal(t, to.Unix(), rows[0].PeriodTo.Unix())
	})

	t.Run("ListByUserID は user で絞る", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "learning_reports")
		require.NoError(t, repo.Create(ctx, &domain.LearningReport{UserID: 7, PeriodFrom: from, PeriodTo: to, Status: domain.LearningReportStatusReady, S3Key: "k7"}))
		require.NoError(t, repo.Create(ctx, &domain.LearningReport{UserID: 8, PeriodFrom: from, PeriodTo: to, Status: domain.LearningReportStatusReady, S3Key: "k8"}))

		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, "k7", rows[0].S3Key, "他ユーザーのレポートは混ざらない")
	})
}
