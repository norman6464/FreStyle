// Command fakegen は internal/usecase/repository の repository interface（port）から、
// テスト用の fake 実装を自動生成する。手書きの fake（fakeLessonProgressRepo 等）を量産する
// 手間をなくし、interface に メソッドが増えても 再生成 するだけで fake が追従するようにする。
//
// 生成物: internal/usecase/repository/repofakes/fakes_gen.go（package repofakes）。
// 各 interface `XxxRepository` / `XxxPresigner` 等に対し、次の形の fake を生成する:
//
//	type FakeXxx struct {
//	    MethodFunc  func(...) (...)   // テストで差し込む挙動
//	    MethodCalls atomic.Int64      // 呼び出し回数（f.MethodCalls.Load() で読む）
//	}
//	var _ repository.Xxx = (*FakeXxx)(nil)   // interface を満たすことを compile-time に保証
//	func (f *FakeXxx) Method(p0 ...) (...) { f.MethodCalls.Add(1); if f.MethodFunc != nil { ... }; return zero }
//
// MethodFunc が nil のときはゼロ値を返す（既存の手書き fake と同じ「未設定メソッドは nil/0 を返す」流儀）。
//
// 呼び出し回数を atomic.Int64 にしているのは、usecase が goroutine から repository を呼ぶ
// ことがあり（SendAiMessageStreamUseCase 等）、CI が go test -race で走るため。素の int だと
// 並行に呼ばれた時点で data race になる。
//
// 使い方:
//
//	go run ./cmd/fakegen            # backend/ 直下で実行（カレントを module root とみなす）
//	go run ./cmd/fakegen <root>    # 別ディレクトリを指定
//
// go generate からも呼べる（internal/usecase/repository/repofakes/doc.go に //go:generate を置く）。
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	repoPkgName = "repository"
	genPkgName  = "repofakes"
	outRelPath  = "internal/usecase/repository/repofakes/fakes_gen.go"
	repoRelPath = "internal/usecase/repository"
	atomicPath  = "sync/atomic"
)

// predeclared は Go の組み込み型（local パッケージ修飾を付けない）。
var predeclared = map[string]bool{
	"bool": true, "string": true, "error": true, "any": true, "comparable": true,
	"byte": true, "rune": true, "uintptr": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"float32": true, "float64": true, "complex64": true, "complex128": true,
}

// method は 1 つの interface メソッド。
type method struct {
	name    string
	params  []string // 型のみ（func フィールド用）。例: ["context.Context", "uint64"]
	results []string // 型のみ。例: ["bool", "error"]
}

// iface は 1 つの repository interface。
type iface struct {
	name    string // "LessonProgressRepository"
	methods []method
}

// genCtx は解析中に集まる副次情報を持つ。
//
// unsupported は「解析できずに近似した箇所」を貯める。近似は生成物を壊すが、黙って成功すると
// 原因が fakegen だと分からないコンパイルエラーとして後から出てくるため、run が stderr に列挙する。
type genCtx struct {
	usedAliases map[string]bool
	unsupported []string
}

func newGenCtx() *genCtx {
	return &genCtx{usedAliases: map[string]bool{}}
}

func (c *genCtx) markUnsupported(msg string) {
	c.unsupported = append(c.unsupported, msg)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run は CLI 本体。0=成功 / 1=エラー を返す（os.Exit を呼ばずテスト可能にする）。
func run(args []string, stdout, stderr io.Writer) int {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	modulePath, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		fmt.Fprintf(stderr, "fakegen: go.mod を読めません: %v\n", err)
		return 1
	}

	repoDir := filepath.Join(root, filepath.FromSlash(repoRelPath))
	ifaces, aliasToPath, ctx, err := collect(repoDir)
	if err != nil {
		fmt.Fprintf(stderr, "fakegen: %v\n", err)
		return 1
	}
	if len(ifaces) == 0 {
		fmt.Fprintln(stdout, "fakegen: 生成対象の interface が見つかりませんでした")
		return 0
	}

	// 近似した箇所は必ず表に出す。実際に壊れていれば生成物の compile-time アサーション
	// （var _ repository.X = (*FakeX)(nil)）がビルドを止めるが、その時に原因を辿れるようにする。
	for _, u := range ctx.unsupported {
		fmt.Fprintf(stderr, "fakegen: 警告: %s\n", u)
	}

	src := generate(modulePath, ifaces, aliasToPath, ctx.usedAliases)
	formatted, err := format.Source([]byte(src))
	formatFailed := err != nil
	if formatFailed {
		// 整形に失敗する＝生成コードが構文的に壊れている。原因を追えるよう生のソースは書き出すが、
		// 成功扱いにするとコンパイルできない fakes_gen.go がそのまま commit されるため 1 を返す。
		fmt.Fprintf(stderr, "fakegen: 生成コードの整形に失敗: %v\n", err)
		formatted = []byte(src)
	}

	// 生成物はリポジトリに commit する普通のソースなので、他のソースと同じ 0644 / 0755 で置く
	// （gosec の G301/G306 は「秘匿ファイル向け」の指摘なのでここでは当たらない。.golangci.yml で除外）。
	outPath := filepath.Join(root, filepath.FromSlash(outRelPath))
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fmt.Fprintf(stderr, "fakegen: 出力ディレクトリ作成に失敗: %v\n", err)
		return 1
	}
	if err := os.WriteFile(outPath, formatted, 0o644); err != nil {
		fmt.Fprintf(stderr, "fakegen: 書き込みに失敗: %v\n", err)
		return 1
	}
	if formatFailed {
		return 1
	}

	total := 0
	for _, i := range ifaces {
		total += len(i.methods)
	}
	fmt.Fprintf(stdout, "fakegen: OK — %d interface / %d メソッドの fake を生成しました → %s\n",
		len(ifaces), total, outRelPath)
	return 0
}

// collect は repository パッケージを解析し、interface 一覧・import エイリアス→path・解析の副次情報を返す。
func collect(repoDir string) (ifaces []iface, aliasToPath map[string]string, ctx *genCtx, err error) {
	entries, err := os.ReadDir(repoDir)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s を読めません: %w", repoRelPath, err)
	}

	aliasToPath = map[string]string{}
	ctx = newGenCtx()
	fset := token.NewFileSet()

	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, filepath.Join(repoDir, name))
	}
	sort.Strings(files)

	for _, path := range files {
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return nil, nil, nil, fmt.Errorf("%s の parse に失敗: %w", path, perr)
		}
		// import エイリアスを集める（alias 省略時はパス末尾を使う簡易解決）。
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			alias := defaultPkgName(p)
			if imp.Name != nil {
				alias = imp.Name.Name
			}
			if alias != "_" && alias != "." {
				aliasToPath[alias] = p
			}
		}
		// interface 型宣言を集める。
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ast.IsExported(ts.Name.Name) {
					continue
				}
				it, ok := ts.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				iv := iface{name: ts.Name.Name}
				for _, m := range it.Methods.List {
					ft, ok := m.Type.(*ast.FuncType)
					if !ok || len(m.Names) == 0 {
						// 埋め込み interface はメソッド集合を展開できないためスキップする
						// （repository port は現状 flat）。黙って落とすと fake が interface を
						// 満たさなくなるので記録して警告する。
						ctx.markUnsupported(fmt.Sprintf(
							"%s: 埋め込み interface %s を展開できないためスキップしました",
							ts.Name.Name, types.ExprString(m.Type),
						))
						continue
					}
					mt := method{name: m.Names[0].Name}
					mt.params = expandFieldTypes(ft.Params, ctx)
					if ft.Results != nil {
						mt.results = expandFieldTypes(ft.Results, ctx)
					}
					iv.methods = append(iv.methods, mt)
				}
				ifaces = append(ifaces, iv)
			}
		}
	}
	return ifaces, aliasToPath, ctx, nil
}

// expandFieldTypes は FieldList を「個々の型文字列」に展開する（グループ化引数を 1 つずつに分解）。
func expandFieldTypes(fl *ast.FieldList, ctx *genCtx) []string {
	if fl == nil {
		return nil
	}
	var out []string
	for _, field := range fl.List {
		t := renderType(field.Type, ctx)
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			out = append(out, t)
		}
	}
	return out
}

// renderType は型 AST を、repofakes パッケージから見た型文字列に変換する。
// repository パッケージの局所型（無修飾の exported 名）は repository. を付ける。
func renderType(expr ast.Expr, ctx *genCtx) string {
	switch t := expr.(type) {
	case *ast.Ident:
		if predeclared[t.Name] {
			return t.Name
		}
		if ast.IsExported(t.Name) {
			ctx.usedAliases[repoPkgName] = true
			return repoPkgName + "." + t.Name // repository パッケージの局所型
		}
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			ctx.usedAliases[x.Name] = true
			return x.Name + "." + t.Sel.Name
		}
		return renderType(t.X, ctx) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + renderType(t.X, ctx)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + renderType(t.Elt, ctx)
		}
		return "[" + renderType(t.Len, ctx) + "]" + renderType(t.Elt, ctx)
	case *ast.Ellipsis:
		return "..." + renderType(t.Elt, ctx)
	case *ast.MapType:
		return "map[" + renderType(t.Key, ctx) + "]" + renderType(t.Value, ctx)
	case *ast.InterfaceType:
		if t.Methods == nil || len(t.Methods.List) == 0 {
			return "any"
		}
		ctx.markUnsupported("メソッドを持つ無名 interface " + types.ExprString(expr) + " を any に置き換えました")
		return "any"
	case *ast.BasicLit:
		return t.Value
	default:
		ctx.markUnsupported("未対応の型 " + types.ExprString(expr) + " を any に置き換えました")
		return "any" // 想定外（func/chan/struct 型の引数等）。repository port では出現しない。
	}
}

// generate は fake 実装のソース文字列を組み立てる。
func generate(modulePath string, ifaces []iface, aliasToPath map[string]string, usedAliases map[string]bool) string {
	var b strings.Builder
	b.WriteString("// Code generated by fakegen. DO NOT EDIT.\n")
	b.WriteString("//\n")
	b.WriteString("// repository interface(port) のテスト用 fake。re-generate: make fakegen\n")
	b.WriteString("package " + genPkgName + "\n\n")

	// import を組み立てる（usedAliases から）。repository と sync/atomic は常に必要。
	imports := map[string]string{} // alias -> path
	imports[repoPkgName] = modulePath + "/" + repoRelPath
	imports["atomic"] = atomicPath
	for alias := range usedAliases {
		if alias == repoPkgName {
			continue
		}
		if p, ok := aliasToPath[alias]; ok {
			imports[alias] = p
		}
	}
	b.WriteString(renderImports(imports))

	for _, iv := range ifaces {
		fakeName := "Fake" + iv.name
		fmt.Fprintf(&b, "// %s は %s.%s の生成済み fake。\n", fakeName, repoPkgName, iv.name)
		b.WriteString("type " + fakeName + " struct {\n")
		for _, m := range iv.methods {
			fmt.Fprintf(&b, "\t%sFunc func(%s)%s\n", m.name, strings.Join(m.params, ", "), resultSig(m.results))
			fmt.Fprintf(&b, "\t%sCalls atomic.Int64\n", m.name)
		}
		b.WriteString("}\n\n")

		// compile-time に interface を満たすことを保証。
		fmt.Fprintf(&b, "var _ %s.%s = (*%s)(nil)\n\n", repoPkgName, iv.name, fakeName)

		for _, m := range iv.methods {
			writeMethod(&b, fakeName, m)
		}
	}
	return b.String()
}

// renderImports は import 宣言を「標準ライブラリ / それ以外」の 2 グループで書き出す。
//
// 生成ファイルは "Code generated ... DO NOT EDIT." ヘッダを持つため gofumpt / gofmt が
// 整形対象から外す。グループ分けを外部ツールに任せられないので、ここで gofmt 準拠の形にする。
func renderImports(imports map[string]string) string {
	type spec struct{ alias, path string }
	var std, ext []spec
	for alias, path := range imports {
		s := spec{alias: alias, path: path}
		if isStdlib(path) {
			std = append(std, s)
		} else {
			ext = append(ext, s)
		}
	}
	byPath := func(s []spec) { sort.Slice(s, func(i, j int) bool { return s[i].path < s[j].path }) }
	byPath(std)
	byPath(ext)

	var b strings.Builder
	b.WriteString("import (\n")
	writeGroup := func(g []spec) {
		for _, s := range g {
			if s.alias == defaultPkgName(s.path) {
				b.WriteString("\t\"" + s.path + "\"\n")
			} else {
				b.WriteString("\t" + s.alias + " \"" + s.path + "\"\n")
			}
		}
	}
	writeGroup(std)
	if len(std) > 0 && len(ext) > 0 {
		b.WriteString("\n")
	}
	writeGroup(ext)
	b.WriteString(")\n\n")
	return b.String()
}

// isStdlib は import path が標準ライブラリかを判定する。
// 標準ライブラリ以外はホスト名から始まるため、先頭要素にドットが含まれるかで見分けられる。
func isStdlib(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
}

// writeMethod は 1 メソッドの実装を書き出す。
func writeMethod(b *strings.Builder, fakeName string, m method) {
	// 引数名 p0,p1,... を付ける。可変長は最後の型が ... で始まる。
	paramDecls := make([]string, len(m.params))
	callArgs := make([]string, len(m.params))
	for i, t := range m.params {
		name := fmt.Sprintf("p%d", i)
		paramDecls[i] = name + " " + t
		if strings.HasPrefix(t, "...") {
			callArgs[i] = name + "..."
		} else {
			callArgs[i] = name
		}
	}

	fmt.Fprintf(b, "func (f *%s) %s(%s)%s {\n",
		fakeName, m.name, strings.Join(paramDecls, ", "), resultSig(m.results))
	fmt.Fprintf(b, "\tf.%sCalls.Add(1)\n", m.name)
	fmt.Fprintf(b, "\tif f.%sFunc != nil {\n", m.name)
	if len(m.results) == 0 {
		fmt.Fprintf(b, "\t\tf.%sFunc(%s)\n\t\treturn\n", m.name, strings.Join(callArgs, ", "))
		b.WriteString("\t}\n")
		b.WriteString("}\n\n")
		return
	}
	fmt.Fprintf(b, "\t\treturn f.%sFunc(%s)\n", m.name, strings.Join(callArgs, ", "))
	b.WriteString("\t}\n")
	// ゼロ値を返す。
	retNames := make([]string, len(m.results))
	for i, t := range m.results {
		rn := fmt.Sprintf("r%d", i)
		retNames[i] = rn
		fmt.Fprintf(b, "\tvar %s %s\n", rn, t)
	}
	b.WriteString("\treturn " + strings.Join(retNames, ", ") + "\n")
	b.WriteString("}\n\n")
}

// resultSig は戻り値の型リストを Go の関数シグネチャ表記にする（0=空 / 1=" T" / 複数=" (T, U)"）。
func resultSig(results []string) string {
	switch len(results) {
	case 0:
		return ""
	case 1:
		return " " + results[0]
	default:
		return " (" + strings.Join(results, ", ") + ")"
	}
}

// defaultPkgName は import path からデフォルトのパッケージ名を返す。
//
// 単純な末尾要素ではない。Go Modules のメジャーバージョン接尾辞（/v2 以降）はパッケージ名では
// ないため、"github.com/foo/bar/v2" のパッケージ名は "bar"。末尾をそのまま使うと alias 解決が
// ずれて、必要な import が生成されずコンパイルできない fake ができる。
func defaultPkgName(path string) string {
	parts := strings.Split(path, "/")
	last := parts[len(parts)-1]
	if len(parts) >= 2 && isMajorVersionSuffix(last) {
		return parts[len(parts)-2]
	}
	return last
}

// isMajorVersionSuffix は Go Modules のメジャーバージョン要素（v2 以降）かを判定する。
// v0 / v1 は import path の接尾辞として現れないため対象外。
func isMajorVersionSuffix(s string) bool {
	if len(s) < 2 || s[0] != 'v' {
		return false
	}
	for _, r := range s[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != "v0" && s != "v1"
}

// readModulePath は go.mod の `module X` 行から module path を取り出す。
func readModulePath(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range bytes.Split(data, []byte("\n")) {
		s := strings.TrimSpace(string(line))
		if !strings.HasPrefix(s, "module ") {
			continue
		}
		p := strings.TrimSpace(strings.TrimPrefix(s, "module "))
		// `module example.com/x // 説明` の行末コメントを落とす。付いたままだと
		// 生成する import path が壊れる。
		if i := strings.Index(p, "//"); i >= 0 {
			p = strings.TrimSpace(p[:i])
		}
		if p == "" {
			return "", fmt.Errorf("module 行にパスがありません")
		}
		return p, nil
	}
	return "", fmt.Errorf("module 行が見つかりません")
}
