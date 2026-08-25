//go:build integration

package persistence_test

import (
	"context"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// 一覧クエリの並びが「ソートキーの同着」で崩れないことを横断的に固定する。
//
// 非一意な列だけで ORDER BY すると同値行の相対順序は SQL 上未定義で、実行計画・ページ境界・
// 物理配置で変わりうる。ページングと組み合わさると同じ行の重複表示や欠落になる。
//
// 各ケースは意図的に同着を作り、さらに「投入順（＝ヒープの物理順）」を「期待順」の逆にしてある。
// タイブレークを外すと素の走査順がそのまま返り、期待順と食い違って必ず落ちる（テストが
// 空回りしていないことの担保）。

// collectAllPages は Limit/Offset を進めて全ページを取得し、出現順の ID 列を返す。
func collectAllPages(t *testing.T, exRepo repository.MasterExerciseRepository, language string, limit int) []uint64 {
	t.Helper()
	ctx := context.Background()
	ids := make([]uint64, 0)
	for offset := 0; ; offset += limit {
		rows, err := exRepo.ListWithStatusByLanguage(ctx, repository.ListWithStatusInput{
			Language: language,
			Offset:   offset,
			Limit:    limit,
		})
		require.NoError(t, err)
		if len(rows) == 0 {
			return ids
		}
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		if len(rows) < limit {
			return ids
		}
	}
}

// TestMasterExerciseListOrder_TiedSortOrder_Integration は sort_order 同着でも
// OFFSET ページングが重複・欠落を起こさないことを検証する。
func TestMasterExerciseListOrder_TiedSortOrder_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	exRepo := persistence.NewMasterExerciseRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "master_exercises", "exercise_submissions")

	// php 40 件 / go 5 件とも sort_order を全行 1 に揃えて同着を作る。ID は降順に投入して
	// 「物理順 ≠ 期待順（ID 昇順）」にする。
	// 40 件あるのは、LIMIT+OFFSET が小さいと PostgreSQL が top-N heapsort を選び、
	// N（= limit+offset）ごとに同着の並びが変わるため。件数が少ないと素の整列で偶然揃ってしまい、
	// 「ページ間で並びが変わる → 重複・欠落」という本来の症状を再現できない。
	var phpIDs, goIDs []uint64
	for id := uint64(140); id >= 101; id-- {
		phpIDs = append(phpIDs, id)
	}
	for id := uint64(205); id >= 201; id-- {
		goIDs = append(goIDs, id)
	}
	insert := func(ids []uint64, language string) {
		for _, id := range ids {
			row := domain.MasterExercise{
				ID:          id,
				Slug:        language + "-tie-" + strconv.FormatUint(id, 10),
				Language:    language,
				Title:       "tie",
				SortOrder:   1,
				IsPublished: true,
			}
			require.NoError(t, db.WithContext(ctx).Create(&row).Error)
		}
	}
	insert(phpIDs, domain.ExerciseLanguagePhp)
	insert(goIDs, domain.ExerciseLanguageGo)

	ascending := func(from, to uint64) []uint64 {
		out := make([]uint64, 0, to-from+1)
		for id := from; id <= to; id++ {
			out = append(out, id)
		}
		return out
	}
	wantPHP := ascending(101, 140)
	wantAll := append(ascending(101, 140), ascending(201, 205)...)

	t.Run("言語指定: 全ページを繋ぐと重複も欠落もなく全件と一致する", func(t *testing.T) {
		got := collectAllPages(t, exRepo, domain.ExerciseLanguagePhp, 5)
		requireNoDuplicates(t, got)
		require.ElementsMatch(t, wantPHP, got, "ページを跨いだ重複・欠落が無い")
		require.Equal(t, wantPHP, got, "sort_order 同着は id 昇順で解決される")
	})

	t.Run("言語指定: 同じページングを繰り返しても ID 列が毎回同一", func(t *testing.T) {
		first := collectAllPages(t, exRepo, domain.ExerciseLanguagePhp, 5)
		for i := 0; i < 4; i++ {
			require.Equal(t, first, collectAllPages(t, exRepo, domain.ExerciseLanguagePhp, 5))
		}
	})

	t.Run("全言語(language 空): 全ページを繋ぐと重複も欠落もなく全件と一致する", func(t *testing.T) {
		// 言語をまたぐと sort_order の同着は更に増える（言語ごとに 1 から採番されるため）。
		for _, limit := range []int{3, 5, 7} {
			got := collectAllPages(t, exRepo, "", limit)
			requireNoDuplicates(t, got)
			require.ElementsMatch(t, wantAll, got, "limit=%d でページを跨いだ重複・欠落が無い", limit)
			require.Equal(t, wantAll, got, "limit=%d", limit)
		}
	})

	t.Run("全言語(language 空): 同じページングを繰り返しても ID 列が毎回同一", func(t *testing.T) {
		first := collectAllPages(t, exRepo, "", 3)
		for i := 0; i < 4; i++ {
			require.Equal(t, first, collectAllPages(t, exRepo, "", 3))
		}
	})

	t.Run("Limit=0(全件)でも同着は id 昇順で解決される", func(t *testing.T) {
		rows, err := exRepo.ListWithStatusByLanguage(ctx, repository.ListWithStatusInput{Language: ""})
		require.NoError(t, err)
		ids := make([]uint64, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		require.Equal(t, wantAll, ids)
	})

	t.Run("非ページング版 ListByLanguage も同じ順序に揃う", func(t *testing.T) {
		rows, err := exRepo.ListByLanguage(ctx, domain.ExerciseLanguagePhp)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		require.Equal(t, wantPHP, ids)
	})

	t.Run("sort_order が異なる行の並び（仕様）は変わらない", func(t *testing.T) {
		// タイブレークは同着の解決だけを担う。sort_order が違えば従来どおり sort_order が優先される。
		require.NoError(t, db.WithContext(ctx).Model(&domain.MasterExercise{}).
			Where("id = ?", uint64(140)).Update("sort_order", 0).Error)
		defer func() {
			require.NoError(t, db.WithContext(ctx).Model(&domain.MasterExercise{}).
				Where("id = ?", uint64(140)).Update("sort_order", 1).Error)
		}()

		rows, err := exRepo.ListByLanguage(ctx, domain.ExerciseLanguagePhp)
		require.NoError(t, err)
		require.Equal(t, uint64(140), rows[0].ID, "sort_order=0 が id に関係なく先頭に来る")
	})
}

// requireNoDuplicates は ID 列に同じ ID が 2 度現れないことを検証する
// （OFFSET ページングで同着が揺れたときに出る症状そのもの）。
func requireNoDuplicates(t *testing.T, ids []uint64) {
	t.Helper()
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		_, dup := seen[id]
		require.False(t, dup, "ID %d がページを跨いで重複した", id)
		seen[id] = struct{}{}
	}
}

// TestRichDocumentListOrder_TiedUpdatedAt_Integration は updated_at が同値でも
// ListByOwner の並びが常に同じになることを検証する。
func TestRichDocumentListOrder_TiedUpdatedAt_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	userRepo := persistence.NewUserRepository(db)
	repo := persistence.NewRichDocumentRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "rich_documents", "users", "user_oidc_identities")

	owner := &domain.User{Email: "rd-tie@example.com", Role: domain.RoleTrainee}
	require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, owner, domain.OidcProviderCognito, "rd-tie"))

	// ID は UUIDv7 なので作成順に単調増加する。作成順（昇順）で投入し、期待順は id 降順にする。
	created := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		d := &domain.RichDocument{OwnerID: owner.ID, Kind: domain.DocumentKindNote, Title: "tie", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, d))
		created = append(created, d.ID)
	}
	// updated_at を明示的に同値へ揃えて同着を作る。
	require.NoError(t, db.WithContext(ctx).
		Exec(`UPDATE rich_documents SET updated_at = TIMESTAMPTZ '2026-01-01 00:00:00+00'`).Error)

	want := make([]string, len(created))
	copy(want, created)
	sort.Sort(sort.Reverse(sort.StringSlice(want))) // 期待は id 降順

	t.Run("updated_at 同着でも id 降順で解決される", func(t *testing.T) {
		rows, err := repo.ListByOwner(ctx, owner.ID, "")
		require.NoError(t, err)
		ids := make([]string, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		require.Equal(t, want, ids)
	})

	t.Run("繰り返し呼んでも並びが変わらない", func(t *testing.T) {
		var first []string
		for i := 0; i < 5; i++ {
			rows, err := repo.ListByOwner(ctx, owner.ID, domain.DocumentKindNote)
			require.NoError(t, err)
			ids := make([]string, 0, len(rows))
			for _, r := range rows {
				ids = append(ids, r.ID)
			}
			if first == nil {
				first = ids
				continue
			}
			require.Equal(t, first, ids)
		}
	})

	t.Run("updated_at が異なる行の並び（仕様）は変わらない", func(t *testing.T) {
		// 一番古い id の行だけを新しくすると、id 降順に関係なく先頭へ来る。
		oldest := created[0]
		require.NoError(t, db.WithContext(ctx).
			Exec(`UPDATE rich_documents SET updated_at = TIMESTAMPTZ '2026-06-01 00:00:00+00' WHERE id = ?`, oldest).Error)

		rows, err := repo.ListByOwner(ctx, owner.ID, "")
		require.NoError(t, err)
		require.Equal(t, oldest, rows[0].ID)
	})
}

// TestListOrderTieBreaks_Integration は横断調査で見つかった残りの一覧クエリについて、
// ソートキー同着時の並びが一意に決まることを検証する。
func TestListOrderTieBreaks_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()
	// 同着を作るための固定時刻（DB / ホストの時計に依存させない）。
	tie := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("notes: updated_at 同着は id 降順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "notes")
		repo := persistence.NewNoteRepository(db)
		for i := uint64(1); i <= 4; i++ { // id 昇順に投入 → 期待は降順
			require.NoError(t, db.WithContext(ctx).Create(&domain.Note{
				ID: i, UserID: 7, Title: "tie", CreatedAt: tie, UpdatedAt: tie,
			}).Error)
		}
		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		require.Equal(t, []uint64{4, 3, 2, 1}, noteIDs(rows))
	})

	t.Run("companies: name 同着は id 昇順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "companies")
		repo := persistence.NewCompanyRepository(db)
		for _, id := range []uint64{4, 3, 2, 1} { // id 降順に投入 → 期待は昇順
			require.NoError(t, db.WithContext(ctx).Create(&domain.Company{
				ID: id, Name: "同名株式会社", CreatedAt: tie, UpdatedAt: tie,
			}).Error)
		}
		rows, err := repo.ListAll(ctx)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		require.Equal(t, []uint64{1, 2, 3, 4}, ids)
	})

	t.Run("company_applications: created_at 同着は id 降順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "company_applications")
		repo := persistence.NewCompanyApplicationRepository(db)
		for i := uint64(1); i <= 4; i++ {
			require.NoError(t, db.WithContext(ctx).Create(&domain.CompanyApplication{
				ID: i, CompanyName: "c", ApplicantName: "a", Email: "a@example.com",
				Status: domain.CompanyApplicationStatusPending, CreatedAt: tie, UpdatedAt: tie,
			}).Error)
		}
		rows, err := repo.ListAll(ctx)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		require.Equal(t, []uint64{4, 3, 2, 1}, ids)
	})

	t.Run("invitations: created_at 同着は id 降順（一覧・単一取得とも）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "invitations")
		repo := persistence.NewAdminInvitationRepository(db)
		for i := uint64(1); i <= 4; i++ {
			require.NoError(t, db.WithContext(ctx).Create(&domain.AdminInvitation{
				ID: i, CompanyID: 1, Email: "inv@example.com", Role: domain.RoleTrainee,
				Status: domain.InvitationStatusPending, ExpiresAt: tie.Add(24 * time.Hour), CreatedAt: tie,
			}).Error)
		}
		all, err := repo.ListAll(ctx)
		require.NoError(t, err)
		require.Equal(t, []uint64{4, 3, 2, 1}, invitationIDs(all))

		byCompany, err := repo.ListByCompanyID(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, []uint64{4, 3, 2, 1}, invitationIDs(byCompany))

		// 同一 email に pending が複数あっても「どれが受理されるか」がぶれない。
		one, err := repo.FindPendingByEmail(ctx, "inv@example.com")
		require.NoError(t, err)
		require.NotNil(t, one)
		require.Equal(t, uint64(4), one.ID)
	})

	t.Run("ai_chat_sessions: created_at 同着は id 降順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "ai_chat_sessions")
		repo := persistence.NewAiChatSessionRepository(db)
		for i := uint64(1); i <= 4; i++ {
			require.NoError(t, db.WithContext(ctx).Create(&domain.AiChatSession{
				ID: i, UserID: 7, Title: "tie", CreatedAt: tie, UpdatedAt: tie,
			}).Error)
		}
		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		require.Equal(t, []uint64{4, 3, 2, 1}, ids)
	})

	t.Run("learning_reports: period_to 同着は id 降順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "learning_reports")
		repo := persistence.NewLearningReportRepository(db)
		for i := uint64(1); i <= 4; i++ {
			require.NoError(t, db.WithContext(ctx).Create(&domain.LearningReport{
				ID: i, UserID: 7, PeriodFrom: tie.Add(-7 * 24 * time.Hour), PeriodTo: tie,
				Status: domain.LearningReportStatusReady, CreatedAt: tie,
			}).Error)
		}
		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		require.Equal(t, []uint64{4, 3, 2, 1}, ids)
	})

	t.Run("user_chapter_views: last_viewed_at 同着は chapter_id 降順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "user_chapter_views")
		repo := persistence.NewUserChapterViewRepository(db)
		// UpsertView は NOW() で書くので同着を作れない。同値の last_viewed_at を直接投入する。
		for i := uint64(1); i <= 4; i++ {
			require.NoError(t, db.WithContext(ctx).Create(&domain.UserChapterView{
				UserID: 7, TeachingMaterialID: i, CourseID: 10,
				FirstViewedAt: tie, LastViewedAt: tie, ViewCount: 1,
			}).Error)
		}
		rows, err := repo.ListRecentByUser(ctx, 7, 10)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.TeachingMaterialID)
		}
		require.Equal(t, []uint64{4, 3, 2, 1}, ids)

		// LIMIT が同着をまたいでも「どの 1 件か」が固定される。
		top, err := repo.ListRecentByUser(ctx, 7, 1)
		require.NoError(t, err)
		require.Len(t, top, 1)
		require.Equal(t, uint64(4), top[0].TeachingMaterialID)

		last, err := repo.GetLastViewedByUserAndCourse(ctx, 7, 10)
		require.NoError(t, err)
		require.NotNil(t, last)
		require.Equal(t, uint64(4), last.TeachingMaterialID)
	})

	t.Run("user_chapter_progress: ORDER BY 無しにせず id 昇順で固定する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "user_chapter_progress")
		repo := persistence.NewLessonProgressRepository(db)
		for _, id := range []uint64{4, 3, 2, 1} { // id 降順に投入 → 期待は昇順
			require.NoError(t, db.WithContext(ctx).Create(&domain.UserLessonProgress{
				ID: id, UserID: 7, TeachingMaterialID: id, CourseID: 10, CompletedAt: tie, CreatedAt: tie,
			}).Error)
		}
		rows, err := repo.ListByUser(ctx, 7)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		require.Equal(t, []uint64{1, 2, 3, 4}, ids)
	})
}

func noteIDs(rows []domain.Note) []uint64 {
	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	return ids
}

func invitationIDs(rows []domain.AdminInvitation) []uint64 {
	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	return ids
}
