package persistence

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// pageSearchTextNodeType / pageSearchPageRefNodeType は inline JSON の中で拾うノードの
// type 名。usecase/kb の kbInlineTextNodeType / kbPageRefNodeType と同じ値だが、
// このパッケージからは import できない定数として独立して持つ
// （extractPageSearchFromBlocks の doc 参照 — 依存の向きの理由）。
const (
	pageSearchTextNodeType    = "text"
	pageSearchPageRefNodeType = "pageRef"
)

// pageSearchInlineNode は inline 配列の 1 要素を最小限に読むための型。
// usecase/kb.kbInlineTextNode と同じ形。
type pageSearchInlineNode struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Attrs struct {
		PageID string `json:"pageId"`
	} `json:"attrs"`
}

// writePageSearchAndLinks は page_search の UPSERT と page_links の張り替えを行う。
// ReplacePageBlocks（本文保存の最終ステップ）と RebuildPageSearchAndLinks（一回限りの
// 再構築）の両方から呼ぶ、書き込みの中核ロジック（コードの重複を避けるための共有関数。
// FRESTYLE-434 段 4）。呼び出し元は同じトランザクションの qtx を渡すこと。
//
// 抽出（doc / blocks から body・pageLinks を作る部分）はここでは行わない。呼び出し元が
// 用意した値をそのまま書き込むだけ — 抽出ロジックの置き場所が呼び出し元によって違うため:
//   - ReplacePageBlocks は usecase/kb.ReplacePageBlocksUseCase.Execute が ProseMirror の
//     doc（保存直前の正規化済みの木）から抽出したものを渡す。
//   - RebuildPageSearchAndLinks はこのファイル内の extractPageSearchFromBlocks が、
//     既に保存済みの blocks 行から抽出したものを渡す。
func writePageSearchAndLinks(
	ctx context.Context, qtx *sqlcgen.Queries, wsID, pgID uuid.UUID, title, body string, pageLinks []repository.PageLinkWrite,
) error {
	// 1. page_search を焼き直す。
	if err := qtx.UpsertPageSearch(ctx, sqlcgen.UpsertPageSearchParams{
		PageID:      pgID,
		WorkspaceID: wsID,
		Title:       title,
		Body:        body,
	}); err != nil {
		return err
	}

	// 2. page_links を張り替える（前半: このページのブロックが持っていたリンクを全消し）。
	if err := qtx.DeletePageLinksBySourceBlockIDsInPage(ctx, sqlcgen.DeletePageLinksBySourceBlockIDsInPageParams{
		WorkspaceID: wsID,
		PageID:      pgID,
	}); err != nil {
		return err
	}
	if len(pageLinks) == 0 {
		return nil
	}

	// 3. 参照先が実在するものだけに絞る（リンク切れは黙って除外する — PageLinkWrite の
	// doc 参照。target_page_id は pages への FK なので、存在しない ID のまま INSERT すると
	// 外部キー違反で保存全体が落ちてしまう）。
	targetSet := make(map[uuid.UUID]struct{}, len(pageLinks))
	targetIDs := make([]uuid.UUID, 0, len(pageLinks))
	for _, l := range pageLinks {
		id, err := uuid.Parse(l.TargetPageID)
		if err != nil {
			// usecase 側（extractPageLinks）は canonicalPageRefID で正規化済みの値しか
			// 積まないが、念のため壊れた値は保存全体を落とさず読み飛ばす。
			continue
		}
		if _, dup := targetSet[id]; dup {
			continue
		}
		targetSet[id] = struct{}{}
		targetIDs = append(targetIDs, id)
	}
	if len(targetIDs) == 0 {
		return nil
	}
	idsJSON, err := json.Marshal(targetIDs)
	if err != nil {
		return err
	}
	existingRows, err := qtx.ListExistingPageIDsAmong(ctx, idsJSON)
	if err != nil {
		return err
	}
	existing := make(map[uuid.UUID]struct{}, len(existingRows))
	for _, id := range existingRows {
		existing[id] = struct{}{}
	}

	// 4. (後半) 実在確認済みの参照先ごとに INSERT する。1 つのブロックが同じページを
	// 複数回参照する場合は InsertPageLink の ON CONFLICT DO NOTHING で 1 行に畳まれる。
	for _, l := range pageLinks {
		tgtID, err := uuid.Parse(l.TargetPageID)
		if err != nil {
			continue
		}
		if _, ok := existing[tgtID]; !ok {
			continue
		}
		srcID, err := uuid.Parse(l.SourceBlockID)
		if err != nil {
			continue
		}
		if err := qtx.InsertPageLink(ctx, sqlcgen.InsertPageLinkParams{
			SourceBlockID: srcID,
			TargetPageID:  tgtID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// RebuildPageSearchAndLinks は既存ページ 1 件の page_search / page_links を、現在の
// blocks から作り直す（repository.KnowledgeBaseRepository の doc 参照）。
// DELETE + UPSERT で書き直すため、同じページに何度呼んでも結果は同じ（冪等）。
func (r *knowledgeBaseRepository) RebuildPageSearchAndLinks(ctx context.Context, workspaceID, pageID string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}
	return r.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
		page, err := findPageWith(ctx, qtx, workspaceID, pageID)
		if err != nil {
			return err
		}
		rows, err := qtx.ListBlocksByPage(ctx, sqlcgen.ListBlocksByPageParams{WorkspaceID: wsID, PageID: pgID})
		if err != nil {
			return err
		}
		blocks := make([]domain.Block, 0, len(rows))
		for _, row := range rows {
			blocks = append(blocks, toDomainBlock(row))
		}
		body, links := extractPageSearchFromBlocks(blocks)
		return writePageSearchAndLinks(ctx, qtx, wsID, pgID, page.Title, body, links)
	})
}

// orderedBlockNode は domain.Block 行を「文書順に近い」木として歩くための最小限の
// 中間表現。usecase/kb.kbDocNode と同じ役割だが、extractPageSearchFromBlocks が必要と
// するフィールド（id / inline / children）だけに絞ってある。
type orderedBlockNode struct {
	id       string
	inline   *string
	children []*orderedBlockNode
}

// buildOrderedBlockForest は blocks 行を parent_id / position から木へ組み直す
// （usecase/kb.treeFromBlocks と同じ考え方の独立した再実装）。
//
// なぜ treeFromBlocks を呼ばずに書き直すのか: このリポジトリのクリーンアーキテクチャは
// handler → usecase → repository/infra → domain の一方通行で、repository の実装である
// このパッケージ（persistence）は usecase/kb を import できない。呼べてしまうと
// 依存が逆流する。抽出ロジックが 2 箇所に分かれるのは望ましくないが、層の境界を守る方を
// 優先する（本チケットの指示 — 抽出は usecase/kb に置く — とも一致する）。
//
// 壊れた親参照（存在しない parent_id）は無視する。treeFromBlocks（表示経路）は同じ状況を
// エラーにするが、こちらは検索キャッシュの再構築という補助的な経路なので、壊れた行が
// あってもそこだけ本文から漏れるだけに留め、再構築全体を失敗させない。
func buildOrderedBlockForest(blocks []domain.Block) []*orderedBlockNode {
	nodes := make(map[string]*orderedBlockNode, len(blocks))
	order := make(map[string]string, len(blocks))
	for _, b := range blocks {
		nodes[b.ID] = &orderedBlockNode{id: b.ID, inline: b.Inline}
		order[b.ID] = b.Position
	}
	rootIDs := make([]string, 0)
	childIDs := make(map[string][]string, len(blocks))
	for _, b := range blocks {
		if b.ParentID == nil {
			rootIDs = append(rootIDs, b.ID)
			continue
		}
		if _, ok := nodes[*b.ParentID]; !ok {
			continue
		}
		childIDs[*b.ParentID] = append(childIDs[*b.ParentID], b.ID)
	}
	sortByPosition := func(ids []string) {
		sort.SliceStable(ids, func(i, j int) bool { return order[ids[i]] < order[ids[j]] })
	}
	sortByPosition(rootIDs)
	for _, ids := range childIDs {
		sortByPosition(ids)
	}
	for pid, ids := range childIDs {
		for _, id := range ids {
			nodes[pid].children = append(nodes[pid].children, nodes[id])
		}
	}
	roots := make([]*orderedBlockNode, 0, len(rootIDs))
	for _, id := range rootIDs {
		roots = append(roots, nodes[id])
	}
	return roots
}

// extractPageSearchFromBlocks は保存済みの blocks 行から body（本文プレーンテキスト）と
// pageLinks（page_links の材料）を組み立てる。RebuildPageSearchAndLinks が使う。
//
// usecase/kb.extractPageBodyText / extractPageLinks と同じ考え方（"text" 型インライン
// ノードの .text を連結する・pageRef ノードの attrs.pageId を集める）を、
// buildOrderedBlockForest の doc に書いた理由でこのパッケージに閉じて独立に実装している。
//
// usecase/kb 側にある「参照先ページの種類数を kbPageRefMaxResolve=100 で打ち切る」天井は
// ここでは掛けない。こちらは通常の保存経路（1 リクエストごとに必ず通る）ではなく、
// 一回限りの再構築（cmd/rebuildsearchindex）専用の補助的な経路なので、同じコスト上限を
// 課さなくても実害が小さい。
func extractPageSearchFromBlocks(blocks []domain.Block) (body string, pageLinks []repository.PageLinkWrite) {
	roots := buildOrderedBlockForest(blocks)
	var textBuf strings.Builder
	var walk func(nodes []*orderedBlockNode)
	walk = func(nodes []*orderedBlockNode) {
		for _, n := range nodes {
			if len(n.children) > 0 {
				walk(n.children)
				continue
			}
			if n.inline == nil {
				continue
			}
			var items []pageSearchInlineNode
			if err := json.Unmarshal([]byte(*n.inline), &items); err != nil {
				continue
			}
			var blockText strings.Builder
			for _, it := range items {
				switch it.Type {
				case pageSearchTextNodeType:
					blockText.WriteString(it.Text)
				case pageSearchPageRefNodeType:
					if id, err := uuid.Parse(it.Attrs.PageID); err == nil {
						pageLinks = append(pageLinks, repository.PageLinkWrite{
							SourceBlockID: n.id,
							TargetPageID:  id.String(),
						})
					}
				}
			}
			if blockText.Len() > 0 {
				if textBuf.Len() > 0 {
					textBuf.WriteByte('\n')
				}
				textBuf.WriteString(blockText.String())
			}
		}
	}
	walk(roots)
	return textBuf.String(), pageLinks
}
