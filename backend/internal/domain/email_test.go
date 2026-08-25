package domain

import "testing"

// NormalizeEmail の正規形を固定する。ここが緩むと「アプリでは同一・DB では別行」の
// 食い違いが戻る（一意索引は lower(btrim(email, EmailTrimCutset)) で張られている）。
func Test_NormalizeEmail_前後空白を落として小文字化する(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"そのまま", "ops@example.com", "ops@example.com"},
		{"大文字は畳む", "OPS@Example.com", "ops@example.com"},
		{"前後の空白は落とす", "  ops@example.com\t", "ops@example.com"},
		{"改行・復帰・垂直タブ・改ページも落とす", "\r\n\v\fops@example.com\v\f\r\n", "ops@example.com"},
		{"空文字は空文字のまま", "   ", ""},
		// ASCII 空白以外の Unicode 空白は落とさない。btrim(email, EmailTrimCutset) で
		// 同じ集合を表せないものを Go 側だけが落とすと、DB の索引キーとずれるため
		// （落とさなければ別アドレスとして扱われるだけで、アプリと DB の見解は一致する）。
		{"NO-BREAK SPACE は落とさない", "\u00A0ops@example.com", "\u00a0ops@example.com"},
		{"NEXT LINE(U+0085) は落とさない", "ops@example.com\u0085", "ops@example.com\u0085"},
		// PostgreSQL の lower() と同じ畳み方（U+212A KELVIN SIGN は 'k' になる）。
		{"KELVIN SIGN は k に畳む", "\u212Aops@example.com", "kops@example.com"},
		// strings.EqualFold は U+017F(ſ) を 's' に畳むが、小文字化では畳まれない。
		// 別のアドレスとして扱う（別物を同じと見なして特権を通す側に倒さない）。
		{"LATIN SMALL LETTER LONG S は s に畳まない", "\u017Fops@example.com", "\u017Fops@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeEmail(tt.input); got != tt.want {
				t.Fatalf("NormalizeEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
