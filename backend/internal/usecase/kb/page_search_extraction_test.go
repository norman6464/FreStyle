package kb

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 本文検索・逆リンクの抽出ロジック（extractPageBodyText /
// extractPageLinks）の単体テスト。どちらも parsePageDoc が返す kbDocNode の木を対象にする
// （flattenPageDoc を経由しない — id の再採番は本文検索の抽出には関係が無いため）。

func Test_本文プレーンテキスト抽出(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		want string
	}{
		{
			name: "1段落",
			doc:  `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"こんにちは"}]}]}`,
			want: "こんにちは",
		},
		{
			name: "複数ブロックは改行区切り",
			doc: `{"type":"doc","content":[
				{"type":"paragraph","content":[{"type":"text","text":"1つ目"}]},
				{"type":"paragraph","content":[{"type":"text","text":"2つ目"}]}
			]}`,
			want: "1つ目\n2つ目",
		},
		{
			name: "空のブロックは区切りを増やさない",
			doc: `{"type":"doc","content":[
				{"type":"paragraph","content":[{"type":"text","text":"1つ目"}]},
				{"type":"paragraph"},
				{"type":"paragraph","content":[{"type":"text","text":"2つ目"}]}
			]}`,
			want: "1つ目\n2つ目",
		},
		{
			name: "同じブロック内の複数textノードは区切り無しで連結する",
			doc: `{"type":"doc","content":[
				{"type":"paragraph","content":[
					{"type":"text","marks":[{"type":"bold"}],"text":"強調"},
					{"type":"text","text":"通常"}
				]}
			]}`,
			want: "強調通常",
		},
		{
			name: "pageRefノードは本文に寄与しない",
			doc: fmt.Sprintf(`{"type":"doc","content":[
				{"type":"paragraph","content":[
					{"type":"text","text":"前"},
					{"type":"pageRef","attrs":{"pageId":%q}},
					{"type":"text","text":"後"}
				]}
			]}`, uuid.NewString()),
			want: "前後",
		},
		{
			name: "ネストしたリスト・引用の中のテキストも拾う",
			doc: `{"type":"doc","content":[
				{"type":"bulletList","content":[
					{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"項目1"}]}]},
					{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"項目2"}]}]}
				]},
				{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"引用"}]}]}
			]}`,
			want: "項目1\n項目2\n引用",
		},
		{
			name: "本文が無いdocは空文字",
			doc:  `{"type":"doc","content":[]}`,
			want: "",
		},
		{
			name: "画像や区切り線など inline を持たないブロックは何も出さない",
			doc: `{"type":"doc","content":[
				{"type":"image","attrs":{"src":"https://example.com/a.png"}},
				{"type":"horizontalRule"},
				{"type":"paragraph","content":[{"type":"text","text":"本文"}]}
			]}`,
			want: "本文",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := parsePageDoc(tc.doc)
			require.NoError(t, err)
			got := extractPageBodyText(tree)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_pageRef収集(t *testing.T) {
	t.Run("異なるブロックの参照はそれぞれ1行になる", func(t *testing.T) {
		target1 := uuid.NewString()
		target2 := uuid.NewString()
		doc := fmt.Sprintf(`{"type":"doc","content":[
			{"type":"paragraph","content":[{"type":"pageRef","attrs":{"pageId":%q}}]},
			{"type":"paragraph","content":[{"type":"pageRef","attrs":{"pageId":%q}}]}
		]}`, target1, target2)
		tree, err := parsePageDoc(doc)
		require.NoError(t, err)
		require.Len(t, tree, 2, "トップレベルは段落2つ")

		links := extractPageLinks(tree)
		require.Len(t, links, 2)
		assert.Equal(t, tree[0].ID, links[0].SourceBlockID)
		assert.Equal(t, target1, links[0].TargetPageID)
		assert.Equal(t, tree[1].ID, links[1].SourceBlockID)
		assert.Equal(t, target2, links[1].TargetPageID)
	})

	t.Run("同じブロックが同じページを複数回参照しても1行に畳む", func(t *testing.T) {
		target := uuid.NewString()
		doc := fmt.Sprintf(`{"type":"doc","content":[
			{"type":"paragraph","content":[
				{"type":"pageRef","attrs":{"pageId":%q}},
				{"type":"text","text":"の間に"},
				{"type":"pageRef","attrs":{"pageId":%q}}
			]}
		]}`, target, target)
		tree, err := parsePageDoc(doc)
		require.NoError(t, err)
		require.Len(t, tree, 1)

		links := extractPageLinks(tree)
		require.Len(t, links, 1, "同じ (ブロック, 参照先) の組は1行に畳む")
		assert.Equal(t, tree[0].ID, links[0].SourceBlockID)
		assert.Equal(t, target, links[0].TargetPageID)
	})

	t.Run("UUIDとして読めないpageIdは無視する", func(t *testing.T) {
		doc := `{"type":"doc","content":[
			{"type":"paragraph","content":[{"type":"pageRef","attrs":{"pageId":"not-a-uuid"}}]}
		]}`
		tree, err := parsePageDoc(doc)
		require.NoError(t, err)
		links := extractPageLinks(tree)
		assert.Empty(t, links)
	})

	t.Run("ネストしたブロックの中のpageRefも拾う", func(t *testing.T) {
		target := uuid.NewString()
		doc := fmt.Sprintf(`{"type":"doc","content":[
			{"type":"bulletList","content":[
				{"type":"listItem","content":[
					{"type":"paragraph","content":[{"type":"pageRef","attrs":{"pageId":%q}}]}
				]}
			]}
		]}`, target)
		tree, err := parsePageDoc(doc)
		require.NoError(t, err)
		links := extractPageLinks(tree)
		require.Len(t, links, 1)
		assert.Equal(t, target, links[0].TargetPageID)
	})

	// kbPageRefMaxResolve（100）は「参照先ページの種類数」の天井。101 個の異なる参照先が
	// それぞれ別ブロックから 1 回ずつ参照される doc を作り、101 個目以降の**新しい種類**が
	// 取りこぼされることを確かめる（pageRefCollector と同じ考え方の上限）。
	t.Run("参照先ページの種類数は上限100件で切り詰める", func(t *testing.T) {
		const total = kbPageRefMaxResolve + 5
		targets := make([]string, total)
		for i := range targets {
			targets[i] = uuid.NewString()
		}
		var blocks strings.Builder
		for i, id := range targets {
			if i > 0 {
				blocks.WriteByte(',')
			}
			fmt.Fprintf(&blocks, `{"type":"paragraph","content":[{"type":"pageRef","attrs":{"pageId":%q}}]}`, id)
		}
		doc := `{"type":"doc","content":[` + blocks.String() + `]}`
		tree, err := parsePageDoc(doc)
		require.NoError(t, err)
		require.Len(t, tree, total)

		links := extractPageLinks(tree)
		assert.Len(t, links, kbPageRefMaxResolve, "天井を超えた分の新しい参照先は取りこぼす")

		gotTargets := make(map[string]struct{}, len(links))
		for _, l := range links {
			gotTargets[l.TargetPageID] = struct{}{}
		}
		// 最初の kbPageRefMaxResolve 件（文書順）だけが残っていること。
		for i, id := range targets {
			_, ok := gotTargets[id]
			if i < kbPageRefMaxResolve {
				assert.True(t, ok, "先頭 %d 番目の参照先 %s は残るはず", i, id)
			} else {
				assert.False(t, ok, "天井を超えた参照先 %s は残らないはず", id)
			}
		}
	})
}
