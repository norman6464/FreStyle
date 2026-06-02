// Command apispec-lint は Gin に登録されたルートと swaggo 注釈の整合を検証する。
// CLAUDE.md §2.7「handler メソッドの直前に swaggo annotation を必ず書く」を機械化し、
// ルートを生やしたのに @Router 注釈を書き忘れた endpoint を CI で弾く。
//
// 仕組み:
//   - internal/handler の全 .go を go/ast で解析
//   - `g.GET("/path", ..., h.Method)` 形式のルート登録から (HTTP method, path, handler メソッド名) を抽出
//   - レシーバ付き func の doc コメントに @Router を含むメソッド名を「注釈あり」として収集
//   - ルートが指す handler メソッド名に @Router が一つも無ければ違反として報告
//
// 使い方:
//
//	go run ./cmd/apispec-lint           # backend/ 直下で実行
//	go run ./cmd/apispec-lint <root>    # 別ディレクトリを指定
//
// 違反は `path:line: メッセージ` 形式で出力し exit 1。
//
// 抑制（SSE / WebSocket / multipart など OpenAPI で表現しない endpoint 用）:
//   - ルート登録行の行末に `//apispec:allow <理由>`
//   - ファイル先頭コメントに `//apispec:ignore-file <理由>` でファイル全体
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// httpMethods は Gin のルート登録メソッド名。
var httpMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true,
}

// route は 1 件のルート登録。
type route struct {
	method     string // GET / POST ...
	path       string // "/company-applications"
	handler    string // ルートが指す handler メソッド名（"Create" 等）
	line       int
	suppressed bool // //apispec:allow が付いていれば true
}

// violation は 1 件の注釈漏れ。
type violation struct {
	file string
	line int
	msg  string
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

// runCLI は CLI 本体。exit code を返す（0=OK / 1=違反あり / 2=実行エラー）。
func runCLI(args []string, stdout, stderr io.Writer) int {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}
	handlerRoot := filepath.Join(root, "internal", "handler")

	violations, err := run(handlerRoot)
	if err != nil {
		fmt.Fprintf(stderr, "apispec-lint: %v\n", err)
		return 2
	}

	if len(violations) == 0 {
		fmt.Fprintln(stdout, "apispec-lint: OK — 全ルートに swaggo @Router 注釈があります")
		return 0
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].file != violations[j].file {
			return violations[i].file < violations[j].file
		}
		return violations[i].line < violations[j].line
	})
	for _, v := range violations {
		fmt.Fprintf(stdout, "%s:%d: %s\n", v.file, v.line, v.msg)
	}
	fmt.Fprintf(stderr, "\napispec-lint: %d 件の @Router 注釈漏れが見つかりました\n", len(violations))
	return 1
}

// run は handlerRoot 配下を 2 パスで解析する。
// 1 パス目で注釈ありメソッド名を全ファイル横断で集め、2 パス目でルートと突き合わせる。
func run(handlerRoot string) ([]violation, error) {
	fset := token.NewFileSet()

	type parsedFile struct {
		path   string
		file   *ast.File
		ignore bool
	}
	var files []parsedFile
	annotated := map[string]bool{}

	err := filepath.WalkDir(handlerRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return fmt.Errorf("%s: %w", path, perr)
		}
		for _, m := range extractAnnotatedMethods(f) {
			annotated[m] = true
		}
		files = append(files, parsedFile{path: path, file: f, ignore: hasIgnoreFile(f)})
		return nil
	})
	if err != nil {
		return nil, err
	}

	var out []violation
	for _, pf := range files {
		if pf.ignore {
			continue
		}
		for _, r := range extractRoutes(fset, pf.file) {
			if r.handler == "" || r.suppressed || annotated[r.handler] {
				continue
			}
			out = append(out, violation{
				file: pf.path,
				line: r.line,
				msg:  fmt.Sprintf("%s %s → %s に swaggo @Router 注釈がありません（CLAUDE.md §2.7）", r.method, r.path, r.handler),
			})
		}
	}
	return out, nil
}

// extractAnnotatedMethods は doc コメントに @Router を含むレシーバ付きメソッド名を返す。
func extractAnnotatedMethods(f *ast.File) []string {
	var names []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Doc == nil {
			continue
		}
		if commentContains(fn.Doc, "@Router") {
			names = append(names, fn.Name.Name)
		}
	}
	return names
}

// extractRoutes は Gin のルート登録呼び出しを抽出する。
func extractRoutes(fset *token.FileSet, f *ast.File) []route {
	allowLines := collectDirectiveLines(fset, f, "apispec:allow")

	var routes []route
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !httpMethods[sel.Sel.Name] || len(call.Args) < 2 {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		path, err := strconv.Unquote(lit.Value)
		if err != nil || !strings.HasPrefix(path, "/") {
			return true
		}

		start := fset.Position(call.Pos()).Line
		end := fset.Position(call.End()).Line
		routes = append(routes, route{
			method:     sel.Sel.Name,
			path:       path,
			handler:    handlerName(call.Args[len(call.Args)-1]),
			line:       start,
			suppressed: lineInRange(allowLines, start, end),
		})
		return true
	})
	return routes
}

// handlerName はルート最終引数（handler）からメソッド名を取り出す。
// `h.Create` のような selector を主に想定。inline func 等は "" を返して検査対象外にする。
func handlerName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.Ident:
		return e.Name
	default:
		return ""
	}
}

// hasIgnoreFile はファイル先頭（package 宣言より前）に //apispec:ignore-file があるかを返す。
func hasIgnoreFile(f *ast.File) bool {
	for _, cg := range f.Comments {
		if cg.End() > f.Package {
			break
		}
		if commentContains(cg, "apispec:ignore-file") {
			return true
		}
	}
	return false
}

// collectDirectiveLines は needle を含むコメントが存在する行番号の集合を返す。
// //apispec:xxx は go ディレクティブ形式で CommentGroup.Text() に出ないため生テキストを走査する。
func collectDirectiveLines(fset *token.FileSet, f *ast.File, needle string) map[int]bool {
	lines := map[int]bool{}
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, needle) {
				lines[fset.Position(c.Slash).Line] = true
			}
		}
	}
	return lines
}

// lineInRange は [start,end] のいずれかの行が set に含まれるかを返す。
func lineInRange(set map[int]bool, start, end int) bool {
	for l := start; l <= end; l++ {
		if set[l] {
			return true
		}
	}
	return false
}

// commentContains はコメント群の生テキストに needle が含まれるかを返す。
func commentContains(cg *ast.CommentGroup, needle string) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if strings.Contains(c.Text, needle) {
			return true
		}
	}
	return false
}
