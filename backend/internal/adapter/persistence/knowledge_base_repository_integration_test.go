//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kbUseCases は結合テストで使う usecase 一式（実 repository を注入した状態）。
type kbUseCases struct {
	create    *usecase.CreatePageUseCase
	get       *usecase.GetPageUseCase
	tree      *usecase.GetPageTreeUseCase
	rename    *usecase.RenamePageUseCase
	move      *usecase.MovePageUseCase
	archive   *usecase.ArchivePageUseCase
	unarchive *usecase.UnarchivePageUseCase
	replace   *usecase.ReplacePageBlocksUseCase
}

func newKbUseCases(repo repository.KnowledgeBaseRepository) kbUseCases {
	return kbUseCases{
		create:    usecase.NewCreatePageUseCase(repo),
		get:       usecase.NewGetPageUseCase(repo),
		tree:      usecase.NewGetPageTreeUseCase(repo),
		rename:    usecase.NewRenamePageUseCase(repo),
		move:      usecase.NewMovePageUseCase(repo),
		archive:   usecase.NewArchivePageUseCase(repo),
		unarchive: usecase.NewUnarchivePageUseCase(repo),
		replace:   usecase.NewReplacePageBlocksUseCase(repo),
	}
}

// mustCreatePage は usecase 経由でページを 1 枚作る（closure も張られる）。
func mustCreatePage(ctx context.Context, t *testing.T, uc kbUseCases, ws, space string, parentID *string, title string) *domain.Page {
	t.Helper()
	page, err := uc.create.Execute(ctx, usecase.CreatePageInput{
		WorkspaceID: ws, SpaceID: space, ParentID: parentID, Title: title, CreatedByUserID: 1,
	})
	require.NoError(t, err)
	return page
}

// queryPagePaths は page_paths の全行を "page→ancestor" → depth で返す（closure の全行検証用）。
func queryPagePaths(t *testing.T, db *sql.DB, workspaceID string) map[string]int {
	t.Helper()
	rows, err := db.Query(
		`SELECT page_id::text, ancestor_id::text, depth FROM page_paths WHERE workspace_id = $1`, workspaceID,
	)
	require.NoError(t, err)
	defer rows.Close()
	got := map[string]int{}
	for rows.Next() {
		var pageID, ancestorID string
		var depth int
		require.NoError(t, rows.Scan(&pageID, &ancestorID, &depth))
		got[pageID+"→"+ancestorID] = depth
	}
	require.NoError(t, rows.Err())
	return got
}

// treeShape はページツリーを "title(子, 子, ...)" の文字列に落とす（木の形の比較用）。
func treeShape(nodes []*usecase.PageTreeNode) string {
	s := ""
	for i, n := range nodes {
		if i > 0 {
			s += ", "
		}
		s += n.Page.Title
		if len(n.Children) > 0 {
			s += "(" + treeShape(n.Children) + ")"
		}
	}
	return s
}

func TestKnowledgeBasePageUseCases_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewKnowledgeBaseRepository(sqlDB)
	uc := newKbUseCases(repo)
	ctx := context.Background()

	// setup は各サブテストの冒頭で呼ぶ共通初期化（ワークスペース + スペース 2 つ）。
	setup := func(t *testing.T) (ws, spaceA, spaceB string) {
		t.Helper()
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws = createWorkspace(t, sqlDB, "ws-main")
		spaceA = createSpace(t, sqlDB, ws, "aaa")
		spaceB = createSpace(t, sqlDB, ws, "bbb")
		return ws, spaceA, spaceB
	}

	t.Run("祖先IDが根から順に返る（パンくずの骨組み）", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "根")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &root.ID, "子")
		grand := mustCreatePage(ctx, t, uc, ws, spaceA, &child.ID, "孫")

		got, err := repo.ListAncestorPageIDs(ctx, ws, grand.ID)
		require.NoError(t, err)
		assert.Equal(t, []string{root.ID, child.ID}, got, "根 → 親 の順（自分は含まない）")

		// 根ページ・実在しない ID は空（エラーにしない）。
		empty, err := repo.ListAncestorPageIDs(ctx, ws, root.ID)
		require.NoError(t, err)
		assert.Empty(t, empty)
		none, err := repo.ListAncestorPageIDs(ctx, ws, "not-a-uuid")
		require.NoError(t, err)
		assert.Empty(t, none)
	})

	t.Run("削除は子孫・closure・本文ごとCASCADEで消える", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "消す根")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &root.ID, "消える子")
		_ = mustCreatePage(ctx, t, uc, ws, spaceA, nil, "残る根")
		_, err := uc.replace.Execute(ctx, usecase.ReplacePageBlocksInput{
			WorkspaceID: ws, PageID: child.ID,
			Doc: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"本文"}]}]}`,
		})
		require.NoError(t, err)

		require.NoError(t, repo.DeletePageSubtree(ctx, ws, root.ID))

		_, err = uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: ws, PageID: child.ID})
		require.ErrorIs(t, err, repository.ErrPageNotFound, "子孫も一緒に消える")
		tree, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "残る根", treeShape(tree), "残す根は無傷で、消した木は形から消える")

		// 派生テーブルにも残骸が無い（CASCADE の確認）。
		var count int
		require.NoError(t, sqlDB.QueryRowContext(ctx,
			`SELECT count(*) FROM page_paths WHERE page_id = $1 OR ancestor_id = $1`, child.ID).Scan(&count))
		assert.Zero(t, count)
		require.NoError(t, sqlDB.QueryRowContext(ctx,
			`SELECT count(*) FROM blocks WHERE page_id = $1`, child.ID).Scan(&count))
		assert.Zero(t, count)
		require.NoError(t, sqlDB.QueryRowContext(ctx,
			`SELECT count(*) FROM page_snapshots WHERE page_id = $1`, child.ID).Scan(&count))
		assert.Zero(t, count)

		// 実在しないページの削除は ErrPageNotFound（冪等にしない — 押した相手が
		// 「もう無い」ことを知れる）。
		require.ErrorIs(t, repo.DeletePageSubtree(ctx, ws, root.ID), repository.ErrPageNotFound)
	})

	t.Run("作成して取得すると木の形とclosureが正しい", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root1 := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root1")
		root2 := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root2")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &root1.ID, "child")
		grand := mustCreatePage(ctx, t, uc, ws, spaceA, &child.ID, "grand")

		got, err := uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: ws, PageID: child.ID})
		require.NoError(t, err)
		assert.Equal(t, "child", got.Page.Title)
		assert.Equal(t, &root1.ID, got.Page.ParentID)
		assert.JSONEq(t, `{"type":"doc","content":[]}`, got.Doc, "未保存ページの本文は空 doc")

		tree, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "root1(child(grand)), root2", treeShape(tree))

		assert.Equal(t, map[string]int{
			root1.ID + "→" + root1.ID: 0,
			root2.ID + "→" + root2.ID: 0,
			child.ID + "→" + child.ID: 0,
			child.ID + "→" + root1.ID: 1,
			grand.ID + "→" + grand.ID: 0,
			grand.ID + "→" + child.ID: 1,
			grand.ID + "→" + root1.ID: 2,
		}, queryPagePaths(t, sqlDB, ws), "closure は自分自身 depth=0 + 全祖先")
	})

	t.Run("同一スペース内の移動でclosureが付け替わる", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root1 := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root1")
		root2 := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root2")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &root1.ID, "child")
		grand := mustCreatePage(ctx, t, uc, ws, spaceA, &child.ID, "grand")

		moved, err := uc.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: ws, PageID: child.ID, NewParentID: &root2.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, &root2.ID, moved.ParentID)
		assert.Equal(t, spaceA, moved.SpaceID)

		tree, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "root1, root2(child(grand))", treeShape(tree))

		assert.Equal(t, map[string]int{
			root1.ID + "→" + root1.ID: 0,
			root2.ID + "→" + root2.ID: 0,
			child.ID + "→" + child.ID: 0,
			child.ID + "→" + root2.ID: 1,
			grand.ID + "→" + grand.ID: 0,
			grand.ID + "→" + child.ID: 1,
			grand.ID + "→" + root2.ID: 2,
		}, queryPagePaths(t, sqlDB, ws), "旧祖先 root1 との組が消え、新祖先 root2 との組に置き換わる")
	})

	t.Run("自分の子孫への移動は拒否される", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &root.ID, "child")
		grand := mustCreatePage(ctx, t, uc, ws, spaceA, &child.ID, "grand")

		_, err := uc.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: ws, PageID: root.ID, NewParentID: &grand.ID,
		})
		require.ErrorIs(t, err, usecase.ErrPageCycle)

		_, err = uc.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: ws, PageID: root.ID, NewParentID: &root.ID,
		})
		require.ErrorIs(t, err, usecase.ErrPageCycle, "自分自身も拒否")

		// 木が壊れていないこと。
		tree, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "root(child(grand))", treeShape(tree))
	})

	t.Run("スペースをまたぐ移動で子孫のspace_idも変わる", func(t *testing.T) {
		ws, spaceA, spaceB := setup(t)
		rootA := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "rootA")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &rootA.ID, "child")
		grand := mustCreatePage(ctx, t, uc, ws, spaceA, &child.ID, "grand")
		rootB := mustCreatePage(ctx, t, uc, ws, spaceB, nil, "rootB")

		// child（+ grand）を spaceB の rootB の下へ。
		moved, err := uc.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: ws, PageID: child.ID, NewParentID: &rootB.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, spaceB, moved.SpaceID)

		grandAfter, err := uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: ws, PageID: grand.ID})
		require.NoError(t, err)
		assert.Equal(t, spaceB, grandAfter.Page.SpaceID, "子孫の space_id も一括で変わる")

		treeA, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "rootA", treeShape(treeA))
		treeB, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceB})
		require.NoError(t, err)
		assert.Equal(t, "rootB(child(grand))", treeShape(treeB))

		assert.Equal(t, map[string]int{
			rootA.ID + "→" + rootA.ID: 0,
			rootB.ID + "→" + rootB.ID: 0,
			child.ID + "→" + child.ID: 0,
			child.ID + "→" + rootB.ID: 1,
			grand.ID + "→" + grand.ID: 0,
			grand.ID + "→" + child.ID: 1,
			grand.ID + "→" + rootB.ID: 2,
		}, queryPagePaths(t, sqlDB, ws))
	})

	t.Run("別スペースのルートへの移動もできる", func(t *testing.T) {
		ws, spaceA, spaceB := setup(t)
		rootA := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "rootA")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &rootA.ID, "child")

		moved, err := uc.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: ws, PageID: child.ID, NewSpaceID: spaceB,
		})
		require.NoError(t, err)
		assert.Nil(t, moved.ParentID)
		assert.Equal(t, spaceB, moved.SpaceID)

		assert.Equal(t, map[string]int{
			rootA.ID + "→" + rootA.ID: 0,
			child.ID + "→" + child.ID: 0,
		}, queryPagePaths(t, sqlDB, ws), "ルートへ出たので祖先との組は自分自身だけ")
	})

	t.Run("アーカイブでツリーから消え復帰で戻る", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &root.ID, "child")
		keep := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "keep")

		require.NoError(t, uc.archive.Execute(ctx, usecase.ArchivePageInput{WorkspaceID: ws, PageID: root.ID}))

		tree, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "keep", treeShape(tree), "サブツリーごと消える")

		childAfter, err := uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: ws, PageID: child.ID})
		require.NoError(t, err)
		assert.NotNil(t, childAfter.Page.ArchivedAt, "子孫もアーカイブされる")

		restored, err := uc.unarchive.Execute(ctx, usecase.UnarchivePageInput{WorkspaceID: ws, PageID: root.ID})
		require.NoError(t, err)
		assert.Nil(t, restored.ArchivedAt)

		tree, err = uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "root(child), keep", treeShape(tree), "サブツリーごと戻る")
		_ = keep
	})

	t.Run("復帰時にpositionが衝突したら末尾へ再採番される", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		first := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "first") // position a0

		require.NoError(t, uc.archive.Execute(ctx, usecase.ArchivePageInput{WorkspaceID: ws, PageID: first.ID}))
		// アーカイブ中は現役の兄弟がいないので、新しいページが同じ position a0 を取る。
		second := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "second")
		assert.Equal(t, first.Position, second.Position, "前提: 部分 UNIQUE は現役だけを守るので同じ position になる")

		restored, err := uc.unarchive.Execute(ctx, usecase.UnarchivePageInput{WorkspaceID: ws, PageID: first.ID})
		require.NoError(t, err)
		assert.Nil(t, restored.ArchivedAt)
		assert.Greater(t, restored.Position, second.Position, "衝突を検出して末尾へ再採番")

		tree, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "second, first", treeShape(tree))
	})

	t.Run("復帰は根と同時にアーカイブされた一括分だけを戻す", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root")
		early := mustCreatePage(ctx, t, uc, ws, spaceA, &root.ID, "early")
		late := mustCreatePage(ctx, t, uc, ws, spaceA, &root.ID, "late")

		// early を先に単独アーカイブ → その後 root ごとアーカイブ。
		require.NoError(t, uc.archive.Execute(ctx, usecase.ArchivePageInput{WorkspaceID: ws, PageID: early.ID}))
		require.NoError(t, uc.archive.Execute(ctx, usecase.ArchivePageInput{WorkspaceID: ws, PageID: root.ID}))

		_, err := uc.unarchive.Execute(ctx, usecase.UnarchivePageInput{WorkspaceID: ws, PageID: root.ID})
		require.NoError(t, err)

		tree, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		assert.Equal(t, "root(late)", treeShape(tree), "先に単独アーカイブした early は戻らない")
		_ = late
	})

	t.Run("親がアーカイブ中のままでは復帰できない", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root")
		child := mustCreatePage(ctx, t, uc, ws, spaceA, &root.ID, "child")
		require.NoError(t, uc.archive.Execute(ctx, usecase.ArchivePageInput{WorkspaceID: ws, PageID: root.ID}))

		_, err := uc.unarchive.Execute(ctx, usecase.UnarchivePageInput{WorkspaceID: ws, PageID: child.ID})
		require.ErrorIs(t, err, usecase.ErrPageParentArchived)
	})

	t.Run("アーカイブ済みの親の下には作成も移動もできない", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root")
		other := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "other")
		require.NoError(t, uc.archive.Execute(ctx, usecase.ArchivePageInput{WorkspaceID: ws, PageID: root.ID}))

		_, err := uc.create.Execute(ctx, usecase.CreatePageInput{
			WorkspaceID: ws, SpaceID: spaceA, ParentID: &root.ID, Title: "x", CreatedByUserID: 1,
		})
		require.ErrorIs(t, err, usecase.ErrPageParentArchived)

		_, err = uc.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: ws, PageID: other.ID, NewParentID: &root.ID,
		})
		require.ErrorIs(t, err, usecase.ErrPageParentArchived)
	})

	t.Run("ブロック書き換えと取得の往復とsnapshot更新", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		page := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "doc-page")

		doc1 := `{"type":"doc","content":[
			{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"タイトル"}]},
			{"type":"bulletList","content":[
				{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"bold"}],"text":"太字"}]}]}
			]}
		]}`
		snap1, err := uc.replace.Execute(ctx, usecase.ReplacePageBlocksInput{WorkspaceID: ws, PageID: page.ID, Doc: doc1})
		require.NoError(t, err)
		assert.JSONEq(t, doc1, snap1.Doc)

		got, err := uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: ws, PageID: page.ID})
		require.NoError(t, err)
		assert.JSONEq(t, doc1, got.Doc, "保存した doc と取得した doc が同値")

		// snapshot を消しても blocks から同じ doc が組み上がる（正本は blocks 側）。
		_, err = sqlDB.Exec(`DELETE FROM page_snapshots WHERE page_id = $1`, page.ID)
		require.NoError(t, err)
		got, err = uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: ws, PageID: page.ID})
		require.NoError(t, err)
		assert.JSONEq(t, doc1, got.Doc, "blocks からの組み立てでも同値")

		// 書き換えると blocks / snapshot が置き換わる。
		doc2 := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"書き換え後"}]}]}`
		snap2, err := uc.replace.Execute(ctx, usecase.ReplacePageBlocksInput{WorkspaceID: ws, PageID: page.ID, Doc: doc2})
		require.NoError(t, err)
		assert.JSONEq(t, doc2, snap2.Doc, "snapshot が焼き直される")

		var blockCount int
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM blocks WHERE page_id = $1`, page.ID).Scan(&blockCount))
		assert.Equal(t, 1, blockCount, "全消し全入れで旧行が残らない")

		// 空 doc で全消しできる。
		empty := `{"type":"doc","content":[]}`
		snap3, err := uc.replace.Execute(ctx, usecase.ReplacePageBlocksInput{WorkspaceID: ws, PageID: page.ID, Doc: empty})
		require.NoError(t, err)
		assert.JSONEq(t, empty, snap3.Doc)
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM blocks WHERE page_id = $1`, page.ID).Scan(&blockCount))
		assert.Equal(t, 0, blockCount)
	})

	t.Run("改名できる", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		page := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "旧名")
		renamed, err := uc.rename.Execute(ctx, usecase.RenamePageInput{WorkspaceID: ws, PageID: page.ID, Title: "新名"})
		require.NoError(t, err)
		assert.Equal(t, "新名", renamed.Title)
		assert.True(t, renamed.UpdatedAt.After(page.UpdatedAt) || renamed.UpdatedAt.Equal(page.UpdatedAt))
	})

	t.Run("別ワークスペースのページには全usecaseで手が届かない", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		wsOther := createWorkspace(t, sqlDB, "ws-other")
		spaceOther := createSpace(t, sqlDB, wsOther, "other")
		victim := mustCreatePage(ctx, t, uc, wsOther, spaceOther, nil, "victim")
		mine := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "mine")

		// 読み: 取得・ツリー。
		_, err := uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: ws, PageID: victim.ID})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceOther})
		require.ErrorIs(t, err, repository.ErrSpaceNotFound)

		// 書き: 作成（親／スペース越え）・改名・移動・アーカイブ・復帰・本文書き換え。
		_, err = uc.create.Execute(ctx, usecase.CreatePageInput{
			WorkspaceID: ws, SpaceID: spaceOther, Title: "x", CreatedByUserID: 1,
		})
		require.ErrorIs(t, err, repository.ErrSpaceNotFound)
		_, err = uc.create.Execute(ctx, usecase.CreatePageInput{
			WorkspaceID: ws, SpaceID: spaceA, ParentID: &victim.ID, Title: "x", CreatedByUserID: 1,
		})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = uc.rename.Execute(ctx, usecase.RenamePageInput{WorkspaceID: ws, PageID: victim.ID, Title: "乗っ取り"})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = uc.move.Execute(ctx, usecase.MovePageInput{WorkspaceID: ws, PageID: victim.ID})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = uc.move.Execute(ctx, usecase.MovePageInput{WorkspaceID: ws, PageID: mine.ID, NewParentID: &victim.ID})
		require.ErrorIs(t, err, repository.ErrPageNotFound, "別テナントのページを親にもできない")
		err = uc.archive.Execute(ctx, usecase.ArchivePageInput{WorkspaceID: ws, PageID: victim.ID})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = uc.unarchive.Execute(ctx, usecase.UnarchivePageInput{WorkspaceID: ws, PageID: victim.ID})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = uc.replace.Execute(ctx, usecase.ReplacePageBlocksInput{
			WorkspaceID: ws, PageID: victim.ID, Doc: `{"type":"doc","content":[]}`,
		})
		require.ErrorIs(t, err, repository.ErrPageNotFound)

		// 相手のページが無傷であること。
		after, err := uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: wsOther, PageID: victim.ID})
		require.NoError(t, err)
		assert.Equal(t, "victim", after.Page.Title)
		assert.Nil(t, after.Page.ArchivedAt)
	})

	t.Run("ワークスペースとスペースの存在確認", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		found, err := repo.FindWorkspaceByID(ctx, ws)
		require.NoError(t, err)
		assert.Equal(t, "ws-main", found.Slug)
		sp, err := repo.FindSpace(ctx, ws, spaceA)
		require.NoError(t, err)
		assert.Equal(t, "aaa", sp.Key)

		_, err = repo.FindWorkspaceByID(ctx, newID())
		require.ErrorIs(t, err, repository.ErrWorkspaceNotFound)
		_, err = repo.FindSpace(ctx, ws, newID())
		require.ErrorIs(t, err, repository.ErrSpaceNotFound)
	})

	// URL 由来の生文字列がそのまま来る想定の入口検証。UUID として不正な ID は
	// DB エラーではなく「存在しない」と同じ結果に落ちること。
	t.Run("不正な形式のIDは存在しない扱いになる", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		bad := "not-a-uuid"

		_, err := repo.FindWorkspaceByID(ctx, bad)
		require.ErrorIs(t, err, repository.ErrWorkspaceNotFound)
		_, err = repo.FindSpace(ctx, ws, bad)
		require.ErrorIs(t, err, repository.ErrSpaceNotFound)
		_, err = repo.FindPage(ctx, ws, bad)
		require.ErrorIs(t, err, repository.ErrPageNotFound)

		pages, err := repo.ListActivePagesBySpace(ctx, ws, bad)
		require.NoError(t, err)
		assert.Empty(t, pages)
		pos, err := repo.LastActiveSiblingPosition(ctx, ws, bad, nil)
		require.NoError(t, err)
		assert.Empty(t, pos)
		conflicted, err := repo.HasActiveSiblingPosition(ctx, ws, spaceA, &bad, "a0", bad)
		require.NoError(t, err)
		assert.False(t, conflicted)
		isDesc, err := repo.HasDescendant(ctx, ws, bad, bad)
		require.NoError(t, err)
		assert.False(t, isDesc)

		err = repo.CreatePage(ctx, &domain.Page{WorkspaceID: ws, SpaceID: bad, Position: "a0", CreatedByUserID: 1})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = repo.UpdatePageTitle(ctx, ws, bad, "x")
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		err = repo.MovePage(ctx, ws, bad, nil, spaceA, "a0")
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		err = repo.ArchivePageSubtree(ctx, ws, bad)
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		err = repo.UnarchivePageSubtree(ctx, ws, bad, time.Now(), nil)
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		blocks, err := repo.ListBlocksByPage(ctx, ws, bad)
		require.NoError(t, err)
		assert.Empty(t, blocks)
		err = repo.ReplacePageBlocks(ctx, ws, bad, nil, `{"type":"doc","content":[]}`)
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = repo.GetPageSnapshot(ctx, ws, bad)
		require.ErrorIs(t, err, repository.ErrPageSnapshotNotFound)
	})

	t.Run("存在しないページの改名と移動はErrPageNotFound", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		_, err := repo.UpdatePageTitle(ctx, ws, newID(), "x")
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		err = repo.MovePage(ctx, ws, newID(), nil, spaceA, "a0")
		require.ErrorIs(t, err, repository.ErrPageNotFound)
	})

	t.Run("文書順が壊れたBlockWriteは保存を拒否する", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		page := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "broken-rows")
		// ParentIndex が自分より後 = 「親が先」の前提違反。
		err := repo.ReplacePageBlocks(ctx, ws, page.ID, []repository.BlockWrite{
			{ParentIndex: 1, Position: "a0", Type: domain.BlockTypeListItem, Attrs: "{}"},
			{ParentIndex: -1, Position: "a0", Type: domain.BlockTypeBulletList, Attrs: "{}"},
		}, `{"type":"doc","content":[]}`)
		require.Error(t, err)
		var blockCount int
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM blocks WHERE page_id = $1`, page.ID).Scan(&blockCount))
		assert.Equal(t, 0, blockCount, "途中まで書いた行がロールバックで残らない")
	})

	t.Run("大きめの木でも作成と取得が破綻しない", func(t *testing.T) {
		ws, spaceA, _ := setup(t)
		root := mustCreatePage(ctx, t, uc, ws, spaceA, nil, "root")
		parent := root.ID
		// 深さ 5 × 各 3 兄弟の木。
		for depth := 0; depth < 5; depth++ {
			var next string
			for i := 0; i < 3; i++ {
				p := mustCreatePage(ctx, t, uc, ws, spaceA, &parent, fmt.Sprintf("d%d-%d", depth, i))
				next = p.ID
			}
			parent = next
		}
		tree, err := uc.tree.Execute(ctx, usecase.GetPageTreeInput{WorkspaceID: ws, SpaceID: spaceA})
		require.NoError(t, err)
		require.Len(t, tree, 1)
		var count func(nodes []*usecase.PageTreeNode) int
		count = func(nodes []*usecase.PageTreeNode) int {
			n := len(nodes)
			for _, node := range nodes {
				n += count(node.Children)
			}
			return n
		}
		assert.Equal(t, 16, count(tree))
	})
}

// TestKnowledgeBaseSimpleProtocol_Integration は simple query protocol（本番の
// transaction pooler と同じ経路）で blocks.inline（NULL 可 jsonb）の INSERT / SELECT が
// 通ることを固定する回帰テスト。extended protocol では型の取り違えが OID で救われてしまい、
// ローカル / CI の既定接続では原理的に検出できない（段 1-a で確定した欠陥の再発防止）。
func TestKnowledgeBaseSimpleProtocol_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDBSimpleProtocol(t)
	repo := persistence.NewKnowledgeBaseRepository(sqlDB)
	uc := newKbUseCases(repo)
	ctx := context.Background()

	testsupport.TruncateAll(t, sqlDB, kbTables...)
	ws := createWorkspace(t, sqlDB, "ws-simple")
	space := createSpace(t, sqlDB, ws, "eng")
	page := mustCreatePage(ctx, t, uc, ws, space, nil, "simple-protocol")

	// inline あり（葉）と inline NULL（容器・区切り線）を両方通す。
	doc := `{"type":"doc","content":[
		{"type":"paragraph","content":[{"type":"text","text":"simple protocol 経由"}]},
		{"type":"horizontalRule"},
		{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"項目"}]}]}]}
	]}`
	snap, err := uc.replace.Execute(ctx, usecase.ReplacePageBlocksInput{WorkspaceID: ws, PageID: page.ID, Doc: doc})
	require.NoError(t, err)
	assert.JSONEq(t, doc, snap.Doc)

	got, err := uc.get.Execute(ctx, usecase.GetPageInput{WorkspaceID: ws, PageID: page.ID})
	require.NoError(t, err)
	assert.JSONEq(t, doc, got.Doc)

	// 行レベルでも inline の NULL / 非 NULL が意図どおり保存されていること。
	blocks, err := repo.ListBlocksByPage(ctx, ws, page.ID)
	require.NoError(t, err)
	require.Len(t, blocks, 5)
	byType := map[domain.BlockType]*domain.Block{}
	for i := range blocks {
		byType[blocks[i].Type] = &blocks[i]
	}
	require.NotNil(t, byType[domain.BlockTypeParagraph].Inline, "葉ノードは inline を持つ")
	require.Nil(t, byType[domain.BlockTypeHorizontalRule].Inline, "content の無いノードは inline NULL")
	require.Nil(t, byType[domain.BlockTypeBulletList].Inline, "容器ノードは inline NULL")
}

// TestKnowledgeBaseDeleteWorkspace_Integration は DeleteWorkspace が人の居るワークスペースを
// 守り、それ以外は配下ごと消すことを実 PostgreSQL で固定する。
func TestKnowledgeBaseDeleteWorkspace_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	repo := persistence.NewKnowledgeBaseRepository(sqlDB)
	truncTables := append([]string{"users"}, kbTables...)

	t.Run("所属している人がいるワークスペースは消さない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, truncTables...)
		ws := createWorkspace(t, sqlDB, "ws-with-members")
		member := createUser(t, sqlDB, "member")
		_, err := sqlDB.Exec(`UPDATE users SET workspace_id = $1 WHERE id = $2`, ws, member)
		require.NoError(t, err)

		err = repo.DeleteWorkspace(ctx, ws)
		assert.ErrorIs(t, err, repository.ErrWorkspaceHasMembers)

		var count int
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM workspaces WHERE id = $1`, ws).Scan(&count))
		assert.Equal(t, 1, count, "人の居るワークスペースは残っていなければならない")
	})

	t.Run("誰も所属していないワークスペースは配下ごと消える", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, truncTables...)
		ws := createWorkspace(t, sqlDB, "ws-personal")
		createSpace(t, sqlDB, ws, "eng")

		require.NoError(t, repo.DeleteWorkspace(ctx, ws))

		var wsCount, spaceCount int
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM workspaces WHERE id = $1`, ws).Scan(&wsCount))
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM spaces WHERE workspace_id = $1`, ws).Scan(&spaceCount))
		assert.Zero(t, wsCount)
		assert.Zero(t, spaceCount, "配下は FK CASCADE で一緒に消える")
	})

	t.Run("存在しないワークスペースはErrWorkspaceNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, truncTables...)
		err := repo.DeleteWorkspace(ctx, newID())
		assert.ErrorIs(t, err, repository.ErrWorkspaceNotFound)
	})
}
