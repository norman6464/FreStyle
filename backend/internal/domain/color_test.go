package domain_test

import (
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_色の正規化(t *testing.T) {
	assert.Equal(t, "#2f6b47", domain.NormalizeHexColor("#2F6B47"))
	assert.Equal(t, "#2f6b47", domain.NormalizeHexColor("#2f6b47"), "既に小文字なら変わらない")
	assert.Equal(t, "", domain.NormalizeHexColor(""))
	assert.Equal(t, " #2f6b47 ", domain.NormalizeHexColor(" #2F6B47 "),
		"空白は落とさない。落とすと不正な入力を黙って直すことになり、"+
			"打ち間違いがそのまま保存される")
}

func Test_色として保存してよい形(t *testing.T) {
	for _, ok := range []string{"#2f6b47", "#000000", "#ffffff", "#0a1b2c"} {
		assert.True(t, domain.ValidHexColor(ok), ok)
	}

	cases := []struct {
		name string
		in   string
	}{
		{name: "空", in: ""},
		{name: "大文字（正規化を通していない）", in: "#2F6B47"},
		{name: "# が無い", in: "2f6b47"},
		{name: "3 桁の短縮形", in: "#fff"},
		{name: "桁が多い", in: "#2f6b47a"},
		{name: "桁が足りない", in: "#2f6b4"},
		{name: "16 進でない文字", in: "#2f6b4g"},
		{name: "前後に空白", in: " #2f6b47"},
		{name: "色名", in: "red"},
		{name: "rgb 記法", in: "rgb(0,0,0)"},
		{name: "改行を含む（^$ ではなく \\A\\z 相当であること）", in: "#2f6b47\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, domain.ValidHexColor(tc.in))
		})
	}
}

// 正規化してから検証する、が呼び出し側の手順。大文字の色は DB の CHECK
// （^#[0-9a-f]{6}$）が拒むので、正規化を忘れると保存時に落ちる。
func Test_色は正規化してから検証すると通る(t *testing.T) {
	assert.False(t, domain.ValidHexColor("#2F6B47"))
	assert.True(t, domain.ValidHexColor(domain.NormalizeHexColor("#2F6B47")))
}
