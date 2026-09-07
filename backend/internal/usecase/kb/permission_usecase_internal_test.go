package kb

import (
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FRESTYLE-434 段 4: 検索結果の抜粋計算（computeSearchExcerpt / runeIndex）と、
// title 一致・body 一致の判定（buildSearchViewablePageResult）の単体テスト。
// どちらも unexported なのでこのパッケージ内（package kb）に置く。

func Test_抜粋計算(t *testing.T) {
	t.Run("ヒット位置の前後30文字程度を切り出す", func(t *testing.T) {
		// 前後にそれぞれ 40 文字ずつ（窓の 30 より多い）の日本語を置き、
		// 窓が実際に切り詰められることを確かめる。
		before := strings.Repeat("あ", 40)
		after := strings.Repeat("い", 40)
		body := before + "検索対象の言葉" + after
		excerpt, start, length, found := computeSearchExcerpt(body, "対象")
		require.True(t, found)
		// 窓は前後 30 rune。ヒット（"対象"）の直前は "あ"×30、直後は "の言葉" + "い"×27。
		assert.Equal(t, 30, start, "ヒット位置の直前に残る rune 数はちょうど窓の大きさ")
		assert.Equal(t, 2, length, "\"対象\" は2 rune")
		gotMatch := []rune(excerpt)[start : start+length]
		assert.Equal(t, "対象", string(gotMatch), "MatchStart/MatchLen は excerpt の中でヒットを指す")
	})

	t.Run("queryが先頭に近い場合は窓がマイナスにならない", func(t *testing.T) {
		body := "検索対象" + strings.Repeat("あ", 40)
		excerpt, start, length, found := computeSearchExcerpt(body, "検索")
		require.True(t, found)
		assert.Equal(t, 0, start, "本文の先頭より前は無いので0")
		assert.Equal(t, 2, length)
		assert.True(t, strings.HasPrefix(excerpt, "検索"), "excerpt はヒットそのものから始まる")
	})

	t.Run("queryが末尾に近い場合は末尾までで切り詰める", func(t *testing.T) {
		body := strings.Repeat("あ", 40) + "検索対象"
		excerpt, start, length, found := computeSearchExcerpt(body, "対象")
		require.True(t, found)
		assert.Equal(t, 2, length)
		gotMatch := []rune(excerpt)[start : start+length]
		assert.Equal(t, "対象", string(gotMatch))
		assert.True(t, strings.HasSuffix(excerpt, "対象"), "excerpt はヒットそのもので終わる（本文の末尾を超えて切り詰めない）")
	})

	t.Run("大文字小文字を区別しない", func(t *testing.T) {
		body := "This is a Docker guide."
		excerpt, start, length, found := computeSearchExcerpt(body, "docker")
		require.True(t, found)
		gotMatch := []rune(excerpt)[start : start+length]
		assert.Equal(t, "Docker", string(gotMatch), "元の大文字小文字はそのまま保つ（比較だけ大文字小文字を無視する）")
	})

	t.Run("一致しなければfoundはfalse", func(t *testing.T) {
		_, _, _, found := computeSearchExcerpt("何も一致しない本文", "存在しない語")
		assert.False(t, found)
	})

	t.Run("空のqueryや空のbodyはfoundにならない", func(t *testing.T) {
		_, _, _, found := computeSearchExcerpt("本文", "")
		assert.False(t, found)
		_, _, _, found = computeSearchExcerpt("", "query")
		assert.False(t, found)
	})

	t.Run("マルチバイト文字が混在してもrune境界を壊さない", func(t *testing.T) {
		// 絵文字（サロゲートペア相当のコードポイント）を本文に混ぜても、
		// 切り出した excerpt が有効な UTF-8 のまま・rune 単位の位置がずれないこと。
		body := "設計メモ 📘 " + strings.Repeat("文", 35) + " Docker の手順 " + strings.Repeat("字", 35)
		excerpt, start, length, found := computeSearchExcerpt(body, "docker")
		require.True(t, found)
		require.True(t, len(excerpt) > 0)
		assert.True(t, utf8Valid(excerpt), "切り出した excerpt は有効な UTF-8 のまま")
		gotMatch := []rune(excerpt)[start : start+length]
		assert.Equal(t, "Docker", string(gotMatch))
	})
}

func utf8Valid(s string) bool {
	for _, r := range s {
		if r == '�' {
			return false
		}
	}
	return true
}

func Test_runeIndex(t *testing.T) {
	cases := []struct {
		name     string
		haystack string
		needle   string
		want     int
	}{
		{"先頭で一致", "abcdef", "abc", 0},
		{"末尾で一致", "abcdef", "def", 3},
		{"中間で一致", "abcdef", "cd", 2},
		{"一致しない", "abcdef", "xyz", -1},
		{"needleがhaystackより長い", "ab", "abc", -1},
		{"needleが空", "abc", "", -1},
		{"日本語", "設計メモの本文", "メモ", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runeIndex([]rune(tc.haystack), []rune(tc.needle))
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_検索結果の一致判定(t *testing.T) {
	t.Run("titleが一致していればexcerptを計算しない", func(t *testing.T) {
		row := repository.PageSearchViewFact{
			PageWithViewFacts: repository.PageWithViewFacts{
				Page: domain.Page{ID: "p-1", Title: "Docker 手順"},
			},
			Body: "本文には Docker という語は出てこない",
		}
		got := buildSearchViewablePageResult(row, "docker")
		assert.Equal(t, SearchMatchFieldTitle, got.MatchField)
		assert.Empty(t, got.Excerpt)
		assert.Zero(t, got.MatchStart)
		assert.Zero(t, got.MatchLen)
	})

	t.Run("titleが一致せずbodyが一致していれば抜粋を計算する", func(t *testing.T) {
		row := repository.PageSearchViewFact{
			PageWithViewFacts: repository.PageWithViewFacts{
				Page: domain.Page{ID: "p-1", Title: "設計メモ"},
			},
			Body: "この段落には Docker の使い方が書いてある",
		}
		got := buildSearchViewablePageResult(row, "docker")
		assert.Equal(t, SearchMatchFieldBody, got.MatchField)
		require.NotEmpty(t, got.Excerpt)
		gotMatch := []rune(got.Excerpt)[got.MatchStart : got.MatchStart+got.MatchLen]
		assert.Equal(t, "Docker", string(gotMatch))
	})

	t.Run("SQL側はbody一致でもGo側の判定で見つからない場合は抜粋なしのbody一致にする", func(t *testing.T) {
		// SQL 側の ILIKE と Go 側の判定がずれる状況（実運用では起きない想定の安全弁）を
		// 直接構成して確かめる。
		row := repository.PageSearchViewFact{
			PageWithViewFacts: repository.PageWithViewFacts{
				Page: domain.Page{ID: "p-1", Title: "設計メモ"},
			},
			Body: "",
		}
		got := buildSearchViewablePageResult(row, "docker")
		assert.Equal(t, SearchMatchFieldBody, got.MatchField)
		assert.Empty(t, got.Excerpt)
	})
}
