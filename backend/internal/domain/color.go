package domain

import "strings"

// 色（#rrggbb）の正規化と検証。チケットの状態・種別・カテゴリー・節目が持つ表示色に使う。
//
// 保存する形を小文字だけに絞るのは、同じ色が `#2F6B47` と `#2f6b47` の 2 通りで
// 保存されるのを防ぐため。名前の一意（大文字小文字を無視した重複禁止）と違い、
// 色は「同じ色を 2 度保存しても困らない」ので DB 側は綴りだけを CHECK で縛り、
// 呼び出し側が保存前に NormalizeHexColor を通す分担にしている。

// HexColorLen は色の列幅（character varying(7)）。`#` + 6 桁。
const HexColorLen = 7

// NormalizeHexColor は色の表記ゆれを畳む（小文字にするだけ）。
//
// 空白は落とさない。落とすと `" #2f6b47 "` のような打ち間違いを黙って直して保存することに
// なり、入力欄の不備が表に出なくなる。形が違うものは ValidHexColor が拒む。
func NormalizeHexColor(s string) string {
	return strings.ToLower(s)
}

// ValidHexColor は色として保存してよい形かを返す。
//
// DB 側の CHECK（`^#[0-9a-f]{6}$`）と同じ判定。PostgreSQL の `~` は複数行の
// アンカーではないので改行を含む値は通らない。Go 側も同じく「全体が一致」で見る
// （長さを先に見ているので、末尾の改行は弾かれる）。
func ValidHexColor(s string) bool {
	if len(s) != HexColorLen || s[0] != '#' {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f':
		default:
			return false
		}
	}
	return true
}
