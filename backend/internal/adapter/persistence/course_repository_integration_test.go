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

// TestCourseRepository_PartialUpdate_Integration は Update が部分更新（title/description/
// sort_order/is_published のみ）であること、すなわち更新対象外の列
// （created_by_user_id / company_id / category / language / created_at）を壊さないことを固定する。
func TestCourseRepository_PartialUpdate_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewCourseRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "courses")

	orig := &domain.Course{
		CompanyID: 1, CreatedByUserID: 42,
		Title: "orig", Description: "orig-desc",
		Category: domain.CourseCategoryBackend, Language: "go",
		SortOrder: 5, IsPublished: false,
	}
	require.NoError(t, repo.Create(ctx, orig))
	created, err := repo.GetByID(ctx, orig.ID)
	require.NoError(t, err)
	origCreatedAt := created.CreatedAt

	// 全フィールドを変えたオブジェクトで Update する（usecase の Update と同じく category/language も
	// 変更した状態で渡す）。repo.Update は 4 列だけ書くので、それ以外は DB 上で変わらないはず。
	created.Title = "updated"
	created.Description = "updated-desc"
	created.Category = domain.CourseCategoryDatabase
	created.Language = "rust"
	created.SortOrder = 99
	created.IsPublished = true
	created.CompanyID = 777       // これは書かれてはいけない
	created.CreatedByUserID = 888 // これも書かれてはいけない
	require.NoError(t, repo.Update(ctx, created))

	got, err := repo.GetByID(ctx, orig.ID)
	require.NoError(t, err)
	// 更新対象（4 列）は反映される。
	require.Equal(t, "updated", got.Title)
	require.Equal(t, "updated-desc", got.Description)
	require.Equal(t, 99, got.SortOrder)
	require.True(t, got.IsPublished)
	// 更新対象外は元のまま（余計な列を書いていない証拠）。
	require.Equal(t, uint64(1), got.CompanyID, "company_id は不変")
	require.Equal(t, uint64(42), got.CreatedByUserID, "created_by_user_id は不変")
	require.Equal(t, domain.CourseCategoryBackend, got.Category, "category は Update の対象外で不変")
	require.Equal(t, "go", got.Language, "language は Update の対象外で不変")
	require.Equal(t, origCreatedAt.Unix(), got.CreatedAt.Unix(), "created_at は不変")
	require.False(t, got.UpdatedAt.Before(origCreatedAt), "updated_at は進む")
}

// TestCourseRepository_CreateDefaults_Integration は sort_order 未指定(0)で作成したとき
// GORM の `default:100` タグ相当で 100 になり、その値が in-memory にも書き戻ることを固定する。
// sortOrder は API で required でないため 0 が入り得る動線で、並び順に効くため保つ。
func TestCourseRepository_CreateDefaults_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewCourseRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "courses")

	c := &domain.Course{CompanyID: 1, CreatedByUserID: 42, Title: "no-sort-order"} // SortOrder=0
	require.NoError(t, repo.Create(ctx, c))
	require.Equal(t, 100, c.SortOrder, "sort_order 未指定は既定 100 が書き戻る")
	require.False(t, c.CreatedAt.IsZero(), "created_at が補完される")
	require.False(t, c.UpdatedAt.IsZero(), "updated_at が補完される")

	got, err := repo.GetByID(ctx, c.ID)
	require.NoError(t, err)
	require.Equal(t, 100, got.SortOrder, "DB 上も 100")
	require.False(t, got.IsPublished, "既定は非公開")
}
