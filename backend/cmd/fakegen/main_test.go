package main

import (
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// renderType の各ケースを検証する（純粋関数なのでテストしやすい）。
func Test_renderType(t *testing.T) {
	cases := map[string]string{
		"uint64":                    "uint64",                               // 組み込み型はそのまま
		"string":                    "string",                               //
		"error":                     "error",                                //
		"context.Context":           "context.Context",                      // selector はそのまま
		"*domain.User":              "*domain.User",                         // ポインタ + selector
		"[]domain.Note":             "[]domain.Note",                        // スライス
		"AiChatAttachmentUploadURL": "repository.AiChatAttachmentUploadURL", // 局所 exported 型は repository. 修飾
		"*UploadURL":                "*repository.UploadURL",                // ポインタ局所型
		"map[string]int":            "map[string]int",                       // map
	}
	for src, want := range cases {
		expr, err := parser.ParseExpr(src)
		if err != nil {
			t.Fatalf("ParseExpr(%q): %v", src, err)
		}
		used := map[string]bool{}
		got := renderType(expr, used)
		if got != want {
			t.Errorf("renderType(%q) = %q, want %q", src, got, want)
		}
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
	got := expandFieldTypes(ft.Params, map[string]bool{})
	want := []string{"uint64", "uint64", "uint64", "string"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("expandFieldTypes = %v, want %v", got, want)
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
}

type Note struct{ ID uint64 }
`)

	code := run([]string{root}, &strings.Builder{}, &strings.Builder{})
	if code != 0 {
		t.Fatalf("run exit = %d, want 0", code)
	}

	out := filepath.Join(root, "internal", "usecase", "repository", "repofakes", "fakes_gen.go")
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("生成ファイルが読めません: %v", err)
	}
	// gofmt の桁揃え(struct フィールド)で空白数が変わるため、水平方向の空白を 1 つに正規化して比較する。
	gen := collapseSpaces(string(data))

	for _, want := range []string{
		"package repofakes",
		"type FakeNoteRepository struct",
		"FindByIDFunc func(context.Context, uint64) (*repository.Note, error)", // 局所型 Note → repository.Note
		"FindByIDCalls int",
		"var _ repository.NoteRepository = (*FakeNoteRepository)(nil)", // interface 充足の保証
		"func (f *FakeNoteRepository) FindByID(p0 context.Context, p1 uint64) (*repository.Note, error)",
		"f.FindByIDCalls++",
		"func (f *FakeNoteRepository) Count(p0 context.Context) int",                       // 単一戻り値
		"func (f *FakeNoteRepository) Save(p0 context.Context, p1 *repository.Note) error", // error のみ
	} {
		if !strings.Contains(gen, collapseSpaces(want)) {
			t.Errorf("生成コードに %q が含まれていません\n---\n%s", want, gen)
		}
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
