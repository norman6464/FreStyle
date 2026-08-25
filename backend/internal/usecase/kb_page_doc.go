package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/pkg/fracindex"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrPageDocInvalid は ProseMirror ドキュメントとして解釈できない入力に返す（400 相当）。
var ErrPageDocInvalid = errors.New("invalid prosemirror doc")

// ErrPageDocUnknownNodeType は domain.BlockType に無いブロックノードを含む入力に返す（400 相当）。
// スキーマに無いノード名を保存すると読み出したドキュメントがエディタで開けなくなるため、入口で弾く。
var ErrPageDocUnknownNodeType = errors.New("unknown block node type")

// kbContainerBlockTypes は子がブロック行になる「容器ノード」。それ以外の既知ノードは
// 「葉ノード」で、content（text ノードとマークの配列）を行にせず inline に丸ごと持つ。
// 粒度の境界はスキーマ設計（blocks.inline のコメント）で決めたもの: 文字単位で行を作ると
// 1 段落の編集が大量の行更新になるため、行はブロックで止める。
var kbContainerBlockTypes = map[domain.BlockType]bool{
	domain.BlockTypeBlockquote:  true,
	domain.BlockTypeBulletList:  true,
	domain.BlockTypeOrderedList: true,
	domain.BlockTypeListItem:    true,
	domain.BlockTypeTaskList:    true,
	domain.BlockTypeTaskItem:    true,
	domain.BlockTypeTable:       true,
	domain.BlockTypeTableRow:    true,
	domain.BlockTypeTableHeader: true,
	domain.BlockTypeTableCell:   true,
}

// kbDocNode はブロック行 1 つに対応する中間表現。分解（doc → 行）と組み立て（行 → doc）が
// この木を共有することで、保存する snapshot が必ず「行から再生成できる形」になる。
type kbDocNode struct {
	Type     domain.BlockType
	Attrs    string  // JSON object。属性が無ければ "{}"（NULL と {} の二通りを作らない）
	Inline   *string // JSON array。葉ノードの content。容器ノード・content 無しは nil
	Children []*kbDocNode
}

// kbRawNode は ProseMirror ノードの JSON を最小限に読むための型。
// text / marks 等ここに無いフィールドは、葉ノードでは content の中に丸ごと残り、
// ブロックノード自身に付いていた場合は正規化で落ちる（行スキーマに置き場が無いため）。
type kbRawNode struct {
	Type    string          `json:"type"`
	Attrs   json.RawMessage `json:"attrs"`
	Content json.RawMessage `json:"content"`
}

// parsePageDoc は ProseMirror ドキュメント（type='doc'）をブロック木に分解する。
// ルート doc ノードは行にせず、doc.content の各ノードがトップレベルブロックになる。
func parsePageDoc(doc string) ([]*kbDocNode, error) {
	var root kbRawNode
	if err := json.Unmarshal([]byte(doc), &root); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPageDocInvalid, err)
	}
	if root.Type != "doc" {
		return nil, fmt.Errorf("%w: ルートは type='doc' が必要（got %q）", ErrPageDocInvalid, root.Type)
	}
	return parseBlockNodes(root.Content)
}

// parseBlockNodes は content 配列（JSON）をブロックノード列として解釈する。空・省略は 0 件。
func parseBlockNodes(content json.RawMessage) ([]*kbDocNode, error) {
	if len(content) == 0 || string(content) == "null" {
		return []*kbDocNode{}, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, fmt.Errorf("%w: content が配列ではありません: %v", ErrPageDocInvalid, err)
	}
	nodes := make([]*kbDocNode, 0, len(items))
	for _, item := range items {
		n, err := parseBlockNode(item)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func parseBlockNode(raw json.RawMessage) (*kbDocNode, error) {
	var rn kbRawNode
	if err := json.Unmarshal(raw, &rn); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPageDocInvalid, err)
	}
	t := domain.BlockType(rn.Type)
	if !t.Valid() {
		return nil, fmt.Errorf("%w: %q", ErrPageDocUnknownNodeType, rn.Type)
	}

	node := &kbDocNode{Type: t, Attrs: "{}"}
	if len(rn.Attrs) > 0 && string(rn.Attrs) != "null" {
		// attrs は object であること（DDL の CHECK と同じ壁を入口にも置く）。
		var m map[string]json.RawMessage
		if err := json.Unmarshal(rn.Attrs, &m); err != nil {
			return nil, fmt.Errorf("%w: attrs が object ではありません: %v", ErrPageDocInvalid, err)
		}
		if len(m) > 0 {
			node.Attrs = string(rn.Attrs)
		}
	}

	if kbContainerBlockTypes[t] {
		children, err := parseBlockNodes(rn.Content)
		if err != nil {
			return nil, err
		}
		node.Children = children
		return node, nil
	}
	// 葉ノード: content はインライン内容として丸ごと inline に持つ。
	if len(rn.Content) > 0 && string(rn.Content) != "null" {
		// 配列であることの検証を兼ねて要素単位で読み直し、空白差を吸収した形で持ち直す
		// （jsonb は保存時に正規化されるため、入力の空白を残しても意味が無い）。
		var items []json.RawMessage
		if err := json.Unmarshal(rn.Content, &items); err != nil {
			return nil, fmt.Errorf("%w: content が配列ではありません: %v", ErrPageDocInvalid, err)
		}
		if len(items) > 0 {
			compact, err := json.Marshal(items)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrPageDocInvalid, err)
			}
			s := string(compact)
			node.Inline = &s
		}
	}
	return node, nil
}

// flattenPageDoc はブロック木を保存用の行（文書順・親が先）へ平坦化する。
// 兄弟の position は fracindex の末尾追加で採番する（i 件目 = Between(直前, "")）。
func flattenPageDoc(nodes []*kbDocNode) ([]repository.BlockWrite, error) {
	out := make([]repository.BlockWrite, 0)
	var walk func(nodes []*kbDocNode, parentIndex int) error
	walk = func(nodes []*kbDocNode, parentIndex int) error {
		prev := ""
		for _, n := range nodes {
			pos, err := fracindex.Between(prev, "")
			if err != nil {
				return err
			}
			prev = pos
			idx := len(out)
			out = append(out, repository.BlockWrite{
				ParentIndex: parentIndex,
				Position:    pos,
				Type:        n.Type,
				Attrs:       n.Attrs,
				Inline:      n.Inline,
			})
			if err := walk(n.Children, idx); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(nodes, -1); err != nil {
		return nil, err
	}
	return out, nil
}

// treeFromBlocks は DB のブロック行をブロック木へ組み直す。兄弟は position のバイト順
// （= COLLATE "C" の ORDER BY と同じ）に並べる。親が見つからない行は closure や FK が
// 壊れているサインなのでエラーにする（黙って本文を欠落させない）。
func treeFromBlocks(blocks []domain.Block) ([]*kbDocNode, error) {
	nodes := make(map[string]*kbDocNode, len(blocks))
	order := make(map[string]string, len(blocks)) // id → position（兄弟ソート用）
	for _, b := range blocks {
		if !b.Type.Valid() {
			return nil, fmt.Errorf("%w: %q", ErrPageDocUnknownNodeType, b.Type)
		}
		n := &kbDocNode{Type: b.Type, Attrs: b.Attrs}
		if b.Attrs == "" {
			n.Attrs = "{}"
		}
		if b.Inline != nil {
			s := *b.Inline
			n.Inline = &s
		}
		nodes[b.ID] = n
		order[b.ID] = b.Position
	}
	roots := make([]*kbDocNode, 0)
	rootIDs := make([]string, 0)
	childIDs := make(map[string][]string, len(blocks))
	for _, b := range blocks {
		if b.ParentID == nil {
			rootIDs = append(rootIDs, b.ID)
			continue
		}
		if _, ok := nodes[*b.ParentID]; !ok {
			return nil, fmt.Errorf("ブロック %s の親 %s がページ内にありません", b.ID, *b.ParentID)
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
			nodes[pid].Children = append(nodes[pid].Children, nodes[id])
		}
	}
	for _, id := range rootIDs {
		roots = append(roots, nodes[id])
	}
	return roots, nil
}

// renderPageDoc はブロック木から ProseMirror ドキュメント（正規形）を組み立てる。
// 正規形: doc の content は空でも必ず配列で出す / 各ノードの attrs は空 object なら出さない /
// content は無ければ出さない。parsePageDoc → renderPageDoc の往復は正規形の入力に対して同値になる。
func renderPageDoc(nodes []*kbDocNode) (string, error) {
	content, err := renderBlockNodes(nodes)
	if err != nil {
		return "", err
	}
	doc, err := json.Marshal(struct {
		Type    string            `json:"type"`
		Content []json.RawMessage `json:"content"`
	}{Type: "doc", Content: content})
	if err != nil {
		return "", err
	}
	return string(doc), nil
}

func renderBlockNodes(nodes []*kbDocNode) ([]json.RawMessage, error) {
	out := make([]json.RawMessage, 0, len(nodes))
	for _, n := range nodes {
		raw, err := renderBlockNode(n)
		if err != nil {
			return nil, err
		}
		out = append(out, raw)
	}
	return out, nil
}

func renderBlockNode(n *kbDocNode) (json.RawMessage, error) {
	node := struct {
		Type    string          `json:"type"`
		Attrs   json.RawMessage `json:"attrs,omitempty"`
		Content json.RawMessage `json:"content,omitempty"`
	}{Type: string(n.Type)}

	if n.Attrs != "" && n.Attrs != "{}" {
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(n.Attrs), &m); err != nil {
			return nil, fmt.Errorf("ブロックの attrs が壊れています: %w", err)
		}
		if len(m) > 0 {
			node.Attrs = json.RawMessage(n.Attrs)
		}
	}

	switch {
	case len(n.Children) > 0:
		children, err := renderBlockNodes(n.Children)
		if err != nil {
			return nil, err
		}
		arr, err := json.Marshal(children)
		if err != nil {
			return nil, err
		}
		node.Content = arr
	case n.Inline != nil:
		node.Content = json.RawMessage(*n.Inline)
	}
	return json.Marshal(node)
}
