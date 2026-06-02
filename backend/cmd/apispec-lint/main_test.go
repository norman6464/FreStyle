package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// parseSrc は Go ソース文字列を *ast.File に解析する小道具。
func parseSrc(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return fset, f
}

func TestExtractAnnotatedMethods(t *testing.T) {
	_, f := parseSrc(t, `package handler

// Create は申請を作る。
//
//	@Router /company-applications [post]
func (h *H) Create(c *gin.Context) {}

// List はドキュメント無しのメソッド。
func (h *H) List(c *gin.Context) {}

// freeFunc は @Router を持つがレシーバ無し（メソッドではない）。
//
//	@Router /x [get]
func freeFunc() {}
`)
	got := extractAnnotatedMethods(f)
	sort.Strings(got)
	if len(got) != 1 || got[0] != "Create" {
		t.Fatalf("extractAnnotatedMethods = %v, want [Create]", got)
	}
}

func TestExtractRoutes(t *testing.T) {
	fset, f := parseSrc(t, `package handler

func reg(g *gin.RouterGroup, h *H) {
	g.GET("/exercises", h.List)
	g.POST("/company-applications", middleware.RateLimitPerMinute(5, 5), h.Create)
	g.GET("/stream/:id", h.Stream) //apispec:allow SSE
	foo.Bar("/not-a-route", h.Nope)
	g.GET(pathVar, h.Dynamic)
}
`)
	routes := extractRoutes(fset, f)

	byHandler := map[string]route{}
	for _, r := range routes {
		byHandler[r.handler] = r
	}

	if r, ok := byHandler["List"]; !ok || r.method != "GET" || r.path != "/exercises" || r.suppressed {
		t.Errorf("List route wrong: %+v ok=%v", r, ok)
	}
	// ミドルウェア引数があっても handler は最後の引数（Create）。
	if r, ok := byHandler["Create"]; !ok || r.method != "POST" || r.suppressed {
		t.Errorf("Create route wrong: %+v ok=%v", r, ok)
	}
	// 行末 //apispec:allow で suppressed。
	if r, ok := byHandler["Stream"]; !ok || !r.suppressed {
		t.Errorf("Stream route should be suppressed: %+v ok=%v", r, ok)
	}
	// HTTP メソッドでない foo.Bar は拾わない。
	if _, ok := byHandler["Nope"]; ok {
		t.Error("foo.Bar should not be treated as a route")
	}
	// 第 1 引数が文字列リテラルでない（pathVar）ものは拾わない。
	if _, ok := byHandler["Dynamic"]; ok {
		t.Error("non-literal path should be skipped")
	}
}

func TestHandlerName(t *testing.T) {
	mustExpr := func(src string) ast.Expr {
		_, f := parseSrc(t, "package p\nvar _ = "+src+"\n")
		return f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0]
	}
	if got := handlerName(mustExpr("h.Create")); got != "Create" {
		t.Errorf("selector handlerName = %q, want Create", got)
	}
	if got := handlerName(mustExpr("BareFunc")); got != "BareFunc" {
		t.Errorf("ident handlerName = %q, want BareFunc", got)
	}
	if got := handlerName(mustExpr("func() {}")); got != "" {
		t.Errorf("funcLit handlerName = %q, want empty", got)
	}
}

func TestHasIgnoreFile(t *testing.T) {
	_, lead := parseSrc(t, "//apispec:ignore-file 生成物\npackage handler\n")
	if !hasIgnoreFile(lead) {
		t.Error("leading //apispec:ignore-file should be detected")
	}

	_, mid := parseSrc(t, "package handler\n\n//apispec:ignore-file ここでは効かない\nvar x = 1\n")
	if hasIgnoreFile(mid) {
		t.Error("non-leading ignore-file must NOT apply")
	}
}

// mkHandlerTree は tmp/internal/handler 配下にファイルを作り handlerRoot を返す。
func mkHandlerTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	handlerRoot := filepath.Join(root, "internal", "handler")
	if err := os.MkdirAll(handlerRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(handlerRoot, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return handlerRoot
}

func TestRunIntegration(t *testing.T) {
	const annotatedHandler = `package handler

// Create は作る。
//
//	@Router /things [post]
func (h *H) Create(c *gin.Context) {}

// Get は取る。
//
//	@Router /things/{id} [get]
func (h *H) Get(c *gin.Context) {}
`

	t.Run("missing annotation is a violation", func(t *testing.T) {
		handlerRoot := mkHandlerTree(t, map[string]string{
			"h.go": annotatedHandler,
			"routes.go": `package handler
func reg(g *gin.RouterGroup, h *H) {
	g.POST("/things", h.Create)
	g.DELETE("/things/:id", h.Destroy)
}
`,
		})
		vs, err := run(handlerRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 1 || !strings.Contains(vs[0].msg, "Destroy") {
			t.Fatalf("want 1 violation for Destroy, got %+v", vs)
		}
	})

	t.Run("all annotated -> no violation", func(t *testing.T) {
		handlerRoot := mkHandlerTree(t, map[string]string{
			"h.go": annotatedHandler,
			"routes.go": `package handler
func reg(g *gin.RouterGroup, h *H) {
	g.POST("/things", h.Create)
	g.GET("/things/:id", h.Get)
}
`,
		})
		vs, err := run(handlerRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Fatalf("want 0 violation, got %+v", vs)
		}
	})

	t.Run("//apispec:allow suppresses", func(t *testing.T) {
		handlerRoot := mkHandlerTree(t, map[string]string{
			"h.go": annotatedHandler,
			"routes.go": `package handler
func reg(g *gin.RouterGroup, h *H) {
	g.GET("/stream/:id", h.Stream) //apispec:allow SSE は OpenAPI 対象外
}
`,
		})
		vs, err := run(handlerRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Fatalf("allow should suppress, got %+v", vs)
		}
	})

	t.Run("//apispec:ignore-file skips file", func(t *testing.T) {
		handlerRoot := mkHandlerTree(t, map[string]string{
			"h.go": annotatedHandler,
			"routes.go": `//apispec:ignore-file レガシー
package handler
func reg(g *gin.RouterGroup, h *H) {
	g.DELETE("/things/:id", h.Destroy)
}
`,
		})
		vs, err := run(handlerRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Fatalf("ignore-file should skip, got %+v", vs)
		}
	})
}

func TestRunCLI(t *testing.T) {
	handlerRoot := func(t *testing.T, violating bool) string {
		routes := "g.POST(\"/things\", h.Create)"
		if violating {
			routes += "\n\tg.DELETE(\"/things/:id\", h.Destroy)"
		}
		return mkHandlerTree(t, map[string]string{
			"h.go":      "package handler\n\n// Create は作る。\n//\n//\t@Router /things [post]\nfunc (h *H) Create(c *gin.Context) {}\n",
			"routes.go": "package handler\nfunc reg(g *gin.RouterGroup, h *H) {\n\t" + routes + "\n}\n",
		})
	}

	t.Run("clean returns 0", func(t *testing.T) {
		root := filepath.Dir(filepath.Dir(handlerRoot(t, false))) // root that contains internal/handler
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{root}, &out, &errBuf); code != 0 {
			t.Fatalf("want 0, got %d (%s)", code, errBuf.String())
		}
		if !strings.Contains(out.String(), "OK") {
			t.Errorf("want OK, got %q", out.String())
		}
	})

	t.Run("violation returns 1", func(t *testing.T) {
		root := filepath.Dir(filepath.Dir(handlerRoot(t, true)))
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{root}, &out, &errBuf); code != 1 {
			t.Fatalf("want 1, got %d", code)
		}
		if !strings.Contains(out.String(), "Destroy") {
			t.Errorf("want Destroy in output, got %q", out.String())
		}
	})
}
