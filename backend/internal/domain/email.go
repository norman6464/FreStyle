package domain

import "strings"

// NormalizeEmail はメールアドレスを突き合わせ・保存の正規形へ畳む。
//
// 正規形は「前後の空白を落として小文字化した値」ひとつだけで、保存も比較もこれを通す。
// 比較に strings.EqualFold のような「畳んでから比べる」方式を使うと、畳めば同じだが
// バイト列が違う 2 つの値（"ops@example.com" と "OPS@example.com"）がアプリでは同一・
// DB の一意索引では別行になり、同じ人として弾かれるはずの行が両方作れてしまう。
// アプリ側の同一性と DB 側の一意性を同じ正規形の上に乗せるため、比較の前に必ずここを通す。
//
// 畳み方は PostgreSQL の lower() と揃えている（users の一意索引が lower(email) で張られる）。
// EqualFold の単純フォールドとは違い U+017F(ſ) は 's' に畳まれない。どちらの側も畳まないので
// 両者は別のアドレスとして扱われる（別物を同じと見なして特権を通す事故が起きない側に倒す）。
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
