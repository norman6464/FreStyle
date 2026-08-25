package domain

import "strings"

// EmailTrimCutset は正規形が email の前後から落とす空白文字の集合。
//
// PostgreSQL 側の btrim(email, E'\t\n\x0B\f\r ') と同じ集合をここで一度だけ決める。
// strings.TrimSpace は unicode.IsSpace（U+0085 / U+00A0 / U+2000 台なども含む）で畳むため
// btrim では表せず、アプリと DB で「前後の空白」の定義がずれる。ずれると、空白付きで保存された
// 既存行がアプリの正規形では引けない・一意索引では別キーとして両方作れる、という食い違いが残る。
// 落とす文字を明示列挙して両側を byte 単位で一致させ、その一致は結合テスト
// （TestEmailNormalForm_Integration）が実 PostgreSQL に同じ入力を通して固定する。
const EmailTrimCutset = "\t\n\v\f\r "

// NormalizeEmail はメールアドレスを突き合わせ・保存の正規形へ畳む。
//
// 正規形は「前後の空白（EmailTrimCutset）を落として小文字化した値」ひとつだけで、保存も比較も
// これを通す。比較に strings.EqualFold のような「畳んでから比べる」方式を使うと、畳めば同じだが
// バイト列が違う 2 つの値（"ops@example.com" と "OPS@example.com"）がアプリでは同一・
// DB の一意索引では別行になり、同じ人として弾かれるはずの行が両方作れてしまう。
// アプリ側の同一性と DB 側の一意性を同じ正規形の上に乗せるため、比較の前に必ずここを通す。
//
// 畳み方は PostgreSQL の lower(btrim(email, ...)) と揃えている（users の一意索引・検索・
// 招待の照会がこの式で張られる）。EqualFold の単純フォールドとは違い U+017F(ſ) は 's' に
// 畳まれない。どちらの側も畳まないので両者は別のアドレスとして扱われる
// （別物を同じと見なして特権を通す事故が起きない側に倒す）。
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.Trim(email, EmailTrimCutset))
}
