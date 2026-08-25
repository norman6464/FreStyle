package fracindex

import (
	"errors"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// maxInteger は表現できる最大の整数部（smallestInteger の対）。上限に張り付いた側の挙動を試すのに使う。
var maxInteger = string(lastPositiveHead) + strings.Repeat(string(largestDigit), integerDigitCount)

// Test_文字集合の前提 は「文字集合の並び == バイト順」という実装全体の前提を固定する。
// ここが崩れると Go の比較と PostgreSQL の ORDER BY（COLLATE "C"）が食い違い、全部が壊れる。
func Test_文字集合の前提(t *testing.T) {
	require.Equal(t, digits[0], smallestDigit)
	require.Equal(t, digits[len(digits)-1], largestDigit)
	require.True(t, sort.SliceIsSorted([]byte(digits), func(i, j int) bool { return digits[i] < digits[j] }))
	require.Len(t, digits, 62)

	// ヘッダは全て文字集合の中にあり、大文字（負）が小文字（非負）より前に並ぶ。
	for _, head := range []byte{firstNegativeHead, lastNegativeHead, firstPositiveHead, lastPositiveHead} {
		require.GreaterOrEqual(t, strings.IndexByte(digits, head), 0)
	}
	require.Less(t, firstNegativeHead, lastNegativeHead)
	require.Less(t, lastNegativeHead, firstPositiveHead)
	require.Less(t, firstPositiveHead, lastPositiveHead)

	// ヘッダが表す整数部の長さ。負側はヘッダが小さいほど長い（＝より小さい整数を表す）。
	for _, tc := range []struct {
		head byte
		want int
	}{
		{'a', 2},
		{'b', 3},
		{'z', 27},
		{'Z', 2},
		{'Y', 3},
		{'A', 27},
	} {
		got, ok := integerLen(tc.head)
		require.Truef(t, ok, "head=%q", string(tc.head))
		require.Equalf(t, tc.want, got, "head=%q", string(tc.head))
	}
	for _, head := range []byte{'0', '5', '9'} {
		_, ok := integerLen(head)
		require.Falsef(t, ok, "数字 %q は桁数ヘッダになってはいけません", string(head))
	}

	require.Len(t, smallestInteger, 27)
	require.Len(t, maxInteger, 27)
	require.Equal(t, "a0", zeroKey)
}

func Test_Between_境界(t *testing.T) {
	cases := []struct {
		name string
		prev string
		next string
		want string // 空なら値は問わず順序と正規形だけ見る
	}{
		{"両端が空（最初の 1 件）", "", "", "a0"},
		{"末尾に足す（整数部 +1 だけ）", "a0", "", "a1"},
		{"先頭に挿す（整数部 -1 だけ）", "", "a0", "Zz"},
		{"整数が隣り合うので小数部を掘る", "a0", "a1", "a0V"},
		{"整数が 2 つ以上離れていれば整数で刻む", "a0", "a5", "a1"},
		{"同じ整数の小数部どうし", "a0", "a0V", "a0G"},
		{"隣接した小数部（1 文字伸びる）", "a0V", "a0W", "a0VV"},
		{"小数部つきキーの手前は整数部だけで足りる", "", "a0V", "a0"},

		// 整数部の桁上がり（末尾追加側）。'z' で桁が溢れるとヘッダが 1 つ進んで桁数が増える。
		{"非負側の桁上がり az → b00", "az", "", "b00"},
		{"非負側の 2 段目の桁上がり bzz → c000", "bzz", "", "c000"},
		{"負から非負への乗り換え Zz → a0", "Zz", "", "a0"},
		{"負側の桁内インクリメント Z0 → Z1", "Z0", "", "Z1"},

		// 整数部の桁下がり（先頭追加側）。'0' で借りるとヘッダが 1 つ戻って桁数が増える。
		{"負側の桁下がり Z0 → Yzz", "", "Z0", "Yzz"},
		{"負側の 2 段目の桁下がり Y00 → Xzzz", "", "Y00", "Xzzz"},
		{"非負側の桁下がり b00 → az", "", "b00", "az"},
		{"非負側の桁内デクリメント a1 → a0", "", "a1", "a0"},

		// 表現域の端。整数をこれ以上動かせない側は小数部に逃げる。
		{"最大整数の後ろ（もう +1 できないので小数部へ）", maxInteger, "", maxInteger + "V"},
		{"最小整数の小数部の手前へ潜り込む", "", smallestInteger + "V", smallestInteger + "G"},
		{"最小整数の小数部どうし", smallestInteger + "V", smallestInteger + "W", smallestInteger + "VV"},
		// -1 した結果が最小整数そのものになるケース。裸の最小整数はキーにできないので小数部を足す
		// （FuzzBetween が見つけた退行。素直に返すと次の Between が ErrInvalidKey で枯れる）。
		{"最小整数 +1 の手前", "", string(firstNegativeHead) + strings.Repeat("0", 25) + "1", smallestInteger + "V"},
		{"最小整数の直後は整数部を +1 できる", smallestInteger + "V", "", string(firstNegativeHead) + strings.Repeat("0", 25) + "1"},

		// 整数部が違うキーどうし。
		{"整数部の桁数が違う（負 → 非負）", "Zz", "a1", "a0"},
		{"整数部の桁数が違う（2 文字 → 3 文字）", "az", "b01", "b00"},
		{"隣り合う整数のあいだ（小数部つき prev）", "a0V", "a1", "a0l"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Between(tc.prev, tc.next)
			require.NoError(t, err)
			requireOrdered(t, tc.prev, got, tc.next)
			requireNormalized(t, got)
			if tc.want != "" {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func Test_Between_不正入力(t *testing.T) {
	cases := []struct {
		name    string
		prev    string
		next    string
		wantErr error
	}{
		{"prev == next", "a0", "a0", ErrOutOfOrder},
		{"prev > next", "b00", "a0", ErrOutOfOrder},
		{"prev が next の prefix で大きい", "a0V", "a0", ErrOutOfOrder},
		{"文字集合外（prev）", "a-", "b00", ErrInvalidKey},
		{"文字集合外（next）", "a0", "a0!", ErrInvalidKey},
		{"先頭が数字で桁数ヘッダにならない（prev）", "0V", "a0", ErrInvalidKey},
		{"先頭が数字で桁数ヘッダにならない（next）", "", "1a", ErrInvalidKey},
		{"桁数ヘッダと長さの不一致（短すぎる・大文字）", "V0", "z0", ErrInvalidKey},
		{"桁数ヘッダと長さの不一致（短すぎる・小文字）", "b0", "c000", ErrInvalidKey},
		{"桁数ヘッダと長さの不一致（ヘッダだけ）", "a", "b00", ErrInvalidKey},
		{"小数部の末尾が最小文字（prev）", "a0V0", "a1", ErrInvalidKey},
		{"小数部の末尾が最小文字（next）", "a0", "a0V0", ErrInvalidKey},
		{"表現域の下端そのもの（prev）", smallestInteger, "a0", ErrInvalidKey},
		{"表現域の下端そのもの（next）", "", smallestInteger, ErrInvalidKey},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Between(tc.prev, tc.next)
			require.ErrorIs(t, err, tc.wantErr)
			require.Empty(t, got)
		})
	}
}

// Test_整数部のインクリメントとデクリメント は、整数部だけを取り出して
// 「+1 は必ず辞書順で大きくなる」「-1 で元に戻る」を桁上がり・ヘッダ跨ぎを含めて確かめる。
// キー長がヘッダ通りであることも毎回見るので、桁数の増減規則が壊れると落ちる。
func Test_整数部のインクリメントとデクリメント(t *testing.T) {
	// 負側の 3 桁から非負側の 3 桁まで通しで歩く（ヘッダ跨ぎを 4 回踏む）。
	current := "Y00"
	seen := []string{current}
	for i := 0; i < 12000; i++ {
		next, ok := incrementInteger(current)
		require.Truef(t, ok, "i=%d current=%q でインクリメントできませんでした", i, current)
		require.Truef(t, current < next, "i=%d: %q < %q が成り立ちません", i, current, next)
		wantLen, headOK := integerLen(next[0])
		require.Truef(t, headOK, "i=%d: %q のヘッダが不正です", i, next)
		require.Lenf(t, next, wantLen, "i=%d: %q の長さがヘッダの示す桁数と違います", i, next)

		back, ok := decrementInteger(next)
		require.Truef(t, ok, "i=%d next=%q でデクリメントできませんでした", i, next)
		require.Equalf(t, current, back, "i=%d: %q を +1 → -1 して戻りませんでした", i, current)

		current = next
		seen = append(seen, current)
	}
	require.True(t, sort.StringsAreSorted(seen), "整数部の並びが辞書順と一致しません")

	// 表現域の端。これ以上動かせないことを返り値で示す（黙って壊れた値を返さない）。
	_, ok := incrementInteger(maxInteger)
	require.False(t, ok, "最大整数はインクリメントできてはいけません")
	_, ok = decrementInteger(smallestInteger)
	require.False(t, ok, "最小整数はデクリメントできてはいけません")
}

// Test_Between_ランダム挿入で順序が保たれる は本パッケージの肝。
// ランダムな位置へ 10,000 回挿し込んでも「キーのバイト順 == 配列の並び」が常に一致すること。
func Test_Between_ランダム挿入で順序が保たれる(t *testing.T) {
	const iterations = 10000
	rng := rand.New(rand.NewSource(20260825)) // 失敗を再現できるよう seed を固定する
	keys := make([]string, 0, iterations)
	maxLen := 0

	for i := 0; i < iterations; i++ {
		at := rng.Intn(len(keys) + 1) // 0 = 先頭, len(keys) = 末尾
		prev, next := "", ""
		if at > 0 {
			prev = keys[at-1]
		}
		if at < len(keys) {
			next = keys[at]
		}

		key, err := Between(prev, next)
		require.NoErrorf(t, err, "i=%d prev=%q next=%q", i, prev, next)
		requireOrdered(t, prev, key, next)
		requireNormalized(t, key)
		maxLen = max(maxLen, len(key))

		keys = append(keys, "")
		copy(keys[at+1:], keys[at:])
		keys[at] = key
	}

	// 配列の並び（= 論理的な並び）とキーのバイト順が一致すること。
	// 挿入位置の左右関係は毎回 requireOrdered が見ているので、全体走査はここで 1 回だけ行う
	// （ループ内で回すと比較回数が O(n^2) になり、-race 付きの CI で無視できない時間を食う）。
	require.True(t, sort.StringsAreSorted(keys), "キーのバイト順が配列の並びとずれました")

	// 同じキーが 2 度出ないこと（position は UNIQUE 索引の対象にもなる）。
	uniq := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		_, dup := uniq[k]
		require.Falsef(t, dup, "キー %q が重複しました", k)
		uniq[k] = struct{}{}
	}

	// ランダム挿入は整数部（最短 2 バイト）を必ず背負うぶん、素の中点方式よりわずかに長い。
	// seed 固定での実測は 7 バイト。ここが跳ねたら小数部の掘り方が壊れている。
	require.LessOrEqualf(t, maxLen, 7, "ランダム挿入のキー長が想定を超えました（maxLen=%d）", maxLen)
}

// Test_Between_末尾追加でキー長が伸びない は今回の作り直しの本命。
// 末尾追加・先頭追加は整数部の ±1 だけで済むので、キー長は件数に対して O(log_62 n) にとどまる。
// 10,000 件で 4 バイト（62 件までが 2 バイト、3,906 件までが 3 バイト、以降 4 バイト）。
func Test_Between_末尾追加でキー長が伸びない(t *testing.T) {
	const iterations = 10000

	t.Run("常に末尾へ足し続ける", func(t *testing.T) {
		prev := ""
		maxLen, total := 0, 0
		for i := 0; i < iterations; i++ {
			key, err := Between(prev, "")
			require.NoErrorf(t, err, "i=%d prev=%q", i, prev)
			requireOrdered(t, prev, key, "")
			requireNormalized(t, key)
			prev = key
			maxLen = max(maxLen, len(key))
			total += len(key)

			// 「n 件目までのキー長は log_62 で決まる桁数を超えない」を毎回見る。
			require.LessOrEqualf(t, len(key), expectedIntegerLen(i+1), "i=%d key=%q が O(log n) を超えました", i, key)
		}
		require.Equal(t, 4, maxLen, "末尾追加 10,000 件の最大キー長")
		require.Equal(t, 36032, total, "末尾追加 10,000 件の合計バイト数")
	})

	t.Run("常に先頭へ挿し続ける", func(t *testing.T) {
		next := ""
		maxLen, total := 0, 0
		for i := 0; i < iterations; i++ {
			key, err := Between("", next)
			require.NoErrorf(t, err, "i=%d next=%q", i, next)
			requireOrdered(t, "", key, next)
			requireNormalized(t, key)
			next = key
			maxLen = max(maxLen, len(key))
			total += len(key)

			require.LessOrEqualf(t, len(key), expectedIntegerLen(i+1), "i=%d key=%q が O(log n) を超えました", i, key)
		}
		require.Equal(t, 4, maxLen, "先頭追加 10,000 件の最大キー長")
		require.Equal(t, 36030, total, "先頭追加 10,000 件の合計バイト数")
	})
}

// Test_Between_同じ隙間へ挿し続けると小数部が伸びる は、方式上避けられない伸びを「どれだけ伸びるか」
// まで固定する。有限の文字集合で常に「間」を作る以上、情報量ぶんは必ず消費するので、
// ここは 0 にはできない。base62 の 1 文字がちょうど 6 回ぶんの挿入を吸収する（i 回目 = 3 + (i-1)/6 バイト）。
func Test_Between_同じ隙間へ挿し続けると小数部が伸びる(t *testing.T) {
	const iterations = 10000

	prev, err := Between("", "")
	require.NoError(t, err)
	next, err := Between(prev, "")
	require.NoError(t, err)

	maxLen := 0
	for i := 1; i <= iterations; i++ {
		key, err := Between(prev, next)
		require.NoErrorf(t, err, "i=%d prev=%q next=%q", i, prev, next)
		requireOrdered(t, prev, key, next)
		requireNormalized(t, key)
		next = key // 隙間は毎回半分になる（最悪ケース）
		maxLen = max(maxLen, len(key))

		// 実測した増加率をそのまま固定する。速くなっても遅くなっても落ちる。
		require.Equalf(t, 3+(i-1)/6, len(key), "i=%d key の長さが実測した増加率と違います", i)
	}
	require.Equal(t, 1669, maxLen, "同じ隙間へ 10,000 回挿したときの最大キー長")
}

// expectedIntegerLen は「末尾 / 先頭に n 件足したときのキー長の上限」を log_62 から求める。
// 整数部はヘッダ 1 文字 + 桁で、桁数 d が表せるのは 62^d 件（片側）。
func expectedIntegerLen(n int) int {
	capacity, digitCount := 62, 1
	for capacity < n {
		capacity *= 62
		digitCount++
	}
	return 1 + digitCount
}

// FuzzBetween は任意の 2 入力に対して「エラーを返す」か「prev < 結果 < next を満たす正規形のキーを返す」
// のどちらかであることを確かめる。無限再帰・パニックで落ちないことも同時に見る。
func FuzzBetween(f *testing.F) {
	seeds := []struct{ prev, next string }{
		{"", ""},
		{"a0", ""},
		{"", "a0"},
		{"a0", "a1"},
		{"a0", "a0V"},
		{"az", "b00"},
		{"Zz", "a0"},
		{"Y00", "Yzz"},
		{maxInteger, ""},
		{"", smallestInteger + "V"},
		{smallestInteger, "a0"}, // 不正入力（表現域の下端）
		{"V0", "z0"},            // 不正入力（桁数ヘッダと長さの不一致）
		{"a0V0", "a1"},          // 不正入力（小数部の末尾が最小文字）
		{"a-", "b00"},           // 不正入力（文字集合外）
		{"b00", "a0"},           // 不正入力（prev > next）
	}
	for _, s := range seeds {
		f.Add(s.prev, s.next)
	}

	f.Fuzz(func(t *testing.T, prev, next string) {
		// 極端に長い入力は探索の役に立たないうえ小数部の再帰が深くなるだけなので切る。
		if len(prev) > 512 || len(next) > 512 {
			t.Skip()
		}

		got, err := Between(prev, next)
		if err != nil {
			require.Empty(t, got, "エラー時はキーを返してはいけません")
			require.Truef(t, errorIsKnown(err), "想定外のエラー種別です: %v", err)
			return
		}

		require.NotEmpty(t, got)
		require.NoErrorf(t, validateKey(got, "got"), "生成キー %q が正規形ではありません", got)
		if prev != "" {
			require.Truef(t, prev < got, "prev=%q < got=%q が成り立ちません", prev, got)
		}
		if next != "" {
			require.Truef(t, got < next, "got=%q < next=%q が成り立ちません", got, next)
		}

		// 生成したキーをそのまま次の境界に使えること（採番を繰り返しても枯れない）。
		if prev != "" {
			_, err = Between(prev, got)
			require.NoErrorf(t, err, "prev=%q と got=%q の間が作れません", prev, got)
		}
		if next != "" {
			_, err = Between(got, next)
			require.NoErrorf(t, err, "got=%q と next=%q の間が作れません", got, next)
		}
	})
}

// errorIsKnown は公開しているエラー値のどちらかであることを確かめる（新種の裸エラーを漏らさない）。
func errorIsKnown(err error) bool {
	return errors.Is(err, ErrInvalidKey) || errors.Is(err, ErrOutOfOrder)
}

// requireOrdered は prev < key < next（空文字は「端」の意味）を検証する。
func requireOrdered(t *testing.T, prev, key, next string) {
	t.Helper()
	require.NotEmpty(t, key)
	if prev != "" {
		require.Truef(t, prev < key, "prev=%q < key=%q が成り立ちません", prev, key)
	}
	if next != "" {
		require.Truef(t, key < next, "key=%q < next=%q が成り立ちません", key, next)
	}
}

// requireNormalized は生成キーが正規形（文字集合内・桁数ヘッダと整合・小数部の末尾が最小文字でない）
// であることを検証する。
func requireNormalized(t *testing.T, key string) {
	t.Helper()
	require.NotEmpty(t, key)
	for i := 0; i < len(key); i++ {
		require.Truef(t, strings.IndexByte(digits, key[i]) >= 0, "key=%q に文字集合外の文字があります", key)
	}
	wantLen, ok := integerLen(key[0])
	require.Truef(t, ok, "key=%q の先頭が桁数ヘッダになっていません", key)
	require.GreaterOrEqualf(t, len(key), wantLen, "key=%q が桁数ヘッダの示す整数部に足りません", key)
	if fraction := key[wantLen:]; fraction != "" {
		require.NotEqualf(t, smallestDigit, fraction[len(fraction)-1],
			"key=%q の小数部の末尾が最小文字です（後続の Between が枯れます）", key)
	}
	// 生成キーをそのまま次の Between に渡せること（正規形の実質的な意味）。
	require.NoError(t, validateKey(key, "key"))
}
