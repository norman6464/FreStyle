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
			Title: title, OrderInCourse: 1, IsPublished: published,
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
		Title: "章", OrderInCourse: 1, IsPublished: true,
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

// TestTeachingMaterialRepository_CRUD_Integration は読み取り 4・書き込み 5 の振る舞い
// （一覧の doc 除外・並び順・published フィルタ・not-found シグナル・Update の列選択と updated_at 更新・
// 物理削除）を実 Postgres で固定する。GORM→sqlc 移行の前後で同一であることを保証する土台。
func TestTeachingMaterialRepository_CRUD_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewTeachingMaterialRepository(db)
	ctx := context.Background()

	mk := func(companyID, courseID uint64, title string, order int, published bool) *domain.TeachingMaterial {
		return &domain.TeachingMaterial{
			CompanyID: companyID, CourseID: courseID, CreatedByUserID: 7,
			Title: title, OrderInCourse: order, IsPublished: published,
		}
	}

	t.Run("Create は id 採番・既定 revision/schema_version=1・created_at を現在時刻で埋める", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		m := mk(1, 10, "章", 1, true)
		require.NoError(t, repo.Create(ctx, m))
		require.NotZero(t, m.ID)
		require.Equal(t, 1, m.Revision)      // GORM default:1 相当
		require.Equal(t, 1, m.SchemaVersion) // GORM default:1 相当
		require.WithinDuration(t, time.Now(), m.CreatedAt, time.Minute)
		require.WithinDuration(t, time.Now(), m.UpdatedAt, time.Minute)
	})

	t.Run("Create は sort_order=0 のとき既定 100 を当てる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		m := mk(1, 10, "章", 0, true)
		require.NoError(t, repo.Create(ctx, m))
		require.Equal(t, 100, m.OrderInCourse) // GORM default:100 相当
	})

	t.Run("Create は int32 を超える revision / schema_version / sort_order を切り詰めず保存する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		// revision / schema_version / sort_order はいずれも bigint 列。パラメータを int4 に
		// 落とすとこの値は負数へ巻き戻り、エラーも出ないまま別の値が保存される。
		const beyondInt32 = math.MaxInt32 + 1
		m := mk(1, 10, "章", beyondInt32, true)
		m.Revision = beyondInt32
		m.SchemaVersion = beyondInt32
		require.NoError(t, repo.Create(ctx, m))

		got, err := repo.GetByID(ctx, m.ID)
		require.NoError(t, err)
		require.Equal(t, beyondInt32, got.Revision)
		require.Equal(t, beyondInt32, got.SchemaVersion)
		require.Equal(t, beyondInt32, got.OrderInCourse)
	})

	t.Run("GetByID は本文 doc を含めて返し、未存在は gorm.ErrRecordNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		m := mk(1, 10, "章", 1, true)
		require.NoError(t, repo.Create(ctx, m))
		// doc を入れてから GetByID で本文込みで往復することを確認する。
		doc := `{"type":"doc","content":[{"type":"paragraph"}]}`
		_, err := repo.UpdateDocWithRevision(ctx, m.ID, doc, 1)
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, m.ID)
		require.NoError(t, err)
		require.Equal(t, "章", got.Title)
		require.NotNil(t, got.Doc)
		require.Contains(t, *got.Doc, "paragraph")

		_, err = repo.GetByID(ctx, 999999)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("ListByCourse は sort_order 昇順・published フィルタ・doc 本体なし", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		c3 := mk(1, 10, "c3", 3, true)
		c1 := mk(1, 10, "c1", 1, true)
		c2 := mk(1, 10, "c2", 2, false) // draft
		other := mk(1, 20, "other-course", 1, true)
		for _, m := range []*domain.TeachingMaterial{c3, c1, c2, other} {
			require.NoError(t, repo.Create(ctx, m))
		}
		// c1 に doc を入れておく（それでも一覧は doc を返さないことを確認する）。
		_, err := repo.UpdateDocWithRevision(ctx, c1.ID, `{"type":"doc","content":[]}`, 1)
		require.NoError(t, err)

		// published のみ（trainee 相当）: c1, c3 が sort_order 昇順で並ぶ。
		pub, err := repo.ListByCourse(ctx, 10, false)
		require.NoError(t, err)
		require.Len(t, pub, 2)
		require.Equal(t, "c1", pub[0].Title)
		require.Equal(t, "c3", pub[1].Title)
		for _, m := range pub {
			require.Nil(t, m.Doc) // 一覧は本文を読み込まない
		}

		// 下書き込み（admin 相当）: c1, c2, c3。
		all, err := repo.ListByCourse(ctx, 10, true)
		require.NoError(t, err)
		require.Len(t, all, 3)
		require.Equal(t, []string{"c1", "c2", "c3"}, []string{all[0].Title, all[1].Title, all[2].Title})
	})

	t.Run("ListByCompany は会社で絞り・published フィルタ・更新日降順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		a := mk(1, 10, "a", 1, true)
		b := mk(1, 20, "b", 1, false) // draft
		foreign := mk(2, 10, "foreign", 1, true)
		for _, m := range []*domain.TeachingMaterial{a, b, foreign} {
			require.NoError(t, repo.Create(ctx, m))
		}
		// a に doc を入れておく（それでも一覧は doc を返さないことを確認する）。
		// updated_at を now() へ進めるので、先に済ませてから並び順を固定する。
		_, err := repo.UpdateDocWithRevision(ctx, a.ID, `{"type":"doc","content":[]}`, 1)
		require.NoError(t, err)
		// updated_at を明示的に置いて降順を固定する（Go 時計と DB 時計の差でフレークしないように）。
		require.NoError(t, db.Exec(`UPDATE course_chapters SET updated_at = TIMESTAMPTZ '2026-01-01 00:00:00+00' WHERE id = ?`, b.ID).Error)
		require.NoError(t, db.Exec(`UPDATE course_chapters SET updated_at = TIMESTAMPTZ '2026-01-02 00:00:00+00' WHERE id = ?`, a.ID).Error)

		pub, err := repo.ListByCompany(ctx, 1, false)
		require.NoError(t, err)
		require.Len(t, pub, 1) // published の a のみ（b は draft、foreign は他社）
		require.Equal(t, "a", pub[0].Title)
		require.Nil(t, pub[0].Doc) // 一覧は本文を読み込まない（応答でも json:"-" で出ない）

		all, err := repo.ListByCompany(ctx, 1, true)
		require.NoError(t, err)
		require.Len(t, all, 2)
		require.Equal(t, "a", all[0].Title) // updated_at 降順
		require.Equal(t, "b", all[1].Title)
	})

	t.Run("Update は title/sort_order/is_published を書き・不変列を保ち・updated_at を進める", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		m := mk(1, 10, "旧", 1, false)
		require.NoError(t, repo.Create(ctx, m))
		// doc を入れて revision を 2 に進めておく（Update が doc/revision を触らないことの確認用）。
		_, err := repo.UpdateDocWithRevision(ctx, m.ID, `{"type":"doc","content":[]}`, 1)
		require.NoError(t, err)
		// updated_at を過去に固定してから Update で now() に進むことを見る。
		require.NoError(t, db.Exec(`UPDATE course_chapters SET updated_at = TIMESTAMPTZ '2020-01-01 00:00:00+00' WHERE id = ?`, m.ID).Error)

		m.Title = "新"
		m.OrderInCourse = 5
		m.IsPublished = true
		require.NoError(t, repo.Update(ctx, m))

		got, err := repo.GetByID(ctx, m.ID)
		require.NoError(t, err)
		require.Equal(t, "新", got.Title)
		require.Equal(t, 5, got.OrderInCourse)
		require.True(t, got.IsPublished)
		require.Equal(t, uint64(1), got.CompanyID)                        // 不変
		require.Equal(t, uint64(10), got.CourseID)                        // 不変
		require.Equal(t, uint64(7), got.CreatedByUserID)                  // 不変
		require.Equal(t, 2, got.Revision)                                 // 不変（doc 更新の版のまま）
		require.NotNil(t, got.Doc)                                        // 不変（doc は保持）
		require.WithinDuration(t, time.Now(), got.UpdatedAt, time.Minute) // now() に進んだ
	})

	t.Run("Update は存在しない id で gorm.ErrRecordNotFound を返す（黙って成功にしない）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		// 巻き添えで他の行が書き換わらないことも同時に見るため 1 件だけ残しておく。
		keep := mk(1, 10, "残す", 1, true)
		require.NoError(t, repo.Create(ctx, keep))

		ghost := mk(1, 10, "消えた章", 1, true)
		ghost.ID = 999999
		require.ErrorIs(t, repo.Update(ctx, ghost), gorm.ErrRecordNotFound)

		got, err := repo.GetByID(ctx, keep.ID)
		require.NoError(t, err)
		require.Equal(t, "残す", got.Title) // 別の行は触られていない
	})

	t.Run("Update は取得後に消えた章でも gorm.ErrRecordNotFound を返す", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		m := mk(1, 10, "章", 1, true)
		require.NoError(t, repo.Create(ctx, m))
		// usecase は Update の前に GetByID する。その隙に行が消える競合を再現する。
		require.NoError(t, repo.Delete(ctx, m.ID))

		m.Title = "編集後"
		require.ErrorIs(t, repo.Update(ctx, m), gorm.ErrRecordNotFound)
	})

	t.Run("Delete / DeleteByCourse は存在しない id でも冪等に nil（後条件は満たされている）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		keep := mk(1, 20, "keep", 1, true)
		require.NoError(t, repo.Create(ctx, keep))

		// 「行が無い」という後条件は既に満たされているため 0 行削除はエラーにしない
		// （移行前の GORM 版も 0 行で nil を返していた。Update と違い失われた書き込みが無い）。
		require.NoError(t, repo.Delete(ctx, 999999))
		require.NoError(t, repo.DeleteByCourse(ctx, 999999))

		survived, err := repo.ListByCourse(ctx, 20, true)
		require.NoError(t, err)
		require.Len(t, survived, 1) // 巻き添え削除が起きていない
	})

	t.Run("Delete は 1 件を物理削除する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		m := mk(1, 10, "章", 1, true)
		require.NoError(t, repo.Create(ctx, m))
		require.NoError(t, repo.Delete(ctx, m.ID))
		_, err := repo.GetByID(ctx, m.ID)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("DeleteByCourse はコース配下を全削除し他コースは残す", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "course_chapters")
		a1 := mk(1, 10, "a1", 1, true)
		a2 := mk(1, 10, "a2", 2, true)
		keep := mk(1, 20, "keep", 1, true)
		for _, m := range []*domain.TeachingMaterial{a1, a2, keep} {
			require.NoError(t, repo.Create(ctx, m))
		}
		require.NoError(t, repo.DeleteByCourse(ctx, 10))

		gone, err := repo.ListByCourse(ctx, 10, true)
		require.NoError(t, err)
		require.Empty(t, gone)
		survived, err := repo.ListByCourse(ctx, 20, true)
		require.NoError(t, err)
		require.Len(t, survived, 1)
	})
}
