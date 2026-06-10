package usecase

import "testing"

// normalizeOutput は採点の出力比較で使う正規化。学習者に見えない行末空白で
// 不合格にしないことを検証する（`fmt.Printf("%d \n")` の余分なスペース等）。
func TestNormalizeOutput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"行末スペースを除去", "田中: 23 \n佐藤: 30 \n", "田中: 23\n佐藤: 30"},
		{"行末タブを除去", "a\t\nb\t", "a\nb"},
		{"CRLF を LF に統一", "a\r\nb\r\n", "a\nb"},
		{"先頭インデントは保持", "  x\ny ", "  x\ny"},
		{"末尾の空行を除去", "a\nb\n\n  \n", "a\nb"},
		{"完全一致はそのまま", "田中: 23\n佐藤: 30", "田中: 23\n佐藤: 30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeOutput(tt.in); got != tt.want {
				t.Errorf("normalizeOutput(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
