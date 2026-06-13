package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPrefix = "github.com/norman6464/FreStyle/backend/internal/"

func TestClassifyRel(t *testing.T) {
	cases := map[string]string{
		"domain":                       layerDomain,
		"domain/sub":                   layerDomain,
		"usecase":                      layerUsecase,
		"usecase/foo":                  layerUsecase,
		"usecase/repository":           layerUsecaseRepo,
		"usecase/repository/sub":       layerUsecaseRepo,
		"handler":                      layerHandler,
		"handler/middleware":           layerHandler,
		"adapter/persistence":          layerPersistence,
		"adapter/persistence/whatever": layerPersistence,
		"adapter":                      layerPersistence,
		"infra":                        layerInfra,
		"infra/cognito":                layerInfra,
		"cmd/server":                   "",
		"":                             "",
	}
	for rel, want := range cases {
		if got := classifyRel(rel); got != want {
			t.Errorf("classifyRel(%q) = %q, want %q", rel, got, want)
		}
	}
}

func TestClassifyImport(t *testing.T) {
	cases := map[string]string{
		"net/http":                         targetNetHTTP,
		"github.com/gin-gonic/gin":         targetGin,
		"github.com/gin-gonic/gin/binding": targetGin,
		testPrefix + "domain":              layerDomain,
		testPrefix + "usecase":             layerUsecase,
		testPrefix + "usecase/repository":  layerUsecaseRepo,
		testPrefix + "handler/middleware":  layerHandler,
		testPrefix + "adapter/persistence": layerPersistence,
		testPrefix + "infra/s3":            layerInfra,
		"github.com/stretchr/testify/mock": "",
		"context":                          "",
		"github.com/gin-gonic/ginkgo":      "", // gin で始まるが別物
	}
	for path, want := range cases {
		if got := classifyImport(path, testPrefix); got != want {
			t.Errorf("classifyImport(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestIsWiringFile(t *testing.T) {
	cases := map[string]bool{
		"router.go":                     true,
		"routes_auth.go":                true,
		"routes_company_application.go": true,
		"auth_handler.go":               false,
		"handler.go":                    false,
		"routes.go":                     false, // routes_ プレフィックスではない
	}
	for name, want := range cases {
		if got := isWiringFile(name); got != want {
			t.Errorf("isWiringFile(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestViolationsFor(t *testing.T) {
	imp := func(target string) importRef { return importRef{path: "p", line: 1, target: target} }

	t.Run("domain が usecase を import するのは違反", func(t *testing.T) {
		vs := violationsFor(layerDomain, "domain", "user.go", "internal/domain/user.go", []importRef{imp(layerUsecase)})
		if len(vs) != 1 {
			t.Fatalf("want 1 violation, got %d", len(vs))
		}
	})

	t.Run("usecase が persistence を import するのは違反（DIP）", func(t *testing.T) {
		vs := violationsFor(layerUsecase, "usecase", "foo_usecase.go", "p", []importRef{imp(layerPersistence)})
		if len(vs) != 1 {
			t.Fatalf("want 1 violation, got %d", len(vs))
		}
	})

	t.Run("usecase が infra を import するのは許容", func(t *testing.T) {
		vs := violationsFor(layerUsecase, "usecase", "foo_usecase.go", "p", []importRef{imp(layerInfra)})
		if len(vs) != 0 {
			t.Fatalf("want 0 violation, got %d", len(vs))
		}
	})

	t.Run("usecase が gin を import するのは違反", func(t *testing.T) {
		vs := violationsFor(layerUsecase, "usecase", "foo_usecase.go", "p", []importRef{imp(targetGin)})
		if len(vs) != 1 {
			t.Fatalf("want 1 violation, got %d", len(vs))
		}
	})

	t.Run("wiring ファイルでの handler→persistence import は許容", func(t *testing.T) {
		vs := violationsFor(layerHandler, "handler", "router.go", "p", []importRef{imp(layerPersistence)})
		if len(vs) != 0 {
			t.Fatalf("wiring file should be exempt, got %d", len(vs))
		}
	})

	t.Run("非 wiring ファイルでの handler→persistence import は違反", func(t *testing.T) {
		vs := violationsFor(layerHandler, "handler", "foo_handler.go", "p", []importRef{imp(layerPersistence)})
		if len(vs) != 1 {
			t.Fatalf("want 1 violation, got %d", len(vs))
		}
	})

	t.Run("handler サブディレクトリの wiring 名ファイルは例外扱いしない", func(t *testing.T) {
		// handler/middleware/routes_x.go のような別ディレクトリでは wiring 例外を効かせない。
		vs := violationsFor(layerHandler, "handler/middleware", "routes_x.go", "p", []importRef{imp(layerPersistence)})
		if len(vs) != 1 {
			t.Fatalf("subdir wiring-named file should NOT be exempt, got %d", len(vs))
		}
	})

	t.Run("handler が infra を import するのは許容（実用上の例外）", func(t *testing.T) {
		vs := violationsFor(layerHandler, "handler", "auth_handler.go", "p", []importRef{imp(layerInfra)})
		if len(vs) != 0 {
			t.Fatalf("want 0 violation, got %d", len(vs))
		}
	})

	t.Run("infra が net/http を import するのは許容（クライアント用途）", func(t *testing.T) {
		vs := violationsFor(layerInfra, "infra/embed", "fetcher.go", "p", []importRef{imp(targetNetHTTP)})
		if len(vs) != 0 {
			t.Fatalf("want 0 violation, got %d", len(vs))
		}
	})

	t.Run("persistence が usecase 本体を import するのは違反", func(t *testing.T) {
		vs := violationsFor(layerPersistence, "adapter/persistence", "user_repository.go", "p", []importRef{imp(layerUsecase)})
		if len(vs) != 1 {
			t.Fatalf("want 1 violation, got %d", len(vs))
		}
	})

	t.Run("persistence が port を import するのは許容", func(t *testing.T) {
		vs := violationsFor(layerPersistence, "adapter/persistence", "user_repository.go", "p", []importRef{imp(layerUsecaseRepo)})
		if len(vs) != 0 {
			t.Fatalf("want 0 violation, got %d", len(vs))
		}
	})

	t.Run("抑制された import はスキップ", func(t *testing.T) {
		ref := importRef{path: "p", line: 1, target: layerUsecase, suppressed: true}
		vs := violationsFor(layerDomain, "domain", "user.go", "p", []importRef{ref})
		if len(vs) != 0 {
			t.Fatalf("suppressed import should be skipped, got %d", len(vs))
		}
	})
}

// TestParseImports は実ファイルを書き出して import 解析と抑制を検証する。
func TestParseImports(t *testing.T) {
	dir := t.TempDir()

	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	t.Run("分類し抑制を検出する", func(t *testing.T) {
		p := write("a.go", `package x

import (
	"context"
	"net/http" //archlint:allow テスト用
	usecase "`+testPrefix+`usecase"
)

var _ = context.Background
`)
		imports, ignore, err := parseImports(p, testPrefix)
		if err != nil {
			t.Fatal(err)
		}
		if ignore {
			t.Fatal("ignoreFile should be false")
		}
		var sawNetHTTP, sawUsecase bool
		for _, imp := range imports {
			switch imp.target {
			case targetNetHTTP:
				sawNetHTTP = true
				if !imp.suppressed {
					t.Error("net/http should be suppressed by //archlint:allow")
				}
			case layerUsecase:
				sawUsecase = true
				if imp.suppressed {
					t.Error("usecase import should not be suppressed")
				}
			}
		}
		if !sawNetHTTP || !sawUsecase {
			t.Errorf("expected both imports parsed, netHTTP=%v usecase=%v", sawNetHTTP, sawUsecase)
		}
	})

	t.Run("file-level ignore (先頭コメント)", func(t *testing.T) {
		p := write("b.go", `//archlint:ignore-file 生成ファイルのため除外
package x

import _ "`+testPrefix+`adapter/persistence"
`)
		_, ignore, err := parseImports(p, testPrefix)
		if err != nil {
			t.Fatal(err)
		}
		if !ignore {
			t.Fatal("ignoreFile should be true")
		}
	})

	t.Run("mid-file ignore-file は無効", func(t *testing.T) {
		// package 宣言より後のコメントでは抑制できない（回避防止）。
		p := write("c.go", `package x

// archlint:ignore-file ここでは効かないはず
import _ "`+testPrefix+`adapter/persistence"
`)
		_, ignore, err := parseImports(p, testPrefix)
		if err != nil {
			t.Fatal(err)
		}
		if ignore {
			t.Fatal("ignoreFile should be false for non-leading comment")
		}
	})
}

// buildTree は層ディレクトリを模した一時ツリーを作り、root を返す。
func buildTree(t *testing.T, withViolation bool) string {
	t.Helper()
	root := t.TempDir()
	internal := filepath.Join(root, "internal")

	if err := os.WriteFile(filepath.Join(root, "go.mod"),
		[]byte("module github.com/norman6464/FreStyle/backend\n\ngo 1.26\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mk := func(rel, content string) {
		p := filepath.Join(internal, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// 正常: usecase が repository(port) を import
	mk("usecase/ok.go", "package usecase\nimport _ \""+testPrefix+"usecase/repository\"\n")
	// 正常: wiring ファイルなので handler→persistence は許容
	mk("handler/router.go", "package handler\nimport _ \""+testPrefix+"adapter/persistence\"\n")
	// テストファイルは無視される（違反 import を入れても拾わない）
	mk("domain/bad_test.go", "package domain\nimport _ \""+testPrefix+"handler\"\n")

	if withViolation {
		// 違反: domain が usecase を import
		mk("domain/bad.go", "package domain\nimport _ \""+testPrefix+"usecase\"\n")
	}
	return root
}

// TestRunIntegration は run の違反検出（とテストファイル無視・wiring 例外）を確認する。
func TestRunIntegration(t *testing.T) {
	root := buildTree(t, true)
	vs, err := run(filepath.Join(root, "internal"), testPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 1 {
		t.Fatalf("want exactly 1 violation (domain→usecase), got %d: %+v", len(vs), vs)
	}
	if filepath.Base(vs[0].file) != "bad.go" {
		t.Errorf("violation should be in bad.go, got %s", vs[0].file)
	}
}

func TestRunCLI(t *testing.T) {
	t.Run("クリーンなツリーは 0 を返す", func(t *testing.T) {
		root := buildTree(t, false)
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{root}, &out, &errBuf); code != 0 {
			t.Fatalf("want exit 0, got %d (stderr=%s)", code, errBuf.String())
		}
		if !strings.Contains(out.String(), "OK") {
			t.Errorf("stdout should report OK, got %q", out.String())
		}
	})

	t.Run("違反のあるツリーは 1 を返す", func(t *testing.T) {
		root := buildTree(t, true)
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{root}, &out, &errBuf); code != 1 {
			t.Fatalf("want exit 1, got %d", code)
		}
		if !strings.Contains(out.String(), "bad.go") {
			t.Errorf("stdout should name the offending file, got %q", out.String())
		}
	})

	t.Run("go.mod なしは 2 を返す", func(t *testing.T) {
		var out, errBuf bytes.Buffer
		if code := runCLI([]string{t.TempDir()}, &out, &errBuf); code != 2 {
			t.Fatalf("want exit 2, got %d", code)
		}
	})
}

func TestReadModulePath(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(good, []byte("// comment\nmodule example.com/foo/bar\n\ngo 1.26\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readModulePath(good)
	if err != nil {
		t.Fatal(err)
	}
	if got != "example.com/foo/bar" {
		t.Errorf("readModulePath = %q, want example.com/foo/bar", got)
	}

	// タブ区切り・複数スペースでも読めること。
	tabbed := filepath.Join(dir, "tab.mod")
	if err := os.WriteFile(tabbed, []byte("module\texample.com/tab\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := readModulePath(tabbed); err != nil || got != "example.com/tab" {
		t.Errorf("readModulePath(tab) = %q, %v; want example.com/tab", got, err)
	}

	if _, err := readModulePath(filepath.Join(dir, "missing.go.mod")); err == nil {
		t.Error("want error for missing go.mod")
	}

	noModule := filepath.Join(dir, "no-module.mod")
	if err := os.WriteFile(noModule, []byte("go 1.26\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readModulePath(noModule); err == nil {
		t.Error("want error when module line is absent")
	}
}
