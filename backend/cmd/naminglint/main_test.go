package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeFile(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", `package usecase

// CreateThingUseCase は作る。
type CreateThingUseCase struct{ r int }

func NewCreateThingUseCase(r int) *CreateThingUseCase { return &CreateThingUseCase{r} }

func (u *CreateThingUseCase) Execute() {}

// CourseUseCase は集約。
//
//naminglint:allow 集約
type CourseUseCase struct{ r int }

func NewCourseUseCase(r int) *CourseUseCase { return &CourseUseCase{r} }

func (u *CourseUseCase) List() {}

// notAUseCase は対象外。
type Helper struct{}
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	structs, ctors, execs := analyzeFile(fset, f, "x.go")

	if len(structs) != 2 {
		t.Fatalf("want 2 UseCase structs, got %d", len(structs))
	}
	byName := map[string]ucStruct{}
	for _, s := range structs {
		byName[s.name] = s
	}
	if s := byName["CreateThingUseCase"]; s.suppressed {
		t.Error("CreateThingUseCase should not be suppressed")
	}
	if s, ok := byName["CourseUseCase"]; !ok || !s.suppressed {
		t.Errorf("CourseUseCase should be suppressed: %+v ok=%v", s, ok)
	}
	if _, ok := byName["Helper"]; ok {
		t.Error("Helper (not *UseCase) should be ignored")
	}

	ctorSet := toSet(ctors)
	if !ctorSet["NewCreateThingUseCase"] || !ctorSet["NewCourseUseCase"] {
		t.Errorf("constructors missing: %v", ctors)
	}
	execSet := toSet(execs)
	if !execSet["CreateThingUseCase"] {
		t.Errorf("Execute recv missing CreateThingUseCase: %v", execs)
	}
	if execSet["CourseUseCase"] {
		t.Error("CourseUseCase has no Execute, should not be in execs")
	}
}

func toSet(ss []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func TestReceiverTypeName(t *testing.T) {
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "x.go", `package p
type T struct{}
func (u *T) A() {}
func (u T) B() {}
func Free() {}
`, parser.ParseComments)

	var got []string
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv != nil {
			got = append(got, receiverTypeName(fn.Recv))
		}
	}
	if len(got) != 2 || got[0] != "T" || got[1] != "T" {
		t.Fatalf("receiverTypeName results = %v, want [T T]", got)
	}
	if receiverTypeName(nil) != "" {
		t.Error("nil receiver should return empty")
	}
}

// mkUsecaseTree は tmp/internal/usecase 配下にファイルを作り usecaseRoot を返す。
func mkUsecaseTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	usecaseRoot := filepath.Join(root, "internal", "usecase")
	if err := os.MkdirAll(filepath.Join(usecaseRoot, "repository"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		full := filepath.Join(usecaseRoot, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return usecaseRoot
}

func TestRunIntegration(t *testing.T) {
	t.Run("missing Execute is a violation", func(t *testing.T) {
		usecaseRoot := mkUsecaseTree(t, map[string]string{
			"a.go": `package usecase
type FooUseCase struct{}
func NewFooUseCase() *FooUseCase { return &FooUseCase{} }
`,
		})
		vs, err := run(usecaseRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 1 || !strings.Contains(vs[0].msg, "Execute") {
			t.Fatalf("want 1 Execute violation, got %+v", vs)
		}
	})

	t.Run("missing constructor is a violation", func(t *testing.T) {
		usecaseRoot := mkUsecaseTree(t, map[string]string{
			"a.go": `package usecase
type FooUseCase struct{}
func (u *FooUseCase) Execute() {}
`,
		})
		vs, err := run(usecaseRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 1 || !strings.Contains(vs[0].msg, "NewFooUseCase") {
			t.Fatalf("want 1 constructor violation, got %+v", vs)
		}
	})

	t.Run("compliant usecase passes", func(t *testing.T) {
		usecaseRoot := mkUsecaseTree(t, map[string]string{
			"a.go": `package usecase
type FooUseCase struct{}
func NewFooUseCase() *FooUseCase { return &FooUseCase{} }
func (u *FooUseCase) Execute() {}
`,
		})
		vs, err := run(usecaseRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Fatalf("want 0 violation, got %+v", vs)
		}
	})

	t.Run("//naminglint:allow suppresses Execute requirement", func(t *testing.T) {
		usecaseRoot := mkUsecaseTree(t, map[string]string{
			"a.go": `package usecase

//naminglint:allow 集約
type FooUseCase struct{}
func NewFooUseCase() *FooUseCase { return &FooUseCase{} }
func (u *FooUseCase) List() {}
`,
		})
		vs, err := run(usecaseRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Fatalf("allow should suppress Execute requirement, got %+v", vs)
		}
	})

	t.Run("repository subpackage is ignored", func(t *testing.T) {
		usecaseRoot := mkUsecaseTree(t, map[string]string{
			"a.go": `package usecase
type FooUseCase struct{}
func NewFooUseCase() *FooUseCase { return &FooUseCase{} }
func (u *FooUseCase) Execute() {}
`,
			"repository/bad.go": `package repository
type BrokenUseCase struct{}
`,
		})
		vs, err := run(usecaseRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Fatalf("repository subdir should be skipped, got %+v", vs)
		}
	})

	t.Run("//naminglint:ignore-file skips file", func(t *testing.T) {
		usecaseRoot := mkUsecaseTree(t, map[string]string{
			"a.go": `//naminglint:ignore-file レガシー
package usecase
type FooUseCase struct{}
`,
		})
		vs, err := run(usecaseRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Fatalf("ignore-file should skip, got %+v", vs)
		}
	})
}

func TestRunCLI(t *testing.T) {
	good := mkUsecaseTree(t, map[string]string{
		"a.go": `package usecase
type FooUseCase struct{}
func NewFooUseCase() *FooUseCase { return &FooUseCase{} }
func (u *FooUseCase) Execute() {}
`,
	})
	root := filepath.Dir(filepath.Dir(good))
	var out, errBuf bytes.Buffer
	if code := runCLI([]string{root}, &out, &errBuf); code != 0 {
		t.Fatalf("want 0, got %d (%s)", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "OK") {
		t.Errorf("want OK, got %q", out.String())
	}

	bad := mkUsecaseTree(t, map[string]string{
		"a.go": `package usecase
type FooUseCase struct{}
`,
	})
	out.Reset()
	errBuf.Reset()
	if code := runCLI([]string{filepath.Dir(filepath.Dir(bad))}, &out, &errBuf); code != 1 {
		t.Fatalf("want 1, got %d", code)
	}
}
