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
// # キーの構造
//
// キーは「整数部 + 小数部」でできている。整数部の先頭 1 文字が整数部の桁数を表すヘッダで、
// 残りが実際の桁。小数部は整数部のすぐ後ろに続く 0 文字以上の桁で、
// 「整数と整数の間」を掘るときだけ伸びる。
//
//	a0                  整数 0（ヘッダ 'a' = 整数部 2 文字）。最初の 1 件目がこれ
//	a1 a2 … az          整数 1..61。ここまで 2 文字
//	b00 … bzz           整数 62..3905。3 文字
//	c000 …              整数 3906..。4 文字
//	Zz Zy … Z0          整数 -1..-62。2 文字（負の側）
//	Yzz … Y00           整数 -63..-3906。3 文字
//	a0V                 整数 0 と整数 1 の間（小数部 "V"）
//
// ヘッダの割り当て:
//   - 小文字 'a'..'z' が非負、大文字 'A'..'Z' が負。ASCII では 'A'-'Z'(65-90) < 'a'-'z'(97-122)
//     なので、負の整数のキーは必ず非負のキーより前に並ぶ。
//   - 整数部の全長は 'a'→2, 'b'→3, … 'z'→27、'Z'→2, 'Y'→3, … 'A'→27。
//     負側はヘッダが小さいほど長い（＝より小さい整数を表す）ので、桁数が変わっても
//     「文字列の辞書順」と「整数としての大小」が一致する。
//   - 表現できる整数の範囲は "A"+"0"×26 〜 "z"+"z"×26。実運用で使い切ることはない。
//
// この構造のおかげで、末尾追加は整数部のインクリメント（+1）だけ、先頭追加は
// デクリメント（-1）だけで済み、キー長は件数 n に対して O(log n) にとどまる。
//
// # 実測したキー長（fracindex_test.go の性質テストが上限を固定している）
//
//	末尾追加 10,000 件      最大 4 バイト / 合計 36,032 バイト
//	先頭追加 10,000 件      最大 4 バイト / 合計 36,030 バイト
//	ランダム位置 10,000 件  最大 7 バイト / 合計 46,212 バイト（seed 固定）
//
// 末尾 / 先頭追加が数バイトで収まるのが整数部を持つ最大の効能で、整数部を持たない
// 「常に中点を取る」方式だと同じ 10,000 件で 1 キー 2,000 バイト・合計 10 MB まで膨らむ。
// 代わりにランダム挿入は最短キーでも整数部 2 バイトを必ず背負うぶんわずかに増える
// （最大 6→7 バイト・合計 32 KB→46 KB）。実際のエディタ操作で圧倒的に多いのは末尾追加なので、
// この交換は割に合う。
//
// 同じ隙間へ挿し続けるケース（毎回 prev と直前に採ったキーの間に入れる = 隙間が毎回半分になる）
// だけは小数部が伸びる。これは方式上避けられない（有限の文字集合で「常に間」を作る以上、
// 情報量ぶんは必ず消費する）。実測では base62 の 1 文字がちょうど 6 回ぶんの挿入を吸収し、
// i 回目のキー長は 3 + (i-1)/6 バイト（整数除算）になる:
//
//	同じ隙間へ 100 回      最大 19 バイト
//	同じ隙間へ 1,000 回    最大 169 バイト
//	同じ隙間へ 10,000 回   最大 1,669 バイト
//
// ただし「同じ 2 件の間へ 1 万回挿し込む」のは人間の編集操作としては現実的でない。
// position は「同じ親の中だけで意味を持つローカルな順序」なので、万一長くなりすぎても
// 親単位で採番し直せば（UPDATE 1 回）縮められる。
package fracindex

import (
	"errors"
	"fmt"
	"strings"
)

// digits は分数インデックスの桁に使う文字集合（base62）。
//
// base62 を選んだ理由:
//   - この並び（0-9 → A-Z → a-z）は ASCII コード順と完全に一致する。つまり Go の文字列比較
//     （バイト比較）と PostgreSQL の C コレーションでの ORDER BY が同じ順序になる。
//     アプリと DB で並びがずれない、というのがこの実装の一番の前提条件になる
//     （DB 側は position 列を COLLATE "C" に固定して担保する。ロケール依存のコレーションだと
//     'a' < 'B' のような順序になり、バイト比較と食い違うため）。
//   - 1 文字あたり log2(62) ≒ 5.95 bit 分の細分ができるので、同じ隙間への連続挿入でも
//     小数部が伸びにくい（base16 等より短く保てる）。
//   - URL・JSON にそのまま置ける安全な文字だけで構成されている。
const digits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// smallestDigit / largestDigit は文字集合の両端。digits と一致していることはテストで固定する
// （定数式では文字列の添字を取れないため直書き）。
const (
	smallestDigit byte = '0'
	largestDigit  byte = 'z'
)

// 整数部のヘッダに使う 4 隅。'a' が非負側の最短（整数 0..61）、'Z' が負側の最短（整数 -1..-62）。
const (
	firstPositiveHead byte = 'a'
	lastPositiveHead  byte = 'z'
	firstNegativeHead byte = 'A'
	lastNegativeHead  byte = 'Z'
)

// integerDigitCount はヘッダ 1 文字が表せる最大の桁数（整数部の全長 - ヘッダ 1 文字）。
const integerDigitCount = 26

var (
	// zeroKey は最初の 1 件目に使うキー（整数 0）。前にも後ろにも同じだけ伸ばせる位置。
	zeroKey = string(firstPositiveHead) + string(smallestDigit)
	// smallestInteger は表現できる最小の整数部。これ単独をキーにすると「その手前」を
	// 作れなくなる（整数も下げられず小数部も無い）ので、キーとしては不正扱いにする。
	smallestInteger = string(firstNegativeHead) + strings.Repeat(string(smallestDigit), integerDigitCount)
)

var (
	// ErrInvalidKey は Between に渡せないキー。文字集合外の文字、整数部の桁数ヘッダと実際の
	// 長さの不一致、正規形でないキー（小数部の末尾が最小文字）、および「これ以上外側を作れない」
	// 表現域の端のキーがこれに当たる。
	ErrInvalidKey = errors.New("fracindex: 不正なキー")
	// ErrOutOfOrder は prev >= next のときに返す。呼び出し側の並び順のバグを黙って通さない。
	ErrOutOfOrder = errors.New("fracindex: prev は next より小さい必要があります")
)

// Between は prev と next の「辞書順で厳密に間」にあるキーを返す。
//
// prev が空文字ならリストの先頭より前、next が空文字なら末尾より後を意味する
// （両方空なら最初の 1 件目のキーになる）。prev >= next のときは [ErrOutOfOrder] を返す。
//
// 返すキーは常に正規形（小数部の末尾が最小文字 '0' にならない）。'0' で終わる小数部を許すと
// 「その手前に挿す」余地が消えるため（小数部 "V0" は "V" < x < "V0" を満たす x が無い）。
// 整数部はヘッダで長さが決まる固定長なので末尾が '0' でもよい（"a0" は正規形。その手前へは
// 整数部のデクリメント "Zz" で挿せる）。
func Between(prev, next string) (string, error) {
	if err := validateKey(prev, "prev"); err != nil {
		return "", err
	}
	if err := validateKey(next, "next"); err != nil {
		return "", err
	}
	if prev != "" && next != "" && prev >= next {
		return "", fmt.Errorf("%w: prev=%q next=%q", ErrOutOfOrder, prev, next)
	}

	switch {
	case prev == "" && next == "":
		return zeroKey, nil
	case prev == "":
		return beforeKey(next)
	case next == "":
		return afterKey(prev)
	}

	intPrev, fracPrev := splitKey(prev)
	intNext, fracNext := splitKey(next)
	if intPrev == intNext {
		// 同じ整数の内側。ここだけ小数部を掘る。
		mid, err := midpoint(fracPrev, fracNext, true)
		if err != nil {
			return "", err
		}
		return intPrev + mid, nil
	}
	// 整数部が違う ＝ 間に整数の目盛りが入る余地がある。まず prev の整数部 +1 を試す。
	stepped, ok := incrementInteger(intPrev)
	if !ok {
		// intPrev が最大整数なら intNext > intPrev はあり得ないので到達しない。
		return "", fmt.Errorf("%w: prev=%q はこれ以上大きい整数部を作れません", ErrInvalidKey, prev)
	}
	if stepped < next {
		return stepped, nil
	}
	// +1 が next 以上になる（next が「prev の整数部 +1」ちょうど）ときは prev の小数部を上に掘る。
	mid, err := midpoint(fracPrev, "", false)
	if err != nil {
		return "", err
	}
	return intPrev + mid, nil
}

// beforeKey は next より前のキーを作る（リストの先頭に挿す）。整数部のデクリメントで済ませる。
func beforeKey(next string) (string, error) {
	intNext, fracNext := splitKey(next)
	if intNext == smallestInteger {
		// 整数をこれ以上下げられないので、最小整数の小数部を掘って前に潜り込む。
		// next 自身は「最小整数 + 空の小数部」ではない（validateKey が弾く）ので fracNext は非空。
		mid, err := midpoint("", fracNext, true)
		if err != nil {
			return "", err
		}
		return intNext + mid, nil
	}
	if fracNext != "" {
		// next に小数部がある ＝ 整数部だけのキーが必ず next より小さい。掘らずに済む。
		return intNext, nil
	}
	lowered, ok := decrementInteger(intNext)
	if !ok {
		// intNext == smallestInteger のときだけ false になるが、それは上で処理済み。
		return "", fmt.Errorf("%w: next=%q より手前のキーは作れません", ErrInvalidKey, next)
	}
	if lowered == smallestInteger {
		// 最小整数そのものはキーにできない（その手前を作れないので validateKey が弾く）。
		// 小数部を足して next のすぐ下に置けば、以後も小数部を掘って前へ挿し続けられる。
		mid, err := midpoint("", "", false)
		if err != nil {
			return "", err
		}
		return lowered + mid, nil
	}
	return lowered, nil
}

// afterKey は prev より後ろのキーを作る（リストの末尾に足す）。整数部のインクリメントで済ませる。
func afterKey(prev string) (string, error) {
	intPrev, fracPrev := splitKey(prev)
	if raised, ok := incrementInteger(intPrev); ok {
		return raised, nil
	}
	// 整数部が上限に張り付いた場合だけ、小数部を上へ伸ばす。
	mid, err := midpoint(fracPrev, "", false)
	if err != nil {
		return "", err
	}
	return intPrev + mid, nil
}

// integerLen はヘッダ 1 文字から整数部の全長（ヘッダを含む）を返す。
//
// 小文字は非負側で 'a'→2 … 'z'→27、大文字は負側で 'Z'→2 … 'A'→27。負側でヘッダが小さいほど
// 長くなるのは、「より小さい整数ほど文字列としても前に来る」を桁数が変わっても保つため。
func integerLen(head byte) (int, bool) {
	switch {
	case head >= firstPositiveHead && head <= lastPositiveHead:
		return int(head-firstPositiveHead) + 2, true
	case head >= firstNegativeHead && head <= lastNegativeHead:
		return int(lastNegativeHead-head) + 2, true
	default:
		return 0, false
	}
}

// splitKey は検証済みのキーを整数部と小数部に分ける。
func splitKey(key string) (integer, fraction string) {
	n, _ := integerLen(key[0])
	return key[:n], key[n:]
}

// incrementInteger は整数部に 1 を足す。上限（"z"+"z"×26）を超えるときは false を返す。
func incrementInteger(integer string) (string, bool) {
	head := integer[0]
	digs := []byte(integer[1:])
	carry := true
	for i := len(digs) - 1; carry && i >= 0; i-- {
		d := strings.IndexByte(digits, digs[i]) + 1
		if d == len(digits) {
			digs[i] = smallestDigit
			continue
		}
		digs[i] = digits[d]
		carry = false
	}
	if !carry {
		return string(head) + string(digs), true
	}
	// 桁が溢れたのでヘッダを 1 つ進める。負側 → 非負側の乗り換えだけ桁数の増減が反転する。
	switch head {
	case lastNegativeHead: // 整数 -1 の次は 0
		return zeroKey, true
	case lastPositiveHead:
		return "", false
	}
	next := head + 1
	if next > firstPositiveHead {
		digs = append(digs, smallestDigit) // 非負側はヘッダが進むほど桁が増える
	} else {
		digs = digs[:len(digs)-1] // 負側はヘッダが進むほど桁が減る
	}
	return string(next) + string(digs), true
}

// decrementInteger は整数部から 1 を引く。下限（smallestInteger）を下回るときは false を返す。
func decrementInteger(integer string) (string, bool) {
	head := integer[0]
	digs := []byte(integer[1:])
	borrow := true
	for i := len(digs) - 1; borrow && i >= 0; i-- {
		d := strings.IndexByte(digits, digs[i]) - 1
		if d < 0 {
			digs[i] = largestDigit
			continue
		}
		digs[i] = digits[d]
		borrow = false
	}
	if !borrow {
		return string(head) + string(digs), true
	}
	switch head {
	case firstPositiveHead: // 整数 0 の 1 つ前は -1
		return string(lastNegativeHead) + string(largestDigit), true
	case firstNegativeHead:
		return "", false
	}
	prev := head - 1
	if prev < lastNegativeHead {
		digs = append(digs, largestDigit) // 負側はヘッダが戻るほど桁が増える
	} else {
		digs = digs[:len(digs)-1] // 非負側はヘッダが戻るほど桁が減る
	}
	return string(prev) + string(digs), true
}

// midpoint は小数部 a と b の「厳密に間」を返す。hasUpper=false は b が上限なしの意味。
// a は空文字（下限なし）を許す。前提（a < b・両者が正規形）は呼び出し側が保証する。
// 返り値は末尾が最小文字にならない（正規形）。
func midpoint(a, b string, hasUpper bool) (string, error) {
	if hasUpper {
		if a >= b {
			return "", fmt.Errorf("%w: 小数部 a=%q b=%q の順序が矛盾しています", ErrOutOfOrder, a, b)
		}
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
			return "", fmt.Errorf("%w: 小数部 a=%q b=%q の順序が矛盾しています", ErrOutOfOrder, a, b)
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
		// 桁が 2 以上離れていれば、その中間の 1 文字だけで足りる（小数部を伸ばさない）。
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

// validateKey は文字集合・整数部の桁数ヘッダ・正規形を検査する。空文字は端の意味なので許す。
func validateKey(key, name string) error {
	if key == "" {
		return nil
	}
	for i := 0; i < len(key); i++ {
		if strings.IndexByte(digits, key[i]) < 0 {
			return fmt.Errorf("%w: %s=%q に文字集合外の文字が含まれます", ErrInvalidKey, name, key)
		}
	}
	if key == smallestInteger {
		return fmt.Errorf("%w: %s=%q は表現域の下端なのでキーに使えません", ErrInvalidKey, name, key)
	}
	n, ok := integerLen(key[0])
	if !ok {
		return fmt.Errorf("%w: %s=%q の先頭 %q は整数部の桁数ヘッダになりません", ErrInvalidKey, name, key, string(key[0]))
	}
	if len(key) < n {
		return fmt.Errorf("%w: %s=%q は桁数ヘッダ %q が要求する整数部 %d 文字に足りません", ErrInvalidKey, name, key, string(key[0]), n)
	}
	if fraction := key[n:]; fraction != "" && fraction[len(fraction)-1] == smallestDigit {
		return fmt.Errorf("%w: %s=%q は小数部の末尾が %q（正規形ではありません）", ErrInvalidKey, name, key, string(smallestDigit))
	}
	return nil
}
