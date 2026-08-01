package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writePersistence は一時ディレクトリに internal/adapter/persistence/<name> を作り root を返す。
func writePersistence(t *testing.T, name, src string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "internal", "adapter", "persistence")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644))
	return root
}

func Test_slicelint(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		src            string
		wantCode       int
		wantContains   []string // 出力に含まれること
		wantNotContain string   // 出力に含まれないこと
	}{
		{
			name:     "var宣言のまま返すと違反",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	var rows []Item
	return rows, nil
}
`,
			wantCode:     1,
			wantContains: []string{"ListAll", "make([]T, 0)"},
		},
		{
			name:     "空スライスで初期化していれば通る",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	rows := make([]Item, 0)
	return rows, nil
}
`,
			wantCode:     0,
			wantContains: []string{"OK"},
		},
		{
			// 早期 return で nil を返すと、この 1 箇所だけで同じ不具合が再発する。
			name:     "成功時にnilを直接返すと違反",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll(enabled bool) ([]Item, error) {
	if !enabled {
		return nil, nil
	}
	rows := make([]Item, 0)
	return rows, nil
}
`,
			wantCode:     1,
			wantContains: []string{"ListAll", "成功時に nil スライス"},
		},
		{
			// エラーを返す経路の nil は正当。ここまで弾くと実装できなくなる。
			name:     "エラー返却時のnilは違反にしない",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	rows, err := query()
	if err != nil {
		return nil, err
	}
	return rows, nil
}
`,
			wantCode: 0,
		},
		{
			name:     "詰め替え用の中間変数は違反にしない",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll() ([]Out, error) {
	var rows []row
	out := make([]Out, 0, len(rows))
	for _, x := range rows {
		out = append(out, Out{ID: x.ID})
	}
	return out, nil
}
`,
			wantCode: 0,
		},
		{
			name:     "スライスを返さないメソッドは対象外",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) CountByUser() (map[uint64]int, error) {
	var rows []row
	counts := map[uint64]int{}
	for _, x := range rows {
		counts[x.ID] = x.N
	}
	return counts, nil
}
`,
			wantCode: 0,
		},
		{
			name:     "単一取得のnilは違反にしない",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) FindByID(id uint64) (*Item, error) {
	if id == 0 {
		return nil, nil
	}
	return &Item{}, nil
}
`,
			wantCode: 0,
		},
		{
			name:     "allowコメントで宣言を抑制できる",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	//slicelint:allow 呼び出し側が nil を許容する内部専用メソッド
	var rows []Item
	return rows, nil
}
`,
			wantCode: 0,
		},
		{
			name:     "allowコメントを行末に書いても抑制できる",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	var rows []Item //slicelint:allow 内部専用
	return rows, nil
}
`,
			wantCode: 0,
		},
		{
			name:     "allowコメントでnil返却も抑制できる",
			fileName: "repo.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll(enabled bool) ([]Item, error) {
	if !enabled {
		//slicelint:allow 呼び出し側が未設定を nil で判定する
		return nil, nil
	}
	return make([]Item, 0), nil
}
`,
			wantCode: 0,
		},
		{
			name:     "ignore_fileでファイル全体を除外できる",
			fileName: "repo.go",
			src: `//slicelint:ignore-file 生成コードのため対象外

package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	var rows []Item
	return rows, nil
}
`,
			wantCode: 0,
		},
		{
			// 先頭コメントとしてのみ有効。package 宣言より後ろに書いても効かせない
			// （ファイル全体の除外を後から紛れ込ませられないようにする）。
			name:     "ignore_fileはpackage宣言より後ろでは効かない",
			fileName: "repo.go",
			src: `package persistence

//slicelint:ignore-file 後から追加された除外

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	var rows []Item
	return rows, nil
}
`,
			wantCode:     1,
			wantContains: []string{"ListAll"},
		},
		{
			name:     "テストファイルは対象外",
			fileName: "repo_test.go",
			src: `package persistence

type r struct{}

func (r *r) ListAll() ([]Item, error) {
	var rows []Item
	return rows, nil
}
`,
			wantCode: 0,
		},
		{
			// レシーバの無い純関数（ヘルパ）は repository のメソッドではないため対象外。
			name:     "レシーバの無い関数は対象外",
			fileName: "repo.go",
			src: `package persistence

func listAll() ([]Item, error) {
	var rows []Item
	return rows, nil
}
`,
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := writePersistence(t, tt.fileName, tt.src)

			var stdout, stderr bytes.Buffer
			code := runCLI([]string{root}, &stdout, &stderr)

			assert.Equal(t, tt.wantCode, code, "stdout=%s", stdout.String())
			for _, want := range tt.wantContains {
				assert.Contains(t, stdout.String(), want)
			}
			if tt.wantNotContain != "" {
				assert.NotContains(t, stdout.String(), tt.wantNotContain)
			}
		})
	}
}

func Test_解析できないディレクトリは実行エラー(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runCLI([]string{filepath.Join(t.TempDir(), "存在しない")}, &stdout, &stderr)

	assert.Equal(t, 2, code)
	assert.Contains(t, stderr.String(), "slicelint:")
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

	var stdout, stderr bytes.Buffer
	code := runCLI([]string{root}, &stdout, &stderr)
	out := stdout.String()

	assert.Equal(t, 1, code)
	assert.Less(t, strings.Index(out, "ListA"), strings.Index(out, "ListB"), "行番号順に並ぶこと")
	assert.Contains(t, stderr.String(), "2 件")
}
