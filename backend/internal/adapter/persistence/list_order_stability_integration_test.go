//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// 一覧クエリの並びが「ソートキーの同着」で崩れないことを横断的に固定する。
//
// 非一意な列だけで ORDER BY すると同値行の相対順序は SQL 上未定義で、実行計画・ページ境界・
// 物理配置で変わりうる。ここでは blocks.position（兄弟内の並び順）を題材にする —
// ListBlocksByPage のコメントが明言するとおり、親の異なるブロック同士は position が
// 偶然一致しうる（兄弟内でしか一意性を強制していないため）ので、id がそのタイブレーク。
//
// タイブレークを外すと素の走査順がそのまま返ってしまうことを確かめるため、各同着ペアは
// 「投入順（＝ヒープの物理順）」を「期待順（id 昇順）」の逆にしてある。

// blockRow は insertBlock 呼び出し 1 回分。
type blockRow struct {
	id       string
	parentID *string
	position string
}

// insertBlocks は rows を渡した順（＝物理投入順）でそのまま INSERT する。
func insertBlocks(t *testing.T, db *sql.DB, workspaceID, pageID string, rows []blockRow) {
	t.Helper()
	for _, r := range rows {
		require.NoError(t, insertBlock(
			db, r.id, workspaceID, pageID, r.parentID, r.position, domain.BlockTypeParagraph, "{}", nil,
		))
	}
}

// TestListBlocksByPage_TiedPosition_Integration は position 同着でも
// ListBlocksByPage の並びが id 昇順で決定的に解決されることを検証する。
func TestListBlocksByPage_TiedPosition_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "blocks", "pages", "spaces", "workspaces")

	ws := createWorkspace(t, sqlDB, "ws-list-order")
	space := createSpace(t, sqlDB, ws, "eng")
	page := createPage(t, sqlDB, ws, space, nil, "a0")

	// 親 2 つ（トップレベル）。position は a0 < a1 で確定するが、id はわざと逆
	// （parentA の id を parentB より大きくする）にして、「position が違えば id の
	// 大小に関係なく position が並びを決める」ことを後段の assertion で確かめる。
	parentA := "00000000-0000-7000-8000-0000000000f9"
	parentB := "00000000-0000-7000-8000-0000000000f1"

	// 子は 2 つの親それぞれの下に、同じ position（b0/b1/b2）で 1 件ずつ置く。
	// 同じ親の中では position は一意（uq_blocks_parent_position）だが、親が違えば
	// 偶然の一致が起こりうる — その一致が全ブロック一覧では同着になる。
	childA0, childA1, childA2 := "00000000-0000-7000-8000-000000000006", "00000000-0000-7000-8000-000000000005", "00000000-0000-7000-8000-000000000004"
	childB0, childB1, childB2 := "00000000-0000-7000-8000-000000000003", "00000000-0000-7000-8000-000000000002", "00000000-0000-7000-8000-000000000001"

	insertBlocks(t, sqlDB, ws, page, []blockRow{
		{id: parentA, position: "a0"},
		{id: parentB, position: "a1"},
		// 各同着ペアは「id が大きい方（=期待順で後ろ）」を先に投入する。
		// タイブレークが無ければ物理投入順がそのまま出て、期待順（id 昇順）と食い違う。
		{id: childA0, parentID: &parentA, position: "b0"},
		{id: childB0, parentID: &parentB, position: "b0"},
		{id: childA1, parentID: &parentA, position: "b1"},
		{id: childB1, parentID: &parentB, position: "b1"},
		{id: childA2, parentID: &parentA, position: "b2"},
		{id: childB2, parentID: &parentB, position: "b2"},
	})

	allIDs := []string{parentA, parentB, childA0, childB0, childA1, childB1, childA2, childB2}
	wantOrder := []string{parentA, parentB, childB0, childA0, childB1, childA1, childB2, childA2}

	list := func() []string {
		rows, err := persistence.NewKnowledgeBaseRepository(sqlDB).ListBlocksByPage(ctx, ws, page)
		require.NoError(t, err)
		ids := make([]string, 0, len(rows))
		for _, b := range rows {
			ids = append(ids, b.ID)
		}
		return ids
	}

	t.Run("同着は id 昇順で解決される", func(t *testing.T) {
		got := list()
		require.Equal(t, wantOrder, got)
	})

	t.Run("position が異なる行は id の大小に関係なく position の順を保つ", func(t *testing.T) {
		got := list()
		posOfA := indexOf(got, parentA)
		posOfB := indexOf(got, parentB)
		require.Less(t, posOfA, posOfB,
			"parentA の id は parentB より大きいが、position(a0<a1) が優先されて先に来ること")
	})

	t.Run("同じクエリを繰り返しても並びは毎回同一", func(t *testing.T) {
		first := list()
		for i := 0; i < 4; i++ {
			require.Equal(t, first, list())
		}
	})

	t.Run("重複も欠落もなく全件と一致する", func(t *testing.T) {
		got := list()
		requireNoDuplicates(t, got)
		require.ElementsMatch(t, allIDs, got)
	})
}

// indexOf は s の中で最初に v と一致する位置を返す（無ければ -1）。
func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

// requireNoDuplicates は ID 列に同じ ID が 2 度現れないことを検証する
// （並びの解決が同着で不安定になったときに出る症状そのもの）。
func requireNoDuplicates(t *testing.T, ids []string) {
	t.Helper()
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		_, dup := seen[id]
		require.False(t, dup, "ID %s がページを跨いで重複した", id)
		seen[id] = struct{}{}
	}
}
