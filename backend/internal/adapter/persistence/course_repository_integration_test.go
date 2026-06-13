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

// TestCourseRepository_Integration は ListByCompany の company 絞り込み / published フィルタ /
// 並び順を実 Postgres で検証する。
func TestCourseRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewCourseRepository(db)
	ctx := context.Background()

	mk := func(companyID uint64, title string, published bool, sortOrder int) *domain.Course {
		return &domain.Course{
			CompanyID: companyID, CreatedByUserID: 1, Title: title,
			IsPublished: published, SortOrder: sortOrder,
		}
	}

	t.Run("ListByCompany は company で絞り published フィルタ + sort_order 昇順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "courses")

		require.NoError(t, repo.Create(ctx, mk(1, "published", true, 20)))
		require.NoError(t, repo.Create(ctx, mk(1, "draft", false, 10)))
		require.NoError(t, repo.Create(ctx, mk(2, "other-company", true, 5)))

		// includeUnpublished=false → company 1 の published のみ。
		pub, err := repo.ListByCompany(ctx, 1, false)
		require.NoError(t, err)
		require.Len(t, pub, 1)
		require.Equal(t, "published", pub[0].Title)

		// includeUnpublished=true → company 1 の全件を sort_order 昇順（draft=10, published=20）。
		all, err := repo.ListByCompany(ctx, 1, true)
		require.NoError(t, err)
		require.Len(t, all, 2)
		require.Equal(t, "draft", all[0].Title, "sort_order 昇順")
		require.Equal(t, "published", all[1].Title)
	})

	t.Run("Create→GetByID→Update→Delete の一連", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "courses")

		c := mk(1, "lifecycle", true, 1)
		require.NoError(t, repo.Create(ctx, c))
		require.NotZero(t, c.ID)

		got, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		require.Equal(t, "lifecycle", got.Title)

		got.Title = "updated"
		require.NoError(t, repo.Update(ctx, got))
		reread, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		require.Equal(t, "updated", reread.Title)

		require.NoError(t, repo.Delete(ctx, c.ID))
		_, err = repo.GetByID(ctx, c.ID)
		require.Error(t, err)
	})
}
