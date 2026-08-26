package persistence

import (
	"errors"
	"fmt"
	"math"
)

// toInt64ID は domain の uint64 id を DB の bigint(int64) へ変換する。
//
// id は採番シーケンス由来で常に非負・int64 範囲内のため実際には溢れないが、CodeQL / gosec の
// 整数オーバーフロー検知（unsigned 64bit → 符号付き 64bit の上限未チェック変換）を満たすために
// 上限を明示チェックする。範囲外（実質あり得ない）の場合は ok=false を返す。
//
// 呼び元の扱いは読み書きで分かれる。読み取りは「存在し得ない id」= 該当レコードなしで良いが、
// 書き込みは 1 行も書けていないので errOutOfRangeID を返す（nil を返すと呼び出し側が
// 作成・更新できたと誤認する）。
func toInt64ID(id uint64) (int64, bool) {
	if id > math.MaxInt64 {
		return 0, false
	}
	return int64(id), true
}

// errOutOfRangeID は uint64 の id が DB の bigint(int64) に収まらないことを表す。
var errOutOfRangeID = errors.New("id が bigint の範囲を超えている")

// outOfRangeIDError は書き込みを諦めた列を添えて errOutOfRangeID を返す。
func outOfRangeIDError(column string, id uint64) error {
	return fmt.Errorf("%w: %s=%d", errOutOfRangeID, column, id)
}
