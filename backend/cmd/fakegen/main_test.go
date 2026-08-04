package main

import (
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mustParseType は型を表す式を parse する。
// 可変長 "...T" は関数の引数位置にしか現れないため、その場合だけ関数を経由して取り出す。
func mustParseType(t *testing.T, src string) ast.Expr {
	t.Helper()
	if strings.HasPrefix(src, "...") {
		expr, err := parser.ParseExpr("func(x " + src + ")")
		if err != nil {
			t.Fatalf("ParseExpr(func(x %s)): %v", src, err)
		}
		return expr.(*ast.FuncType).Params.List[0].Type
	}
	expr, err := parser.ParseExpr(src)
	if err != nil {
		t.Fatalf("ParseExpr(%q): %v", src, err)
	}
	return expr
}

// renderType の各分岐を検証する。usedAliases への書き込みは import 生成の入力になるため、
// 戻り値と同じくらい重要な出力として一緒に検証する。
func Test_renderType(t *testing.T) {
	cases := []struct {
		name            string
		src             string
		want            string
		wantAliases     []string // renderType が立てるべきエイリアス
		wantUnsupported bool     // 近似したことを記録すべきか
	}{
		{name: "組み込み型はそのまま", src: "uint64", want: "uint64"},
		{name: "string もそのまま", src: "string", want: "string"},
		{name: "error もそのまま", src: "error", want: "error"},
		{name: "selector はそのまま/エイリアスを記録", src: "context.Context", want: "context.Context", wantAliases: []string{"context"}},
		{name: "ポインタ + selector", src: "*domain.User", want: "*domain.User", wantAliases: []string{"domain"}},
		{name: "スライス", src: "[]domain.Note", want: "[]domain.Note", wantAliases: []string{"domain"}},
		{name: "固定長配列", src: "[4]byte", want: "[4]byte"},
		{name: "局所 exported 型は repository. 修飾", src: "AiChatAttachmentUploadURL", want: "repository.AiChatAttachmentUploadURL", wantAliases: []string{"repository"}},
		{name: "ポインタ局所型", src: "*UploadURL", want: "*repository.UploadURL", wantAliases: []string{"repository"}},
		{name: "非 exported な局所型は無修飾", src: "internalKey", want: "internalKey"},
		{name: "map", src: "map[string]int", want: "map[string]int"},
		{name: "map の値が局所型", src: "map[string]Note", want: "map[string]repository.Note", wantAliases: []string{"repository"}},
		{name: "可変長", src: "...uint64", want: "...uint64"},
		{name: "可変長 + selector", src: "...domain.Note", want: "...domain.Note", wantAliases: []string{"domain"}},
		{name: "空 interface は any", src: "interface{}", want: "any"},
		{name: "メソッド付き無名 interface は any に近似して記録", src: "interface{ Foo() }", want: "any", wantUnsupported: true},
		{name: "未対応の型は any に近似して記録", src: "chan int", want: "any", wantUnsupported: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newGenCtx()
			got := renderType(mustParseType(t, tc.src), ctx)
			if got != tc.want {
				t.Errorf("renderType(%q) = %q, want %q", tc.src, got, tc.want)
			}
			for _, alias := range tc.wantAliases {
				if !ctx.usedAliases[alias] {
					t.Errorf("renderType(%q): usedAliases[%q] が立っていません (%v)", tc.src, alias, ctx.usedAliases)
				}
			}
			if len(ctx.usedAliases) != len(tc.wantAliases) {
				t.Errorf("renderType(%q): usedAliases = %v, want %v", tc.src, ctx.usedAliases, tc.wantAliases)
			}
			if gotUnsupported := len(ctx.unsupported) > 0; gotUnsupported != tc.wantUnsupported {
				t.Errorf("renderType(%q): unsupported = %v, want %v", tc.src, ctx.unsupported, tc.wantUnsupported)
			}
		})
	}
}

// expandFieldTypes はグループ化引数を 1 つずつに展開する。
func Test_expandFieldTypes_groupedParams(t *testing.T) {
	// func(userID, materialID, courseID uint64, name string) のパラメータ部を作る。
	expr, err := parser.ParseExpr("func(userID, materialID, courseID uint64, name string)")
	if err != nil {
		t.Fatal(err)
	}
	ft := expr.(*ast.FuncType)
	got := expandFieldTypes(ft.Params, newGenCtx())
	want := []string{"uint64", "uint64", "uint64", "string"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("expandFieldTypes = %v, want %v", got, want)
	}
}

func Test_expandFieldTypes_nilFieldList(t *testing.T) {
	if got := expandFieldTypes(nil, newGenCtx()); got != nil {
		t.Errorf("expandFieldTypes(nil) = %v, want nil", got)
	}
}

// defaultPkgName は import path から実際のパッケージ名を返す。
// メジャーバージョン接尾辞を素通しすると alias 解決がずれて import 漏れになる。
func Test_defaultPkgName(t *testing.T) {
	cases := []struct{ path, want string }{
		{"context", "context"},
		{"sync/atomic", "atomic"},
		{"github.com/foo/bar", "bar"},
		{"github.com/foo/bar/v2", "bar"},
		{"github.com/foo/bar/v10", "bar"},
		{"github.com/foo/bar/v1", "v1"},     // v0/v1 は接尾辞として現れないのでそのまま
		{"github.com/foo/bar/v0", "v0"},     //
		{"github.com/foo/verify", "verify"}, // v で始まるだけの要素はバージョンではない
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got := defaultPkgName(tc.path); got != tc.want {
				t.Errorf("defaultPkgName(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func Test_isStdlib(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"context", true},
		{"sync/atomic", true},
		{"github.com/foo/bar", false},
		{"example.com/x/internal/usecase/repository", false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got := isStdlib(tc.path); got != tc.want {
				t.Errorf("isStdlib(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func Test_readModulePath(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{name: "通常", content: "module example.com/x\n\ngo 1.22\n", want: "example.com/x"},
		{name: "行末コメントを落とす", content: "module example.com/x // 説明\n", want: "example.com/x"},
		{name: "module 行が無い", content: "go 1.22\n", wantErr: true},
		{name: "module 行にパスが無い", content: "module // 説明のみ\n", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "go.mod")
			mustWrite(t, path, tc.content)
			got, err := readModulePath(path)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("readModulePath = %q, エラーを期待", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("readModulePath: %v", err)
			}
			if got != tc.want {
				t.Errorf("readModulePath = %q, want %q", got, tc.want)
			}
		})
	}
}

func Test_readModulePath_ファイルが無い(t *testing.T) {
	if _, err := readModulePath(filepath.Join(t.TempDir(), "go.mod")); err == nil {
		t.Fatal("存在しない go.mod でエラーを期待")
	}
}

// run の end-to-end: 小さな repository パッケージを fixture に、fake が生成されることを確認する。
func Test_run_generatesFakes(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module example.com/x\n\ngo 1.22\n")
	repoDir := filepath.Join(root, "internal", "usecase", "repository")
	mustWrite(t, filepath.Join(repoDir, "note.go"), `package repository

import "context"

type NoteRepository interface {
	FindByID(ctx context.Context, id uint64) (*Note, error)
	Save(ctx context.Context, n *Note) error
	Count(ctx context.Context) int
	Touch(ctx context.Context)
	NotifyAll(ctx context.Context, ids ...uint64) error
}

type Note struct{ ID uint64 }
`)
	// _test.go は解析対象外。ここの interface が生成されたら除外が壊れている。
	mustWrite(t, filepath.Join(repoDir, "note_test.go"), `package repository

type ShouldNotBeGenerated interface {
	Nope()
}
`)

	var stdout, stderr strings.Builder
	if code := run([]string{root}, &stdout, &stderr); code != 0 {
		t.Fatalf("run exit = %d, want 0 (stderr=%s)", code, stderr.String())
	}

	out := filepath.Join(repoDir, "repofakes", "fakes_gen.go")
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("生成ファイルが読めません: %v", err)
	}
	raw := string(data)
	// gofmt の桁揃え(struct フィールド)で空白数が変わるため、水平方向の空白を 1 つに正規化して比較する。
	gen := collapseSpaces(raw)

	for _, want := range []string{
		"package repofakes",
		"type FakeNoteRepository struct",
		"FindByIDFunc func(context.Context, uint64) (*repository.Note, error)", // 局所型 Note → repository.Note
		"FindByIDCalls atomic.Int64",
		"var _ repository.NoteRepository = (*FakeNoteRepository)(nil)", // interface 充足の保証
		"func (f *FakeNoteRepository) FindByID(p0 context.Context, p1 uint64) (*repository.Note, error)",
		"f.FindByIDCalls.Add(1)",
		"func (f *FakeNoteRepository) Count(p0 context.Context) int",                       // 単一戻り値
		"func (f *FakeNoteRepository) Save(p0 context.Context, p1 *repository.Note) error", // error のみ
		"func (f *FakeNoteRepository) Touch(p0 context.Context) {",                         // 戻り値なし
		"f.TouchFunc(p0)",
		"func (f *FakeNoteRepository) NotifyAll(p0 context.Context, p1 ...uint64) error", // 可変長
		"return f.NotifyAllFunc(p0, p1...)",                                              // 可変長の呼び出しは展開する
	} {
		if !strings.Contains(gen, collapseSpaces(want)) {
			t.Errorf("生成コードに %q が含まれていません\n---\n%s", want, gen)
		}
	}
	if strings.Contains(gen, "ShouldNotBeGenerated") {
		t.Error("_test.go の interface が生成されています（除外が効いていません）")
	}

	// import は標準ライブラリと外部を空行で分ける。生成物は gofumpt の対象外なので
	// fakegen 自身がこの形で出す必要がある。
	wantImports := "import (\n\t\"context\"\n\t\"sync/atomic\"\n\n\t\"example.com/x/internal/usecase/repository\"\n)"
	if !strings.Contains(raw, wantImports) {
		t.Errorf("import のグループ分けが期待と異なります\n--- want ---\n%s\n--- got ---\n%s", wantImports, raw)
	}
}

func Test_run_goModが無い場合(t *testing.T) {
	var stdout, stderr strings.Builder
	if code := run([]string{t.TempDir()}, &stdout, &stderr); code != 1 {
		t.Fatalf("run exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "go.mod を読めません") {
		t.Errorf("stderr に理由が出ていません: %q", stderr.String())
	}
}

func Test_run_repositoryディレクトリが無い場合(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module example.com/x\n")

	var stdout, stderr strings.Builder
	if code := run([]string{root}, &stdout, &stderr); code != 1 {
		t.Fatalf("run exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "を読めません") {
		t.Errorf("stderr に理由が出ていません: %q", stderr.String())
	}
}

func Test_run_構文エラーのファイルがある場合(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module example.com/x\n")
	mustWrite(t, filepath.Join(root, "internal", "usecase", "repository", "broken.go"),
		"package repository\n\ntype Broken interface {\n")

	var stdout, stderr strings.Builder
	if code := run([]string{root}, &stdout, &stderr); code != 1 {
		t.Fatalf("run exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "parse に失敗") {
		t.Errorf("stderr に理由が出ていません: %q", stderr.String())
	}
}

// interface が 1 つも無いときは exit 0 だが生成もしない。
// 「成功したのにファイルが無い」は誤解しやすいので挙動を固定しておく。
func Test_run_interfaceが1つも無い場合(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module example.com/x\n")
	repoDir := filepath.Join(root, "internal", "usecase", "repository")
	mustWrite(t, filepath.Join(repoDir, "types.go"), "package repository\n\ntype Note struct{ ID uint64 }\n")

	var stdout, stderr strings.Builder
	if code := run([]string{root}, &stdout, &stderr); code != 0 {
		t.Fatalf("run exit = %d, want 0 (stderr=%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "見つかりませんでした") {
		t.Errorf("stdout に通知が出ていません: %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(repoDir, "repofakes", "fakes_gen.go")); !os.IsNotExist(err) {
		t.Errorf("interface が無いのに生成ファイルができています (err=%v)", err)
	}
}

// 埋め込み interface は展開できずスキップする。黙って落とすと fake が interface を
// 満たさなくなるため、警告が stderr に出ることを固定する。
func Test_run_埋め込みinterfaceを警告する(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module example.com/x\n")
	mustWrite(t, filepath.Join(root, "internal", "usecase", "repository", "note.go"), `package repository

type Base interface {
	Ping() error
}

type NoteRepository interface {
	Base
	Save() error
}
`)

	var stdout, stderr strings.Builder
	if code := run([]string{root}, &stdout, &stderr); code != 0 {
		t.Fatalf("run exit = %d, want 0 (stderr=%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "埋め込み interface") {
		t.Errorf("埋め込み interface の警告が出ていません: %q", stderr.String())
	}
}

// collapseSpaces は水平方向の連続空白(スペース/タブ)を 1 つにまとめる（gofmt 桁揃え対策）。
func collapseSpaces(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !prevSpace {
				b.WriteByte(' ')
			}
			prevSpace = true
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
