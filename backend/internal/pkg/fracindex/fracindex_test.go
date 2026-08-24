package fracindex

import (
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test_smallestDigit_matchesDigits は「最小文字の直書き」が文字集合の先頭とずれていないことを固定する。
func Test_smallestDigit_matchesDigits(t *testing.T) {
	require.Equal(t, digits[0], smallestDigit)
	// 文字集合自体が昇順であること（バイト比較と一致する前提が崩れると全体が壊れる）。
	require.True(t, sort.SliceIsSorted([]byte(digits), func(i, j int) bool { return digits[i] < digits[j] }))
}

func Test_Between_境界(t *testing.T) {
	cases := []struct {
		name string
		prev string
		next string
	}{
		{"両端が空（最初の 1 件）", "", ""},
		{"先頭に挿す（prev だけ空）", "", "V"},
		{"末尾に足す（next だけ空）", "V", ""},
		{"隣接キーの間（a と b）", "a", "b"},
		{"1 文字ずつ離れた桁の間", "1", "2"},
		{"最小キーの手前", "", "1"},
		{"最大文字の後ろ", "z", ""},
		{"同じ prefix の深い入れ子", "V0V", "V0W"},
		{"prefix が一致し next が長い", "V", "V1"},
		{"prefix が一致し prev が長い", "V0V", "V1"},
		{"深い位置での隣接", "zzzzy", "zzzzz"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Between(tc.prev, tc.next)
			require.NoError(t, err)
			requireOrdered(t, tc.prev, got, tc.next)
			requireNormalized(t, got)
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
		{"prev == next", "V", "V", ErrOutOfOrder},
		{"prev > next", "b", "a", ErrOutOfOrder},
		{"prev が next の prefix で大きい", "V1", "V", ErrOutOfOrder},
		{"文字集合外（prev）", "V-", "z", ErrInvalidKey},
		{"文字集合外（next）", "1", "V!", ErrInvalidKey},
		{"正規形でない prev（末尾が 0）", "V0", "z", ErrInvalidKey},
		{"正規形でない next（末尾が 0）", "1", "V0", ErrInvalidKey},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Between(tc.prev, tc.next)
			require.ErrorIs(t, err, tc.wantErr)
			require.Empty(t, got)
		})
	}
}

// Test_Between_ランダム挿入で順序が保たれる は本パッケージの肝。
// ランダムな位置へ 1000 回挿し込んでも「キーの辞書順 == 配列の並び」が常に一致することを確かめる。
func Test_Between_ランダム挿入で順序が保たれる(t *testing.T) {
	rng := rand.New(rand.NewSource(20260824)) // 失敗を再現できるよう seed を固定する
	var keys []string

	for i := 0; i < 1000; i++ {
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

		keys = append(keys, "")
		copy(keys[at+1:], keys[at:])
		keys[at] = key

		// 配列の並び（= 論理的な並び）とキーの辞書順が一致し続けること。
		require.Truef(t, sort.StringsAreSorted(keys), "i=%d でキーの辞書順が配列の並びとずれました: %v", i, keys)
	}

	// 同じキーが 2 度出ないこと（position は UNIQUE 制約の対象にもなる）。
	uniq := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		_, dup := uniq[k]
		require.Falsef(t, dup, "キー %q が重複しました", k)
		uniq[k] = struct{}{}
	}
}

// Test_Between_同じ位置への連続挿入でキー長が爆発しない は、最悪ケース（常に同じ隙間へ挿す）でも
// キー長が挿入回数に比例して伸びないことを確かめる。base62 は 1 文字あたり約 5.95 bit 細分できるため、
// 100 回挿しても 20 文字台に収まる（整数連番の振り直しを避けるための性能上の前提）。
func Test_Between_同じ位置への連続挿入でキー長が爆発しない(t *testing.T) {
	const iterations = 100

	t.Run("固定の prev と直前に採ったキーの間へ挿し続ける", func(t *testing.T) {
		prev, err := Between("", "")
		require.NoError(t, err)
		next, err := Between(prev, "")
		require.NoError(t, err)

		maxLen := 0
		for i := 0; i < iterations; i++ {
			key, err := Between(prev, next)
			require.NoErrorf(t, err, "i=%d prev=%q next=%q", i, prev, next)
			requireOrdered(t, prev, key, next)
			requireNormalized(t, key)
			next = key // 隙間は毎回半分になる（最悪ケース）
			maxLen = max(maxLen, len(key))
		}
		require.LessOrEqualf(t, maxLen, 32, "キー長が想定以上に伸びました（maxLen=%d）", maxLen)
	})

	t.Run("常に先頭へ挿し続ける", func(t *testing.T) {
		next, err := Between("", "")
		require.NoError(t, err)
		maxLen := 0
		for i := 0; i < iterations; i++ {
			key, err := Between("", next)
			require.NoErrorf(t, err, "i=%d next=%q", i, next)
			requireOrdered(t, "", key, next)
			requireNormalized(t, key)
			next = key
			maxLen = max(maxLen, len(key))
		}
		require.LessOrEqualf(t, maxLen, 32, "キー長が想定以上に伸びました（maxLen=%d）", maxLen)
	})

	t.Run("常に末尾へ足し続ける", func(t *testing.T) {
		prev := ""
		maxLen := 0
		for i := 0; i < iterations; i++ {
			key, err := Between(prev, "")
			require.NoErrorf(t, err, "i=%d prev=%q", i, prev)
			requireOrdered(t, prev, key, "")
			requireNormalized(t, key)
			prev = key
			maxLen = max(maxLen, len(key))
		}
		require.LessOrEqualf(t, maxLen, 32, "キー長が想定以上に伸びました（maxLen=%d）", maxLen)
	})
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

// requireNormalized は生成キーが正規形（文字集合内・末尾が最小文字でない）であることを検証する。
func requireNormalized(t *testing.T, key string) {
	t.Helper()
	require.NotEmpty(t, key)
	for i := 0; i < len(key); i++ {
		require.Truef(t, strings.IndexByte(digits, key[i]) >= 0, "key=%q に文字集合外の文字があります", key)
	}
	require.NotEqualf(t, smallestDigit, key[len(key)-1], "key=%q の末尾が最小文字です（後続の Between が枯れます）", key)
	// 生成キーをそのまま次の Between に渡せること（正規形の実質的な意味）。
	require.NoError(t, validateKey(key, "key"))
}
