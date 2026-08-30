//go:build integration

package persistence_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestCourseRepository_Integration は ListByWorkspaceID のワークスペース絞り込み /
// published フィルタ / 並び順を実 Postgres で検証する。
func TestCourseRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCourseRepository(sqlDB)
	ctx := context.Background()

	mk := func(workspaceID *string, title string, published bool, sortOrder int) *domain.Course {
		return &domain.Course{
			WorkspaceID: workspaceID, CreatedByUserID: 1, Title: title,
			IsPublished: published, SortOrder: sortOrder,
		}
	}

	t.Run("ListByWorkspaceID はワークスペースで絞り published フィルタ + sort_order 昇順", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, append([]string{"courses"}, tenantBridgeTables...)...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)
		require.True(t, ws1.Valid)
		ws1Str := ws1.UUID.String()
		ws2 := companyWorkspaceID(t, sqlDB, 2)
		require.True(t, ws2.Valid)
		ws2Str := ws2.UUID.String()

		// Create は呼び出し側（usecase）が解決した workspace_id をそのまま書く。
		require.NoError(t, repo.Create(ctx, mk(&ws1Str, "published", true, 20)))
		require.NoError(t, repo.Create(ctx, mk(&ws1Str, "draft", false, 10)))
		require.NoError(t, repo.Create(ctx, mk(&ws2Str, "other-workspace", true, 5)))

		// includeUnpublished=false → ワークスペース A の published のみ。
		pub, err := repo.ListByWorkspaceID(ctx, ws1Str, false)
		require.NoError(t, err)
		require.Len(t, pub, 1)
		require.Equal(t, "published", pub[0].Title)

		// includeUnpublished=true → ワークスペース A の全件を sort_order 昇順（draft=10, published=20）。
		all, err := repo.ListByWorkspaceID(ctx, ws1Str, true)
		require.NoError(t, err)
		require.Len(t, all, 2)
		require.Equal(t, "draft", all[0].Title, "sort_order 昇順")
		require.Equal(t, "published", all[1].Title)
	})

	t.Run("Create→GetByID→Update→Delete の一連", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "courses")
		ws1 := companyWorkspaceID(t, sqlDB, 1)
		require.True(t, ws1.Valid)
		ws1Str := ws1.UUID.String()

		c := mk(&ws1Str, "lifecycle", true, 1)
		require.NoError(t, repo.Create(ctx, c))
		require.NotZero(t, c.ID)

		got, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		require.Equal(t, "lifecycle", got.Title)
		// GetByID が workspace_id も返すこと（canReadCourse の対象側比較が使う値）。
		require.NotNil(t, got.WorkspaceID)
		require.Equal(t, ws1.UUID.String(), *got.WorkspaceID)

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
// （created_by_user_id / workspace_id / category / language / created_at）を壊さないことを固定する。
func TestCourseRepository_PartialUpdate_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCourseRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, append([]string{"courses"}, tenantBridgeTables...)...)
	insertCompany(t, sqlDB, 1, "会社 A", true)
	insertCompany(t, sqlDB, 2, "会社 B", true)
	runStartupBackfill(ctx, t, sqlDB)
	ws1Str := companyWorkspaceID(t, sqlDB, 1).UUID.String()
	ws2Str := companyWorkspaceID(t, sqlDB, 2).UUID.String()

	orig := &domain.Course{
		WorkspaceID: &ws1Str, CreatedByUserID: 42,
		Title: "orig", Description: "orig-desc",
		Category: domain.CourseCategoryBackend, Language: "go",
		SortOrder: 5, IsPublished: false,
	}
	require.NoError(t, repo.Create(ctx, orig))
	created, err := repo.GetByID(ctx, orig.ID)
	require.NoError(t, err)
	origCreatedAt := created.CreatedAt

	// updated_at を過去へ固定してから Update で進むことを見る。作成時刻（Go の時計で書く）と
	// 更新時刻（DB の now()）を直に突き合わせると、両者のわずかなずれだけで結果が入れ替わる。
	pastUpdatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = sqlDB.Exec(`UPDATE courses SET updated_at = $2 WHERE id = $1`, orig.ID, pastUpdatedAt)
	require.NoError(t, err)

	// 全フィールドを変えたオブジェクトで Update する（usecase の Update と同じく category/language も
	// 変更した状態で渡す）。repo.Update は 4 列だけ書くので、それ以外は DB 上で変わらないはず。
	created.Title = "updated"
	created.Description = "updated-desc"
	created.Category = domain.CourseCategoryDatabase
	created.Language = "rust"
	created.SortOrder = 99
	created.IsPublished = true
	created.WorkspaceID = &ws2Str // これは書かれてはいけない
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
	require.Equal(t, ws1Str, *got.WorkspaceID, "workspace_id は不変")
	require.Equal(t, uint64(42), got.CreatedByUserID, "created_by_user_id は不変")
	require.Equal(t, domain.CourseCategoryBackend, got.Category, "category は Update の対象外で不変")
	require.Equal(t, "go", got.Language, "language は Update の対象外で不変")
	require.Equal(t, origCreatedAt.Unix(), got.CreatedAt.Unix(), "created_at は不変")
	require.True(t, got.UpdatedAt.After(pastUpdatedAt), "updated_at は進む")
}

// TestCourseRepository_CreateDefaults_Integration は sort_order 未指定(0)で作成したとき
// GORM の `default:100` タグ相当で 100 になり、その値が in-memory にも書き戻ることを固定する。
// sortOrder は API で required でないため 0 が入り得る動線で、並び順に効くため保つ。
func TestCourseRepository_CreateDefaults_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCourseRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "courses")

	c := &domain.Course{CreatedByUserID: 42, Title: "no-sort-order"} // SortOrder=0
	require.NoError(t, repo.Create(ctx, c))
	require.Equal(t, 100, c.SortOrder, "sort_order 未指定は既定 100 が書き戻る")
	require.False(t, c.CreatedAt.IsZero(), "created_at が補完される")
	require.False(t, c.UpdatedAt.IsZero(), "updated_at が補完される")

	got, err := repo.GetByID(ctx, c.ID)
	require.NoError(t, err)
	require.Equal(t, 100, got.SortOrder, "DB 上も 100")
	require.False(t, got.IsPublished, "既定は非公開")
}

// TestCourseRepository_GetByIDNotFound_Integration は未存在の GetByID が domain.ErrNotFound を
// 返すこと（handler が 404 に分岐する契約）を固定する。sql.ErrNoRows を素通しすると 404 判定が壊れる。
func TestCourseRepository_GetByIDNotFound_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCourseRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "courses")

	_, err := repo.GetByID(ctx, 999_999)
	require.ErrorIs(t, err, domain.ErrNotFound, "未存在は domain.ErrNotFound（404 シグナル）")
}

// TestCourseRepository_SortOrderBigint_Integration は sort_order が bigint 列であることを
// 往復で固定する。パラメータを int4 に落とすとこの値は負数へ巻き戻り、
// エラーも出ないまま別の値が保存される。
func TestCourseRepository_SortOrderBigint_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewCourseRepository(sqlDB)
	ctx := context.Background()

	testsupport.TruncateAll(t, sqlDB, "courses")
	const beyondInt32 = math.MaxInt32 + 1
	c := &domain.Course{
		CreatedByUserID: 1, Title: "並び順が int32 を超えるコース",
		IsPublished: true, SortOrder: beyondInt32,
	}
	require.NoError(t, repo.Create(ctx, c))

	got, err := repo.GetByID(ctx, c.ID)
	require.NoError(t, err)
	require.Equal(t, beyondInt32, got.SortOrder)
}
