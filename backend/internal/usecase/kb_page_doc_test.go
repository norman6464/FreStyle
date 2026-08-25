package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// writesToDomainBlocks は repository 実装の採番（ParentIndex → ID 解決）を模して、
// 保存用の行を DB から読み出した形（domain.Block）へ変換する。往復テスト用。
func writesToDomainBlocks(t *testing.T, rows []repository.BlockWrite) []domain.Block {
	t.Helper()
	blocks := make([]domain.Block, 0, len(rows))
	for i, r := range rows {
		b := domain.Block{
			ID:       fmt.Sprintf("block-%04d", i),
			PageID:   "page-1",
			Position: r.Position,
			Type:     r.Type,
			Attrs:    r.Attrs,
		}
		if r.ParentIndex >= 0 {
			require.Less(t, r.ParentIndex, i, "親は文書順で自分より前にあること")
			pid := fmt.Sprintf("block-%04d", r.ParentIndex)
			b.ParentID = &pid
		}
		if r.Inline != nil {
			s := *r.Inline
			b.Inline = &s
		}
		blocks = append(blocks, b)
	}
	return blocks
}

// requireJSONEq は 2 つの JSON 文字列を意味的に比較する（キー順・空白の差を無視）。
func requireJSONEq(t *testing.T, want, got string) {
	t.Helper()
	var w, g any
	require.NoError(t, json.Unmarshal([]byte(want), &w))
	require.NoError(t, json.Unmarshal([]byte(got), &g))
	require.Equal(t, w, g)
}

// Test_doc往復_分解して組み立てると同値 は decompose → assemble の往復同値を固定する。
// 入力は正規形（attrs が空の object を持たない・doc の content は配列）の ProseMirror doc。
func Test_doc往復_分解して組み立てると同値(t *testing.T) {
	cases := []struct {
		name string
		doc  string
	}{
		{
			name: "空のdoc",
			doc:  `{"type":"doc","content":[]}`,
		},
		{
			name: "段落だけ",
			doc: `{"type":"doc","content":[
				{"type":"paragraph","content":[{"type":"text","text":"こんにちは"}]},
				{"type":"paragraph"},
				{"type":"paragraph","content":[{"type":"text","text":"二段落目"},{"type":"hardBreak"},{"type":"text","text":"改行後"}]}
			]}`,
		},
		{
			name: "見出しとマーク",
			doc: `{"type":"doc","content":[
				{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"見出し"}]},
				{"type":"paragraph","content":[
					{"type":"text","marks":[{"type":"bold"},{"type":"italic"}],"text":"強調"},
					{"type":"text","marks":[{"type":"link","attrs":{"href":"https://example.com","target":"_blank"}}],"text":"リンク"}
				]}
			]}`,
		},
		{
			name: "ネストしたリスト",
			doc: `{"type":"doc","content":[
				{"type":"bulletList","content":[
					{"type":"listItem","content":[
						{"type":"paragraph","content":[{"type":"text","text":"親項目"}]},
						{"type":"orderedList","attrs":{"start":3},"content":[
							{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"子項目1"}]}]},
							{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"子項目2"}]}]}
						]}
					]},
					{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"兄弟項目"}]}]}
				]}
			]}`,
		},
		{
			name: "表",
			doc: `{"type":"doc","content":[
				{"type":"table","content":[
					{"type":"tableRow","content":[
						{"type":"tableHeader","attrs":{"colspan":1,"rowspan":1},"content":[{"type":"paragraph","content":[{"type":"text","text":"列A"}]}]},
						{"type":"tableHeader","attrs":{"colspan":1,"rowspan":1},"content":[{"type":"paragraph","content":[{"type":"text","text":"列B"}]}]}
					]},
					{"type":"tableRow","content":[
						{"type":"tableCell","attrs":{"colspan":1,"rowspan":1},"content":[{"type":"paragraph","content":[{"type":"text","text":"a1"}]}]},
						{"type":"tableCell","attrs":{"colspan":2,"rowspan":1},"content":[{"type":"paragraph"}]}
					]}
				]}
			]}`,
		},
		{
			name: "タスクリスト",
			doc: `{"type":"doc","content":[
				{"type":"taskList","content":[
					{"type":"taskItem","attrs":{"checked":true},"content":[{"type":"paragraph","content":[{"type":"text","text":"済み"}]}]},
					{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"未着手"}]}]}
				]}
			]}`,
		},
		{
			name: "コードブロック",
			doc: `{"type":"doc","content":[
				{"type":"codeBlock","attrs":{"language":"go"},"content":[{"type":"text","text":"package main\nfunc main() {}"}]}
			]}`,
		},
		{
			name: "画像と区切り線と引用",
			doc: `{"type":"doc","content":[
				{"type":"image","attrs":{"src":"https://example.com/a.png","alt":"代替","title":null}},
				{"type":"horizontalRule"},
				{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"引用文"}]}]}
			]}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parsePageDoc(tc.doc)
			require.NoError(t, err)
			rows, err := flattenPageDoc(tree)
			require.NoError(t, err)

			// 行（DB の形）を経由して組み立て直しても同値であること。
			rebuilt, err := treeFromBlocks(writesToDomainBlocks(t, rows))
			require.NoError(t, err)
			got, err := renderPageDoc(rebuilt)
			require.NoError(t, err)
			requireJSONEq(t, tc.doc, got)

			// 行を経由しない直接の組み立てでも同値であること。
			direct, err := renderPageDoc(tree)
			require.NoError(t, err)
			requireJSONEq(t, tc.doc, direct)
		})
	}
}

func Test_doc分解_不正な入力を弾く(t *testing.T) {
	cases := []struct {
		name    string
		doc     string
		wantErr error
	}{
		{"JSONでない", `not-json`, ErrPageDocInvalid},
		{"ルートがdocでない", `{"type":"paragraph"}`, ErrPageDocInvalid},
		{"contentが配列でない", `{"type":"doc","content":{"type":"paragraph"}}`, ErrPageDocInvalid},
		{"未知のブロックノード", `{"type":"doc","content":[{"type":"iframe"}]}`, ErrPageDocUnknownNodeType},
		{"容器の中の未知ノード", `{"type":"doc","content":[{"type":"bulletList","content":[{"type":"video"}]}]}`, ErrPageDocUnknownNodeType},
		{"インラインノードがトップレベルに来る", `{"type":"doc","content":[{"type":"text","text":"裸のテキスト"}]}`, ErrPageDocUnknownNodeType},
		{"attrsがobjectでない", `{"type":"doc","content":[{"type":"paragraph","attrs":[1,2]}]}`, ErrPageDocInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsePageDoc(tc.doc)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func Test_doc分解_行の形が正しい(t *testing.T) {
	doc := `{"type":"doc","content":[
		{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"h"}]},
		{"type":"bulletList","content":[
			{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"item"}]}]}
		]}
	]}`
	tree, err := parsePageDoc(doc)
	require.NoError(t, err)
	rows, err := flattenPageDoc(tree)
	require.NoError(t, err)
	require.Len(t, rows, 4) // heading / bulletList / listItem / paragraph

	// トップレベル: heading と bulletList（文書順・親なし）。
	require.Equal(t, domain.BlockTypeHeading, rows[0].Type)
	require.Equal(t, -1, rows[0].ParentIndex)
	require.JSONEq(t, `{"level":1}`, rows[0].Attrs)
	require.NotNil(t, rows[0].Inline, "葉ノードは content を inline に持つ")
	require.Equal(t, domain.BlockTypeBulletList, rows[1].Type)
	require.Equal(t, -1, rows[1].ParentIndex)
	require.Nil(t, rows[1].Inline, "容器ノードの inline は NULL")
	require.Equal(t, "{}", rows[1].Attrs, "属性なしは空 object（NULL と {} の二通りを作らない）")

	// 入れ子: listItem の親は bulletList、paragraph の親は listItem。
	require.Equal(t, domain.BlockTypeListItem, rows[2].Type)
	require.Equal(t, 1, rows[2].ParentIndex)
	require.Equal(t, domain.BlockTypeParagraph, rows[3].Type)
	require.Equal(t, 2, rows[3].ParentIndex)

	// 兄弟の position は辞書順で増える（トップレベルの 2 行）。
	require.Less(t, rows[0].Position, rows[1].Position)
}

func Test_doc組み立て_兄弟をposition順に並べ直す(t *testing.T) {
	inline := `[{"type":"text","text":"x"}]`
	// わざと position の逆順・親子バラバラの順で渡す（DB の ORDER BY に頼らない検証）。
	blocks := []domain.Block{
		{ID: "b2", Type: domain.BlockTypeParagraph, Position: "a1", Attrs: "{}", Inline: &inline},
		{ID: "b1", Type: domain.BlockTypeParagraph, Position: "a0", Attrs: "{}", Inline: &inline},
	}
	tree, err := treeFromBlocks(blocks)
	require.NoError(t, err)
	rows, err := flattenPageDoc(tree)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	doc, err := renderPageDoc(tree)
	require.NoError(t, err)
	requireJSONEq(t, `{"type":"doc","content":[
		{"type":"paragraph","content":[{"type":"text","text":"x"}]},
		{"type":"paragraph","content":[{"type":"text","text":"x"}]}
	]}`, doc)
}

func Test_doc組み立て_親が見つからない行はエラー(t *testing.T) {
	missing := "no-such-parent"
	blocks := []domain.Block{
		{ID: "b1", Type: domain.BlockTypeParagraph, Position: "a0", Attrs: "{}", ParentID: &missing},
	}
	_, err := treeFromBlocks(blocks)
	require.Error(t, err)
}

func Test_doc組み立て_未知のtypeの行はエラー(t *testing.T) {
	blocks := []domain.Block{
		{ID: "b1", Type: domain.BlockType("iframe"), Position: "a0", Attrs: "{}"},
	}
	_, err := treeFromBlocks(blocks)
	require.True(t, errors.Is(err, ErrPageDocUnknownNodeType))
}
