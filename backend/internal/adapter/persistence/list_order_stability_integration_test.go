//go:build integration

package persistence_test

import (
	"context"
	"strconv"
	"testing"

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
	// limit <= 0 だと offset が進まず、ListWithStatusByLanguage も LIMIT/OFFSET を付けないため
	// 毎回全件が返って終了条件（len(rows) < limit）も成立しない。放置するとテストが
	// タイムアウトまでハングし、原因の分かりにくい CI ハングになるので入口で落とす。
	require.Positive(t, limit, "collectAllPages は limit > 0 を前提とする")
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
	sqlDB := testsupport.OpenTestDB(t)
	exRepo := persistence.NewMasterExerciseRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "master_exercises", "exercise_submissions")

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
			insertMasterExercise(ctx, t, sqlDB, &row)
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
		setSortOrder := func(sortOrder int) {
			_, err := sqlDB.ExecContext(ctx,
				`UPDATE master_exercises SET sort_order = $1, updated_at = now() WHERE id = $2`,
				sortOrder, uint64(140))
			require.NoError(t, err)
		}
		setSortOrder(0)
		defer setSortOrder(1)

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
