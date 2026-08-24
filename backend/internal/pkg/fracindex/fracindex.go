// Package fracindex は「順序」を整数の連番ではなく文字列キー（分数インデックス）で表す。
//
// なぜ連番ではないか: ページ / ブロックの並び替えは「1 行だけ動かす」操作なのに、
// 整数 sort_order だと挿入のたびに後続を全部振り直す UPDATE が必要になる。
// キーが文字列なら「隣り合う 2 キーの間」を必ず新しく作れるため、並び替えは
// 動かした 1 行の position を書き換えるだけで済む（同時編集でも衝突面積が小さい）。
//
// 使い方は [Between] だけ。prev / next にはリスト上の隣り合うキーを渡し、
// 返ってきたキーを新しい行の position に入れる。
//
// キー長について（既知の性質・将来の逃げ道）:
// 本実装は「2 つのキーの中点を取る」だけの素直な方式で、桁の大小を表すヘッダ（整数部）を持たない。
// そのため同じ隙間へ挿し続けると 6 回あたり 1 文字ずつ伸びる（100 回で 20 文字前後）。
// 通常のページ・ブロック数なら問題にならないが、1 つの親に数千件を末尾追加し続けると
// キーが数百文字に達する。position は「同じ親の中だけで意味を持つローカルな順序」なので、
// 長くなりすぎたら親単位で採番し直せば（UPDATE 1 回）縮められる。整数部を持つ方式へ
// 差し替える必要が出た場合も、同じく親単位の再採番で移行できる。
package fracindex

import (
	"errors"
	"fmt"
	"strings"
)

// digits は分数インデックスに使う文字集合（base62）。
//
// base62 を選んだ理由:
//   - この並び（0-9 → A-Z → a-z）は ASCII コード順と完全に一致する。つまり Go の文字列比較
//     （バイト比較）と PostgreSQL の C コレーションでの ORDER BY が同じ順序になる。
//     アプリと DB で並びがずれない、というのがこの実装の一番の前提条件になる
//     （DB 側は position 列を COLLATE "C" に固定して担保する。ロケール依存のコレーションだと
//     'a' < 'B' のような順序になり、バイト比較と食い違うため）。
//   - 1 文字あたり log2(62) ≒ 5.95 bit 分の細分ができるので、同じ場所への連続挿入でも
//     キー長が伸びにくい（base16 等より短く保てる）。
//   - URL・JSON にそのまま置ける安全な文字だけで構成されている。
const digits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// smallestDigit は文字集合の最小文字。「その桁が無い」ことの既定値として使う。
// digits[0] と一致していることはテストで固定する（定数式では文字列の添字を取れないため直書き）。
const smallestDigit byte = '0'

var (
	// ErrInvalidKey は文字集合外の文字を含むキー、または正規形でないキー（末尾が最小文字）。
	ErrInvalidKey = errors.New("fracindex: 不正なキー")
	// ErrOutOfOrder は prev >= next のときに返す。呼び出し側の並び順のバグを黙って通さない。
	ErrOutOfOrder = errors.New("fracindex: prev は next より小さい必要があります")
)

// Between は prev と next の「辞書順で厳密に間」にあるキーを返す。
//
// prev が空文字ならリストの先頭より前、next が空文字なら末尾より後を意味する
// （両方空なら最初の 1 件目のキーになる）。prev >= next のときは [ErrOutOfOrder] を返す。
//
// 返すキーは末尾が最小文字 '0' にならない正規形を保つ。'0' で終わるキー（例: "V0"）を作ると、
// その手前に挿す余地（"V" < x < "V0" を満たす x）が無くなり、以後 Between が失敗するため。
func Between(prev, next string) (string, error) {
	if err := validateKey(prev, "prev"); err != nil {
		return "", err
	}
	if err := validateKey(next, "next"); err != nil {
		return "", err
	}
	hasUpper := next != ""
	if hasUpper && prev >= next {
		return "", fmt.Errorf("%w: prev=%q next=%q", ErrOutOfOrder, prev, next)
	}
	return midpoint(prev, next, hasUpper)
}

// midpoint は a < 返り値 < b を満たすキーを作る。hasUpper=false は b が上限なし（末尾より後）の意味。
// a は空文字（下限なし）を許す。前提（a < b・両者が正規形）は [Between] が検証済み。
func midpoint(a, b string, hasUpper bool) (string, error) {
	if hasUpper {
		// 共通 prefix はそのまま持ち越し、最初に食い違う桁から採番する。
		// a 側が尽きている桁は「最小文字が並んでいる」とみなす（"V" は "V0" と同じ位置から始まる）。
		n := 0
		for n < len(b) {
			da := smallestDigit
			if n < len(a) {
				da = a[n]
			}
			if da != b[n] {
				break
			}
			n++
		}
		if n >= len(b) {
			// b が丸ごと a（最小文字で埋めたもの）の prefix になるケース。a < b かつ両者が
			// 正規形なら起こり得ないが、前提が崩れたまま黙って壊れた順序を返さないようエラーにする。
			return "", fmt.Errorf("%w: prev=%q next=%q の順序が矛盾しています", ErrOutOfOrder, a, b)
		}
		if n > 0 {
			var restA string
			if n < len(a) {
				restA = a[n:]
			}
			mid, err := midpoint(restA, b[n:], true)
			if err != nil {
				return "", err
			}
			return b[:n] + mid, nil
		}
	}

	// 先頭桁の「数値」を取る。a が空なら下限（最小文字の手前）、b が無いなら上限（最大文字の次）。
	digitA := 0
	if a != "" {
		digitA = strings.IndexByte(digits, a[0])
	}
	digitB := len(digits)
	if hasUpper {
		digitB = strings.IndexByte(digits, b[0])
	}

	if digitB-digitA > 1 {
		// 桁が 2 以上離れていれば、その中間の 1 文字だけで足りる（キーを伸ばさない）。
		return string(digits[(digitA+digitB+1)/2]), nil
	}

	// 先頭桁が隣り合っている場合。
	if hasUpper && len(b) > 1 {
		// b の先頭 1 文字は a より大きく b より小さい（b は 2 文字以上なので真に短い）。
		return b[:1], nil
	}
	// b が無い / b が 1 文字しかない場合は、a の先頭桁に留まったまま次の桁を掘る。
	var restA string
	if len(a) > 1 {
		restA = a[1:]
	}
	tail, err := midpoint(restA, "", false)
	if err != nil {
		return "", err
	}
	return string(digits[digitA]) + tail, nil
}

// validateKey は文字集合と正規形（末尾が最小文字でない）を検査する。空文字は端の意味なので許す。
func validateKey(key, name string) error {
	if key == "" {
		return nil
	}
	for i := 0; i < len(key); i++ {
		if strings.IndexByte(digits, key[i]) < 0 {
			return fmt.Errorf("%w: %s=%q に文字集合外の文字が含まれます", ErrInvalidKey, name, key)
		}
	}
	if key[len(key)-1] == smallestDigit {
		return fmt.Errorf("%w: %s=%q は末尾が %q（正規形ではありません）", ErrInvalidKey, name, key, string(smallestDigit))
	}
	return nil
}
