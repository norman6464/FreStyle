//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/kb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FRESTYLE-434 段 4（本文検索と逆リンク）の結合テスト。
//
// page_search / page_links は blocks / pages.title から作り直せる派生データ
// （schema.hcl のコメント参照）なので、ここでは「保存のたびに正しく張り替わるか」
// 「CASCADE で正しく消えるか」「再構築が冪等か」を実 PostgreSQL で固定する。

// queryPageSearchRow は page_search を 1 行だけ読む（無ければ found=false）。
func queryPageSearchRow(t *testing.T, db *sql.DB, pageID string) (title, body string, found bool) {
	t.Helper()
	err := db.QueryRow(`SELECT title, body FROM page_search WHERE page_id = $1`, pageID).Scan(&title, &body)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", false
	}
	require.NoError(t, err)
	return title, body, true
}

// countPageLinksForPage はページ 1 枚が持つ page_links の行数（そのページ配下の
// ブロックが source_block_id になっている行）を数える。
func countPageLinksForPage(t *testing.T, db *sql.DB, workspaceID, pageID string) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(`
		SELECT count(*) FROM page_links pl
		JOIN blocks b ON b.id = pl.source_block_id
		WHERE b.workspace_id = $1 AND b.page_id = $2
	`, workspaceID, pageID).Scan(&count))
	return count
}

// pageRefDoc は 1 段落・1 pageRef だけの最小 doc を組み立てる（本文テキストと参照先の
// 両方を持たせたいテストのための小道具）。
func pageRefDoc(text, targetPageID string) string {
	return fmt.Sprintf(
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":%q},{"type":"pageRef","attrs":{"pageId":%q}}]}]}`,
		text, targetPageID,
	)
}

func TestKnowledgeBasePageSearchAndLinksWrite_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	uc := newKbUseCases(sqlDB)
	repo := persistence.NewKnowledgeBaseRepository(sqlDB)
	ctx := context.Background()

	setup := func(t *testing.T) (ws, space string) {
		t.Helper()
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws = createWorkspace(t, sqlDB, "ws-search-write")
		space = createSpace(t, sqlDB, ws, "eng")
		return ws, space
	}

	t.Run("本文を書き換えて2回保存すると古い内容ではなく新しい内容が反映される", func(t *testing.T) {
		ws, space := setup(t)
		target := mustCreatePage(ctx, t, uc, ws, space, nil, "参照先")
		page := mustCreatePage(ctx, t, uc, ws, space, nil, "対象ページ")

		_, err := uc.replace.Execute(ctx, kb.ReplacePageBlocksInput{
			WorkspaceID: ws, PageID: page.ID, Doc: pageRefDoc("最初の内容", target.ID), EditorUserID: 1,
		})
		require.NoError(t, err)

		title, body, found := queryPageSearchRow(t, sqlDB, page.ID)
		require.True(t, found, "1回目の保存で page_search 行ができる")
		assert.Equal(t, "対象ページ", title, "titleはpages.titleの写し")
		assert.Contains(t, body, "最初の内容")
		assert.Equal(t, 1, countPageLinksForPage(t, sqlDB, ws, page.ID), "1回目の保存でpage_linksが1行できる")

		_, err = uc.replace.Execute(ctx, kb.ReplacePageBlocksInput{
			WorkspaceID: ws, PageID: page.ID,
			Doc:          `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"書き換え後の内容"}]}]}`,
			EditorUserID: 1,
		})
		require.NoError(t, err)

		_, body2, found2 := queryPageSearchRow(t, sqlDB, page.ID)
		require.True(t, found2)
		assert.NotContains(t, body2, "最初の内容", "古い内容は残らない（UPSERTで完全に置き換わる）")
		assert.Contains(t, body2, "書き換え後の内容", "新しい内容が反映される")
		assert.Equal(t, 0, countPageLinksForPage(t, sqlDB, ws, page.ID),
			"2回目の保存はpageRefを含まないのでpage_linksは張り替わって0行になる")
	})

	t.Run("ブロック削除でそのブロックのpage_linksが消える（CASCADE経由）", func(t *testing.T) {
		ws, space := setup(t)
		target := mustCreatePage(ctx, t, uc, ws, space, nil, "参照先2")
		page := mustCreatePage(ctx, t, uc, ws, space, nil, "対象ページ2")
		_, err := uc.replace.Execute(ctx, kb.ReplacePageBlocksInput{
			WorkspaceID: ws, PageID: page.ID, Doc: pageRefDoc("本文", target.ID), EditorUserID: 1,
		})
		require.NoError(t, err)

		var blockID string
		require.NoError(t, sqlDB.QueryRow(`SELECT id::text FROM blocks WHERE page_id = $1`, page.ID).Scan(&blockID))
		var linkCountBefore int
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM page_links WHERE source_block_id = $1`, blockID).Scan(&linkCountBefore))
		require.Equal(t, 1, linkCountBefore, "前提: ブロックがpage_linksを1行持つ")

		// 保存経路（ReplacePageBlocks 自身の DELETE+INSERT）を経由せず、ブロックそのものを
		// 直接消す。CASCADE（fk_page_links_source_block）そのものを確かめるため。
		_, err = sqlDB.Exec(`DELETE FROM blocks WHERE id = $1`, blockID)
		require.NoError(t, err)

		var linkCountAfter int
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM page_links WHERE source_block_id = $1`, blockID).Scan(&linkCountAfter))
		assert.Equal(t, 0, linkCountAfter, "ブロックが消えたらそのブロックのpage_linksもCASCADEで消える")
	})

	t.Run("参照先ページ削除でpage_linksが消える（CASCADE経由）", func(t *testing.T) {
		ws, space := setup(t)
		target := mustCreatePage(ctx, t, uc, ws, space, nil, "消される参照先")
		page := mustCreatePage(ctx, t, uc, ws, space, nil, "対象ページ3")
		_, err := uc.replace.Execute(ctx, kb.ReplacePageBlocksInput{
			WorkspaceID: ws, PageID: page.ID, Doc: pageRefDoc("本文", target.ID), EditorUserID: 1,
		})
		require.NoError(t, err)
		require.Equal(t, 1, countPageLinksForPage(t, sqlDB, ws, page.ID))

		require.NoError(t, repo.DeletePageSubtree(ctx, ws, target.ID))

		assert.Equal(t, 0, countPageLinksForPage(t, sqlDB, ws, page.ID), "参照先ページが消えたらリンクもCASCADEで消える")
	})

	t.Run("再構築（RebuildPageSearchAndLinks）を2回流しても同じ結果になる（冪等性）", func(t *testing.T) {
		ws, space := setup(t)
		target := mustCreatePage(ctx, t, uc, ws, space, nil, "参照先4")
		page := mustCreatePage(ctx, t, uc, ws, space, nil, "対象ページ4")
		_, err := uc.replace.Execute(ctx, kb.ReplacePageBlocksInput{
			WorkspaceID: ws, PageID: page.ID, Doc: pageRefDoc("再構築の確認", target.ID), EditorUserID: 1,
		})
		require.NoError(t, err)

		title0, body0, found0 := queryPageSearchRow(t, sqlDB, page.ID)
		require.True(t, found0)
		require.Equal(t, 1, countPageLinksForPage(t, sqlDB, ws, page.ID))

		require.NoError(t, repo.RebuildPageSearchAndLinks(ctx, ws, page.ID))
		title1, body1, found1 := queryPageSearchRow(t, sqlDB, page.ID)
		require.True(t, found1)
		assert.Equal(t, title0, title1)
		assert.Equal(t, body0, body1)
		assert.Equal(t, 1, countPageLinksForPage(t, sqlDB, ws, page.ID), "1回目の再構築後も1行のまま")

		require.NoError(t, repo.RebuildPageSearchAndLinks(ctx, ws, page.ID))
		title2, body2, found2 := queryPageSearchRow(t, sqlDB, page.ID)
		require.True(t, found2)
		assert.Equal(t, title0, title2)
		assert.Equal(t, body0, body2)
		assert.Equal(t, 1, countPageLinksForPage(t, sqlDB, ws, page.ID), "2回目の再構築後も重複せず1行のまま（冪等）")
	})
}

// TestKnowledgeBaseSearchBodyMatch_Integration は本文検索（FRESTYLE-434 段 4）の
// 日本語の部分一致・matchField/excerpt の判定・可視性のふるいを実 PostgreSQL で固定する。
func TestKnowledgeBaseSearchBodyMatch_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "設計メモ")
	_, err := f.pageUC.replace.Execute(ctx, kb.ReplacePageBlocksInput{
		WorkspaceID: f.ws, PageID: page.ID,
		Doc:          `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"この段落には Docker の使い方が書いてある"}]}]}`,
		EditorUserID: 1,
	})
	require.NoError(t, err)

	alice := f.principalFor(ctx, t, f.alice)
	f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)

	t.Run("日本語の本文一致を拾い、matchField=bodyと抜粋を返す", func(t *testing.T) {
		results, err := kb.NewSearchViewablePagesUseCase(f.perm).Execute(ctx,
			kb.SearchViewablePagesInput{WorkspaceID: f.ws, UserID: f.alice, Query: "使い方"})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, page.ID, results[0].Page.ID)
		assert.Equal(t, kb.SearchMatchFieldBody, results[0].MatchField)
		assert.Contains(t, results[0].Excerpt, "使い方")
	})

	t.Run("titleが一致するときはmatchField=titleでexcerptを出さない", func(t *testing.T) {
		results, err := kb.NewSearchViewablePagesUseCase(f.perm).Execute(ctx,
			kb.SearchViewablePagesInput{WorkspaceID: f.ws, UserID: f.alice, Query: "設計"})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, kb.SearchMatchFieldTitle, results[0].MatchField)
		assert.Empty(t, results[0].Excerpt)
	})

	t.Run("題名にも本文にも一致しなければ出ない", func(t *testing.T) {
		results, err := kb.NewSearchViewablePagesUseCase(f.perm).Execute(ctx,
			kb.SearchViewablePagesInput{WorkspaceID: f.ws, UserID: f.alice, Query: "無関係な語"})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("見えないスペースの本文一致は出ない", func(t *testing.T) {
		secretSpace := createSpace(t, sqlDB, f.ws, "search-secret")
		f.makePrivate(t, secretSpace)
		secret := mustCreatePage(ctx, t, f.pageUC, f.ws, secretSpace, nil, "非公開ページ")
		_, err := f.pageUC.replace.Execute(ctx, kb.ReplacePageBlocksInput{
			WorkspaceID: f.ws, PageID: secret.ID,
			Doc:          `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Docker の秘密の手順"}]}]}`,
			EditorUserID: 1,
		})
		require.NoError(t, err)

		// alice はこの private スペースへの付与を持たない。
		results, err := kb.NewSearchViewablePagesUseCase(f.perm).Execute(ctx,
			kb.SearchViewablePagesInput{WorkspaceID: f.ws, UserID: f.alice, Query: "秘密の手順"})
		require.NoError(t, err)
		assert.Empty(t, results, "見えないスペースの本文一致は検索に出ない")
	})
}

// TestKnowledgeBaseBacklinks_Integration は逆リンク（FRESTYLE-434 段 4）の可視判定を
// 実 PostgreSQL で固定する。見えない参照元ページ（権限の無いスペース）からのリンクは
// 一覧に出ないこと。
func TestKnowledgeBaseBacklinks_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	target := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "参照される側")

	visibleSource := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "見える参照元")
	_, err := f.pageUC.replace.Execute(ctx, kb.ReplacePageBlocksInput{
		WorkspaceID: f.ws, PageID: visibleSource.ID, Doc: pageRefDoc("参照", target.ID), EditorUserID: 1,
	})
	require.NoError(t, err)

	// 見えない参照元: private スペース（alice には付与しない）。
	secretSpace := createSpace(t, sqlDB, f.ws, "backlink-secret")
	f.makePrivate(t, secretSpace)
	hiddenSource := mustCreatePage(ctx, t, f.pageUC, f.ws, secretSpace, nil, "見えない参照元")
	_, err = f.pageUC.replace.Execute(ctx, kb.ReplacePageBlocksInput{
		WorkspaceID: f.ws, PageID: hiddenSource.ID, Doc: pageRefDoc("参照", target.ID), EditorUserID: 1,
	})
	require.NoError(t, err)

	alice := f.principalFor(ctx, t, f.alice)
	f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)

	backlinks, err := kb.NewListPageBacklinksUseCase(f.perm).Execute(ctx, kb.ListPageBacklinksInput{
		WorkspaceID: f.ws, UserID: f.alice, PageID: target.ID,
	})
	require.NoError(t, err)
	ids := make([]string, 0, len(backlinks))
	for _, p := range backlinks {
		ids = append(ids, p.ID)
	}
	assert.ElementsMatch(t, []string{visibleSource.ID}, ids,
		"見える参照元だけが出て、権限の無いスペースの参照元は出ない")

	// bob は secretSpace にも付与を持つので、両方見える。
	bob := f.principalFor(ctx, t, f.bob)
	f.grantSpace(ctx, t, f.spaceA, bob.ID, domain.GrantRoleViewer)
	f.grantSpace(ctx, t, secretSpace, bob.ID, domain.GrantRoleViewer)
	backlinksForBob, err := kb.NewListPageBacklinksUseCase(f.perm).Execute(ctx, kb.ListPageBacklinksInput{
		WorkspaceID: f.ws, UserID: f.bob, PageID: target.ID,
	})
	require.NoError(t, err)
	idsForBob := make([]string, 0, len(backlinksForBob))
	for _, p := range backlinksForBob {
		idsForBob = append(idsForBob, p.ID)
	}
	assert.ElementsMatch(t, []string{visibleSource.ID, hiddenSource.ID}, idsForBob,
		"両方のスペースが見える相手には両方の参照元が出る")
}
