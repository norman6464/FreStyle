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

// TestCompanyApplicationRepository_Integration は本物の PostgreSQL に対して
// CompanyApplicationRepository の Create / ListAll / UpdateStatus を検証する結合テスト。
func TestCompanyApplicationRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	testsupport.TruncateAll(t, db, "company_applications")

	repo := persistence.NewCompanyApplicationRepository(db)
	ctx := context.Background()

	t.Run("Create で保存し ListAll で取得できる", func(t *testing.T) {
		app := &domain.CompanyApplication{
			CompanyName:   "結合テスト株式会社",
			ApplicantName: "山田 太郎",
			Email:         "taro@example.com",
			Message:       "利用したいです",
			Status:        domain.CompanyApplicationStatusPending,
		}
		require.NoError(t, repo.Create(ctx, app))
		require.NotZero(t, app.ID, "autoIncrement で ID が採番される")

		rows, err := repo.ListAll(ctx)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, "結合テスト株式会社", rows[0].CompanyName)
		require.Equal(t, domain.CompanyApplicationStatusPending, rows[0].Status)
	})

	t.Run("ListAll は created_at 降順で返す", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "company_applications")

		older := &domain.CompanyApplication{
			CompanyName: "古い申請", ApplicantName: "A", Email: "a@example.com",
			Status: domain.CompanyApplicationStatusPending,
			// 明示的に古い時刻を入れて並び順を確定させる。
			CreatedAt: time.Now().Add(-time.Hour),
		}
		newer := &domain.CompanyApplication{
			CompanyName: "新しい申請", ApplicantName: "B", Email: "b@example.com",
			Status:    domain.CompanyApplicationStatusPending,
			CreatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, older))
		require.NoError(t, repo.Create(ctx, newer))

		rows, err := repo.ListAll(ctx)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.Equal(t, "新しい申請", rows[0].CompanyName, "新しい順")
		require.Equal(t, "古い申請", rows[1].CompanyName)
	})

	t.Run("UpdateStatus で status を更新できる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "company_applications")

		app := &domain.CompanyApplication{
			CompanyName: "承認される会社", ApplicantName: "C", Email: "c@example.com",
			Status: domain.CompanyApplicationStatusPending,
		}
		require.NoError(t, repo.Create(ctx, app))

		require.NoError(t, repo.UpdateStatus(ctx, app.ID, domain.CompanyApplicationStatusApproved))

		rows, err := repo.ListAll(ctx)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, domain.CompanyApplicationStatusApproved, rows[0].Status)
	})
}
