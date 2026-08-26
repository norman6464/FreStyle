// Command slicelint は persistence 層の「一覧を返すメソッドが 0 件で nil を返す」書き方を検出する。
//
// nil スライスは encoding/json で null になり、handler がそのまま返すとフロントの
// map / filter / for-of が TypeError で落ちる。データがまだ無い新規ユーザー・新規コース・
// 未提出演習など、使い始めの動線を直撃する（FRESTYLE-70 で staging 実機で観測、
// FRESTYLE-77 で 14 メソッドの残存が判明）。
//
// この linter は単独では不変条件を守れない。sqlc の emit_empty_slices と二段構えになっている。
//
//  1. backend/sqlc.yaml の emit_empty_slices: true … sqlc 生成物（sqlcgen）の :many が
//     0 件のとき nil ではなく空スライスを返す
//  2. この linter                                   … 生成物を使う手書きの repository が
//     nil を返していないかを CI で検査する
//
// JSON までは 1 と 2 の両方を通る。片方だけ外すと穴が開く: emit_empty_slices を false に
// すると生成物が nil を返し始めるが、この linter は手書き側しか見ていないので気づけない。
// 逆にこの linter を外すと、生成物から domain 型へ詰め替えるところで nil に戻る書き方が復活する。
// どちらかを外すときは必ずもう片方も一緒に見直すこと（sqlc.yaml 側にも同じ注意書きがある）。
//
// 検査対象は internal/adapter/persistence 直下。スライスを返すメソッドの中で
// 戻り値になる変数が `var x []T`（nil）のまま宣言されているものを違反とする。
// `x := make([]T, 0)` のように空スライスで初期化すれば通る。
//
// 使い方:
//
//	go run ./cmd/slicelint            # backend/ 直下で実行
//	go run ./cmd/slicelint <root>     # 別ディレクトリを指定
//
// 違反は `path:line: メッセージ` 形式で出力し exit 1。
//
// 抑制（呼び出し側へ渡す前に必ず詰め替える中間変数など、正当な例外用）:
//   - 宣言の doc / 行末コメントに `//slicelint:allow <理由>`
//   - ファイル先頭コメントに `//slicelint:ignore-file <理由>`
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
	"strings"
)

// violation は 1 件の規約違反。
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
	persistenceRoot := filepath.Join(root, "internal", "adapter", "persistence")

	violations, err := run(persistenceRoot)
	if err != nil {
		fmt.Fprintf(stderr, "slicelint: %v\n", err)
		return 2
	}

	if len(violations) == 0 {
		fmt.Fprintln(stdout, "slicelint: OK — 0 件で nil を返す一覧メソッドなし")
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
	fmt.Fprintf(stderr, "\nslicelint: %d 件の nil スライス返却が見つかりました\n", len(violations))
	return 1
}

// run は persistenceRoot 直下の .go を解析し、違反を集める。
func run(persistenceRoot string) ([]violation, error) {
	fset := token.NewFileSet()
	var out []violation

	err := filepath.WalkDir(persistenceRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// サブパッケージは対象外（persistence 直下のみを見る）。
			if path != persistenceRoot {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return fmt.Errorf("parse %s: %w", path, perr)
		}
		if hasIgnoreFile(f) {
			return nil
		}
		out = append(out, analyzeFile(fset, f, path)...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// analyzeFile は 1 ファイルからスライス返却メソッドを取り出し、nil 宣言のまま返しているものを集める。
func analyzeFile(fset *token.FileSet, f *ast.File, path string) []violation {
	var out []violation
	allowLines := allowCommentLines(fset, f)

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Recv == nil {
			continue
		}
		sliceIdx := sliceResultIndex(fn)
		if sliceIdx < 0 {
			continue
		}

		// (1) `var x []T` のまま返している。
		// 戻り値としてそのまま返される識別子だけを対象にする。
		// 別スライスへ詰め替える中間変数（`out := make(...)` に append する等）は問題ない。
		returned := returnedIdents(fn)
		for _, d := range nilSliceDecls(fn) {
			if !returned[d.name] || d.suppressed {
				continue
			}
			out = append(out, violation{
				file: path,
				line: fset.Position(d.pos).Line,
				msg: fmt.Sprintf(
					"%s: %s は `var %s []T` のままで 0 件のとき nil を返します。"+
						"JSON が null になりフロントが落ちるため `%s := make([]T, 0)` にしてください"+
						"（FRESTYLE-77。中間変数など正当な例外は //slicelint:allow）",
					fn.Name.Name, d.name, d.name, d.name,
				),
			})
		}

		// (2) 成功時に nil リテラルをそのまま返している（`return nil, nil`）。
		// エラーを伴う `return nil, err` は正当なので除外する。
		for _, ret := range successNilReturns(fn, sliceIdx) {
			line := fset.Position(ret.Pos()).Line
			if allowLines[line] {
				continue
			}
			out = append(out, violation{
				file: path,
				line: line,
				msg: fmt.Sprintf(
					"%s: 成功時に nil スライスを返しています。JSON が null になりフロントが落ちるため "+
						"`make([]T, 0)` を返してください"+
						"（FRESTYLE-77。エラー返却時の nil は対象外。正当な例外は //slicelint:allow）",
					fn.Name.Name,
				),
			})
		}
	}
	return out
}

// sliceResultIndex は戻り値の中でスライス型が現れる位置（0 始まり）を返す。無ければ -1。
// `(a, b []T)` のように 1 フィールドに複数名がある形も展開して数える。
func sliceResultIndex(fn *ast.FuncDecl) int {
	if fn.Type.Results == nil {
		return -1
	}
	idx := 0
	for _, r := range fn.Type.Results.List {
		n := 1
		if len(r.Names) > 0 {
			n = len(r.Names)
		}
		if isSliceType(r.Type) {
			return idx
		}
		idx += n
	}
	return -1
}

// successNilReturns は「スライス位置が nil で、かつ他の戻り値もすべて nil」の return 文を返す。
// エラーを返している（error 位置が nil でない）return は正当なので含めない。
func successNilReturns(fn *ast.FuncDecl, sliceIdx int) []ast.Node {
	var out []ast.Node
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		// 内側の関数リテラル（コールバック）は別スコープなので追わない。
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || sliceIdx >= len(ret.Results) {
			return true
		}
		if !isNilIdent(ret.Results[sliceIdx]) {
			return true
		}
		for i, r := range ret.Results {
			if i == sliceIdx {
				continue
			}
			if !isNilIdent(r) {
				return true // エラー等を返しているので正当
			}
		}
		out = append(out, ret)
		return true
	})
	return out
}

// isNilIdent は式が nil リテラルかを返す。
func isNilIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "nil"
}

// isSliceType は式がスライス型（配列ではない）かを返す。
func isSliceType(e ast.Expr) bool {
	at, ok := e.(*ast.ArrayType)
	return ok && at.Len == nil
}

// nilDecl は `var x []T` 形式の宣言 1 件。
type nilDecl struct {
	name       string
	pos        token.Pos
	suppressed bool
}

// nilSliceDecls は関数本体から `var x []T`（初期値なし＝nil）の宣言を集める。
func nilSliceDecls(fn *ast.FuncDecl) []nilDecl {
	var out []nilDecl
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ds, ok := n.(*ast.DeclStmt)
		if !ok {
			return true
		}
		gd, ok := ds.Decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			return true
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Values) > 0 || vs.Type == nil {
				continue // 初期値ありは nil ではない
			}
			if !isSliceType(vs.Type) {
				continue
			}
			suppressed := commentContains(gd.Doc, "//slicelint:allow") ||
				commentContains(vs.Doc, "//slicelint:allow") ||
				commentContains(vs.Comment, "//slicelint:allow")
			for _, name := range vs.Names {
				out = append(out, nilDecl{name: name.Name, pos: name.Pos(), suppressed: suppressed})
			}
		}
		return true
	})
	return out
}

// returnedIdents は return 文でそのまま返している識別子名の集合を返す。
func returnedIdents(fn *ast.FuncDecl) map[string]bool {
	out := map[string]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, r := range ret.Results {
			if id, ok := r.(*ast.Ident); ok {
				out[id.Name] = true
			}
		}
		return true
	})
	return out
}

// hasIgnoreFile はファイル先頭コメントに //slicelint:ignore-file があるかを返す。
func hasIgnoreFile(f *ast.File) bool {
	for _, cg := range f.Comments {
		// package 宣言より後ろのコメントは対象外（先頭コメントのみ見る）。
		if cg.Pos() > f.Package {
			break
		}
		if commentContains(cg, "//slicelint:ignore-file") {
			return true
		}
	}
	return false
}

// commentContains はコメント群に needle を含む行があるかを返す。
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

// allowCommentLines は //slicelint:allow が書かれた行と、その次の行を集める。
// 行末コメント（同じ行）と、直前行に書いたコメントの両方を抑制対象にする。
func allowCommentLines(fset *token.FileSet, f *ast.File) map[int]bool {
	out := map[int]bool{}
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if !strings.Contains(c.Text, "//slicelint:allow") {
				continue
			}
			line := fset.Position(c.Pos()).Line
			out[line] = true
			out[line+1] = true
		}
	}
	return out
}
