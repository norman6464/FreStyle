package oidc

import "sort"

// RolesFromClaim は「役割の一覧」を表すクレームから役割名を取り出す。
//
// 役割の入れ方は発行者ごとに違う。よくあるのは次の 3 つで、どれで来ても読めるようにする。
//
//	["admin", "editor"]                      文字列の配列
//	"admin"                                  文字列 1 つ
//	{"admin": {...}, "editor": {...}}        役割名を鍵にした表
//
// 3 つ目は Zitadel のプロジェクトロールの形で、値の側には役割が有効な組織が入る。
// ここでは鍵（役割名）だけを見る。
//
// 配列だと決めつけて型アサーションを書くと、表で来た瞬間に「役割ゼロ」になる。
// **弾かれるのではなく静かに権限が消える**ので、画面には管理機能が出ないのに
// エラーもログも残らない、という一番気づきにくい壊れ方をする。
func RolesFromClaim(v any) []string {
	switch raw := v.(type) {
	case string:
		if raw == "" {
			return nil
		}
		return []string{raw}
	case []any:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case map[string]any:
		out := make([]string, 0, len(raw))
		for name := range raw {
			if name != "" {
				out = append(out, name)
			}
		}
		// map の反復順は毎回変わる。並びが揺れると、ログや比較のたびに
		// 差分が出て読みにくいので固定する。
		sort.Strings(out)
		return out
	default:
		return nil
	}
}

// HasRole は roles に name が含まれるかを返す。
func HasRole(roles []string, name string) bool {
	for _, r := range roles {
		if r == name {
			return true
		}
	}
	return false
}
