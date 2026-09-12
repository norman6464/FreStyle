package domain

import "strings"

// EmailTrimCutset は正規形が email の前後から落とす空白文字の集合。PostgreSQL 側の
// btrim(email, E'\t\n\x0B\f\r ') と同じ集合をここで一度だけ決める。strings.TrimSpace は
// unicode.IsSpace（U+0085 / U+00A0 等も含む）で畳むため btrim とはずれ、アプリと DB で
// 「前後の空白」の定義が食い違う。落とす文字を明示列挙して両側を byte 単位で一致させ、その
// 一致は結合テスト（TestEmailNormalForm_Integration）が実 PostgreSQL で固定する。
const EmailTrimCutset = "\t\n\v\f\r "

// NormalizeEmail はメールアドレスを突き合わせ・保存の正規形へ畳む。
//
// 正規形は「前後の空白（EmailTrimCutset）を落として小文字化した値」ひとつだけで、保存も比較も
// これを通す。strings.EqualFold のような「畳んでから比べる」方式だと、畳めば同じだがバイト列が
// 違う 2 つの値（"ops@example.com" と "OPS@example.com"）がアプリでは同一・DB の一意索引では
// 別行になり、同じ人として弾かれるはずの行が両方作れてしまう。
//
// 畳み方は PostgreSQL の lower(btrim(email, ...)) と揃える（users の一意索引・検索・招待の照会が
// この式で張られる）。EqualFold と違い U+017F(ſ) は 's' に畳まれない — どちらの側も畳まないので
// 別のアドレスとして扱われる（別物を同じと見なして特権を通す事故を避ける側に倒す）。
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.Trim(email, EmailTrimCutset))
}
