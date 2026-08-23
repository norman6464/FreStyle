//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

	testsupport.TruncateAll(t, db, "course_chapters")

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

// TestTeachingMaterialRepository_UpdateDocWithRevision_Integration は
// リッチ本文（tiptap JSON）の jsonb 往復と revision 楽観ロックを実 Postgres で検証する。
func TestTeachingMaterialRepository_UpdateDocWithRevision_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewTeachingMaterialRepository(db)
	ctx := context.Background()

	testsupport.TruncateAll(t, db, "course_chapters")
	m := &domain.TeachingMaterial{
		CompanyID: 1, CourseID: 10, CreatedByUserID: 1,
		Title: "章", Content: "旧 Markdown", OrderInCourse: 1, IsPublished: true,
	}
	require.NoError(t, repo.Create(ctx, m))
	require.Equal(t, 1, m.Revision) // 既定 revision

	doc := `{"type":"doc","content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"見出し"}]}]}`

	t.Run("revision 一致で保存され +1 される（jsonb 往復）", func(t *testing.T) {
		got, err := repo.UpdateDocWithRevision(ctx, m.ID, doc, 1)
		require.NoError(t, err)
		require.Equal(t, 2, got.Revision)
		require.NotNil(t, got.Doc)
		require.Contains(t, *got.Doc, `"heading"`)
		// content（Markdown）は据え置き（移行期間の互換）。
		require.Equal(t, "旧 Markdown", got.Content)
	})

	t.Run("revision 不一致は ErrChapterDocConflict", func(t *testing.T) {
		_, err := repo.UpdateDocWithRevision(ctx, m.ID, doc, 1) // 既に 2 へ進んでいる
		require.ErrorIs(t, err, repository.ErrChapterDocConflict)
	})

	t.Run("存在しない章は gorm.ErrRecordNotFound", func(t *testing.T) {
		_, err := repo.UpdateDocWithRevision(ctx, 99999, doc, 1)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("NUL(U+0000) を含む doc は ErrChapterDocInvalidData", func(t *testing.T) {
		bad := "{\"type\":\"doc\",\"content\":[{\"type\":\"paragraph\",\"content\":[{\"type\":\"text\",\"text\":\"a\\u0000b\"}]}]}"
		_, err := repo.UpdateDocWithRevision(ctx, m.ID, bad, 2)
		require.ErrorIs(t, err, repository.ErrChapterDocInvalidData)
	})
}
