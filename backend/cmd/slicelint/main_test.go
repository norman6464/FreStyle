package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writePersistence は一時ディレクトリに internal/adapter/persistence/<name> を作る。
func writePersistence(t *testing.T, name, src string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "internal", "adapter", "persistence")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644))
	return root
}

func runOn(t *testing.T, root string) (code int, stdout string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = runCLI([]string{root}, &out, &errOut)
	return code, out.String()
}

func Test_一覧メソッドがnilスライスのままなら違反(t *testing.T) {
	root := writePersistence(t, "repo.go", `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	var rows []Item
	return rows, nil
}
`)
	code, out := runOn(t, root)

	assert.Equal(t, 1, code)
	assert.Contains(t, out, "ListAll")
	assert.Contains(t, out, "make([]T, 0)")
}

func Test_空スライスで初期化していれば通る(t *testing.T) {
	root := writePersistence(t, "repo.go", `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	rows := make([]Item, 0)
	return rows, nil
}
`)
	code, out := runOn(t, root)

	assert.Equal(t, 0, code)
	assert.Contains(t, out, "OK")
}

// 別スライスへ詰め替える中間変数は、返却値ではないので違反にしない。
func Test_詰め替え用の中間変数は違反にしない(t *testing.T) {
	root := writePersistence(t, "repo.go", `package persistence

type r struct{}

func (r *r) ListAll() ([]Out, error) {
	var rows []row
	out := make([]Out, 0, len(rows))
	for _, x := range rows {
		out = append(out, Out{ID: x.ID})
	}
	return out, nil
}
`)
	code, _ := runOn(t, root)

	assert.Equal(t, 0, code)
}

// スライスを返さないメソッド（単一取得・件数集計など）は対象外。
func Test_スライスを返さないメソッドは対象外(t *testing.T) {
	root := writePersistence(t, "repo.go", `package persistence

type r struct{}

func (r *r) CountByUser() (map[uint64]int, error) {
	var rows []row
	counts := map[uint64]int{}
	for _, x := range rows {
		counts[x.ID] = x.N
	}
	return counts, nil
}
`)
	code, _ := runOn(t, root)

	assert.Equal(t, 0, code)
}

func Test_allowコメントで個別に抑制できる(t *testing.T) {
	root := writePersistence(t, "repo.go", `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	//slicelint:allow 呼び出し側が nil を許容する内部専用メソッド
	var rows []Item
	return rows, nil
}
`)
	code, out := runOn(t, root)

	assert.Equal(t, 0, code, "抑制コメントがあれば通ること: %s", out)
}

func Test_ignore_fileでファイル全体を除外できる(t *testing.T) {
	root := writePersistence(t, "repo.go", `//slicelint:ignore-file 生成コードのため対象外

package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	var rows []Item
	return rows, nil
}
`)
	code, _ := runOn(t, root)

	assert.Equal(t, 0, code)
}

func Test_テストファイルは対象外(t *testing.T) {
	root := writePersistence(t, "repo_test.go", `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	var rows []Item
	return rows, nil
}
`)
	code, _ := runOn(t, root)

	assert.Equal(t, 0, code)
}

func Test_解析できないディレクトリは実行エラー(t *testing.T) {
	var out, errOut bytes.Buffer
	code := runCLI([]string{filepath.Join(t.TempDir(), "存在しない")}, &out, &errOut)

	assert.Equal(t, 2, code)
	assert.Contains(t, errOut.String(), "slicelint:")
}

func Test_複数の違反を行番号順に並べて報告する(t *testing.T) {
	root := writePersistence(t, "repo.go", `package persistence

type r struct{}

func (r *r) ListA() ([]Item, error) {
	var a []Item
	return a, nil
}

func (r *r) ListB() ([]Item, error) {
	var b []Item
	return b, nil
}
`)
	code, out := runOn(t, root)

	assert.Equal(t, 1, code)
	assert.Contains(t, out, "ListA")
	assert.Contains(t, out, "ListB")
	assert.Less(t, indexOf(out, "ListA"), indexOf(out, "ListB"), "行番号順に並ぶこと")
}

func indexOf(s, sub string) int {
	return bytes.Index([]byte(s), []byte(sub))
}
