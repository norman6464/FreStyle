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

// TestTeachingMaterialRepository_CountByCourseForCompany_Integration は
// course_id ごとの件数集計 (company 絞り込み / published フィルタ) を実 Postgres で検証する。
func TestTeachingMaterialRepository_CountByCourseForCompany_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewTeachingMaterialRepository(db)
	ctx := context.Background()

	mk := func(companyID, courseID uint64, title string, published bool) *domain.TeachingMaterial {
		return &domain.TeachingMaterial{
			CompanyID: companyID, CourseID: courseID, CreatedByUserID: 1,
			Title: title, Content: "本文", OrderInCourse: 1, IsPublished: published,
		}
	}

	testsupport.TruncateAll(t, db, "teaching_materials")

	// company 1: course 10 に published 2 + draft 1、course 20 に published 1
	require.NoError(t, repo.Create(ctx, mk(1, 10, "c10-pub-1", true)))
	require.NoError(t, repo.Create(ctx, mk(1, 10, "c10-pub-2", true)))
	require.NoError(t, repo.Create(ctx, mk(1, 10, "c10-draft", false)))
	require.NoError(t, repo.Create(ctx, mk(1, 20, "c20-pub", true)))
	// company 2: 他社分は集計に含まれない
	require.NoError(t, repo.Create(ctx, mk(2, 10, "other-company", true)))

	t.Run("published のみ (trainee 相当)", func(t *testing.T) {
		counts, err := repo.CountByCourseForCompany(ctx, 1, false)
		require.NoError(t, err)
		require.Equal(t, map[uint64]int{10: 2, 20: 1}, counts)
	})

	t.Run("下書き込み (admin 相当)", func(t *testing.T) {
		counts, err := repo.CountByCourseForCompany(ctx, 1, true)
		require.NoError(t, err)
		require.Equal(t, map[uint64]int{10: 3, 20: 1}, counts)
	})

	t.Run("教材が無い company は空 map", func(t *testing.T) {
		counts, err := repo.CountByCourseForCompany(ctx, 999, true)
		require.NoError(t, err)
		require.Empty(t, counts)
	})
}
