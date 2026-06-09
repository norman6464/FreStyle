package persistence

import "math"

// toInt64ID は domain の uint64 id を DB の bigint(int64) へ変換する。
//
// id は採番シーケンス由来で常に非負・int64 範囲内のため実際には溢れないが、CodeQL / gosec の
// 整数オーバーフロー検知（unsigned 64bit → 符号付き 64bit の上限未チェック変換）を満たすために
// 上限を明示チェックする。範囲外（実質あり得ない）の場合は ok=false を返し、呼び元は
// 「該当レコードなし」として扱う（存在し得ない id だから）。
func toInt64ID(id uint64) (int64, bool) {
	if id > math.MaxInt64 {
		return 0, false
	}
	return int64(id), true
}
