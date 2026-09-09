package ticket_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tkDocSample = `{
  "type": "doc",
  "content": [
    {"type": "paragraph", "content": [
      {"type": "text", "text": "本文の一部です。"},
      {"type": "pageRef", "attrs": {"pageId": "01a00000-0000-7000-8000-0000000000e1", "title": "設計メモ"}},
      {"type": "text", "text": " と "},
      {"type": "ticketRef", "attrs": {"ticketId": "01a00000-0000-7000-8000-0000000000e2", "title": "関連チケット"}}
    ]},
    {"type": "paragraph", "content": [
      {"type": "text", "text": "同じチケットへの再参照"},
      {"type": "ticketRef", "attrs": {"ticketId": "01a00000-0000-7000-8000-0000000000e2", "title": "関連チケット"}}
    ]}
  ]
}`

func Test_本文からの参照抽出(t *testing.T) {
	pageIDs, ticketIDs, err := ticket.ExtractDocRefs([]byte(tkDocSample))
	require.NoError(t, err)

	assert.Equal(t, []string{"01a00000-0000-7000-8000-0000000000e1"}, pageIDs)
	assert.Equal(t, []string{"01a00000-0000-7000-8000-0000000000e2"}, ticketIDs, "同じチケットへの2回目の参照は1件に畳む")
}

func Test_本文からの参照抽出_参照が無ければ空(t *testing.T) {
	pageIDs, ticketIDs, err := ticket.ExtractDocRefs([]byte(`{"type":"doc","content":[]}`))
	require.NoError(t, err)
	assert.Empty(t, pageIDs)
	assert.Empty(t, ticketIDs)
}

func Test_本文からの参照抽出_壊れたJSONはエラー(t *testing.T) {
	_, _, err := ticket.ExtractDocRefs([]byte(`not json`))
	require.Error(t, err)
}

func Test_本文からの参照抽出_不正なUUIDは無視する(t *testing.T) {
	doc := `{"type":"doc","content":[{"type":"paragraph","content":[
		{"type":"pageRef","attrs":{"pageId":"not-a-uuid"}},
		{"type":"ticketRef","attrs":{"ticketId":""}}
	]}]}`
	pageIDs, ticketIDs, err := ticket.ExtractDocRefs([]byte(doc))
	require.NoError(t, err)
	assert.Empty(t, pageIDs)
	assert.Empty(t, ticketIDs)
}

// StripDocRefTitles は保存前の doc から pageRef/ticketRef の title を取り除く。
// title は読み手ごとに読み出し時へ解決する派生値で、保存してはいけない
// （page 側の StripPageRefTitles と同じ理由。設計 Ⅳ-E）。
func Test_本文の参照タイトルを剥がす(t *testing.T) {
	stripped, err := ticket.StripDocRefTitles([]byte(tkDocSample))
	require.NoError(t, err)

	assert.NotContains(t, string(stripped), "設計メモ")
	assert.NotContains(t, string(stripped), "関連チケット")
	// 参照先の id 自体は保存する（剥がすのは title だけ）。
	assert.Contains(t, string(stripped), "01a00000-0000-7000-8000-0000000000e1")
	assert.Contains(t, string(stripped), "01a00000-0000-7000-8000-0000000000e2")
	// 本文のプレーンテキストは変わらない。
	assert.Contains(t, string(stripped), "本文の一部です。")
}

func Test_本文の参照タイトルを剥がす_参照が無ければそのまま(t *testing.T) {
	doc := []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"ただの文章"}]}]}`)
	stripped, err := ticket.StripDocRefTitles(doc)
	require.NoError(t, err)
	assert.JSONEq(t, string(doc), string(stripped))
}

// BuildPlainText は検索用の派生値。pageRef/ticketRef の属性は含めず、text ノードの
// 内容だけを集める（設計 Ⅳ-E: 検索は plain_text の ILIKE）。
func Test_プレーンテキストの構築(t *testing.T) {
	got := ticket.BuildPlainText([]byte(tkDocSample))
	assert.Contains(t, got, "本文の一部です。")
	assert.Contains(t, got, "同じチケットへの再参照")
	assert.NotContains(t, got, "設計メモ", "pageRef の title は本文に含めない")
	assert.NotContains(t, got, "0000000000e1", "id そのものも本文に含めない")
}

func Test_プレーンテキストの構築_壊れたJSONは空文字(t *testing.T) {
	assert.Equal(t, "", ticket.BuildPlainText([]byte(`not json`)))
}

// DocsEqual はバイト列ではなく木として比べる。JSON のキー順が違っても意味が同じなら
// 等しいと判定する（UpdateTicketUseCase が「本文が変わったか」を誤検知しないための土台）。
func Test_doc等価性の判定はキー順に依存しない(t *testing.T) {
	a := []byte(`{"type":"doc","content":[]}`)
	b := []byte(`{"content":[],"type":"doc"}`)
	assert.True(t, ticket.DocsEqual(a, b))

	c := []byte(`{"type":"doc","content":[{"type":"paragraph"}]}`)
	assert.False(t, ticket.DocsEqual(a, c), "中身が違えば等しくない")

	assert.False(t, ticket.DocsEqual(a, []byte(`not json`)), "壊れたJSONは常に等しくない")
	assert.False(t, ticket.DocsEqual([]byte(`not json`), a))
}
