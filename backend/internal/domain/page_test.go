package domain_test

import (
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test_ページアイコン_絵文字として保存できる形 は domain.PageIcon.Valid() の境界を固定する。
//
// 「厳密に 1 grapheme か」までは見ない（依存を増やさず、見た目の一意性はピッカー側の
// 選択肢で担保する）ため、複数の絵文字を ZWJ で繋いだ家族の絵文字も許容する。
func Test_ページアイコン_絵文字として保存できる形(t *testing.T) {
	controlChar := string([]byte{0x00})
	cases := []struct {
		name  string
		icon  domain.PageIcon
		valid bool
	}{
		{
			name:  "単一の絵文字は OK",
			icon:  domain.PageIcon{Type: domain.PageIconTypeEmoji, Value: "📘"},
			valid: true,
		},
		{
			name:  "ZWJ で繋いだ家族の絵文字は OK（複数 code point でも 1 つの見た目）",
			icon:  domain.PageIcon{Type: domain.PageIconTypeEmoji, Value: "👨‍👩‍👧‍👦"},
			valid: true,
		},
		{
			name:  "空文字は NG",
			icon:  domain.PageIcon{Type: domain.PageIconTypeEmoji, Value: ""},
			valid: false,
		},
		{
			name:  "種類が emoji 以外は NG",
			icon:  domain.PageIcon{Type: domain.PageIconType("url"), Value: "📘"},
			valid: false,
		},
		{
			name:  "空白を含むと NG",
			icon:  domain.PageIcon{Type: domain.PageIconTypeEmoji, Value: "📘 "},
			valid: false,
		},
		{
			name:  "17 rune は NG（上限は 16 rune）",
			icon:  domain.PageIcon{Type: domain.PageIconTypeEmoji, Value: "aaaaaaaaaaaaaaaaa"},
			valid: false,
		},
		{
			name:  "制御文字を含むと NG",
			icon:  domain.PageIcon{Type: domain.PageIconTypeEmoji, Value: "📘" + controlChar},
			valid: false,
		},
		{
			name:  "壊れた UTF-8 は NG",
			icon:  domain.PageIcon{Type: domain.PageIconTypeEmoji, Value: string([]byte{0xff, 0xfe})},
			valid: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.valid, tc.icon.Valid())
		})
	}
}

// Test_ページアイコン_16runeちょうどは境界内 は 16 rune 上限の端を固定する
// （17 rune だけを NG にする、が上のテーブルで唯一の不正点になるようにするため）。
func Test_ページアイコン_16runeちょうどは境界内(t *testing.T) {
	icon := domain.PageIcon{Type: domain.PageIconTypeEmoji, Value: "aaaaaaaaaaaaaaaa"}
	assert.True(t, icon.Valid())
}
