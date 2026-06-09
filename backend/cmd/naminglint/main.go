// Command naminglint は usecase 層の命名・構造規約（CLAUDE.md §2.3 / §3.2）を検証する。
// 「1 usecase = struct + コンストラクタ + Execute メソッド」を機械化し、規約から外れた
// usecase を CI で弾く。archlint / apispec-lint と同じ枠組み（go/ast のみ・外部依存なし）。
//
// 検査対象は internal/usecase 直下（repository サブパッケージは対象外）。
//
// ルール:
//   - 公開 struct `XxxUseCase` には コンストラクタ `NewXxxUseCase` が必要
//   - 公開 struct `XxxUseCase` には メソッド `Execute` が必要
//
// 使い方:
//
//	go run ./cmd/naminglint            # backend/ 直下で実行
//	go run ./cmd/naminglint <root>     # 別ディレクトリを指定
//
// 違反は `path:line: メッセージ` 形式で出力し exit 1。
//
// 抑制（集約 read usecase など Execute を持たない正当な例外用）:
//   - struct の doc / 行末コメントに `//naminglint:allow <理由>`
//   - ファイル先頭コメントに `//naminglint:ignore-file <理由>`
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

// ucStruct は 1 つの `XxxUseCase` struct 宣言。
type ucStruct struct {
	name       string
	file       string
	line       int
	suppressed bool // //naminglint:allow が付いていれば Execute 必須を免除
}

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
	usecaseRoot := filepath.Join(root, "internal", "usecase")

	violations, err := run(usecaseRoot)
	if err != nil {
		fmt.Fprintf(stderr, "naminglint: %v\n", err)
		return 2
	}

	if len(violations) == 0 {
		fmt.Fprintln(stdout, "naminglint: OK — usecase 命名規約違反なし")
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
	fmt.Fprintf(stderr, "\nnaminglint: %d 件の命名規約違反が見つかりました\n", len(violations))
	return 1
}

// run は usecaseRoot 直下の .go を解析し、UseCase 規約違反を集める。
func run(usecaseRoot string) ([]violation, error) {
	fset := token.NewFileSet()

	var structs []ucStruct
	constructors := map[string]bool{} // "NewXxxUseCase"
	executeRecv := map[string]bool{}  // Execute メソッドを持つレシーバ型名

	err := filepath.WalkDir(usecaseRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// repository サブパッケージ等、usecase 直下以外は対象外。
			if path != usecaseRoot {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return fmt.Errorf("%s: %w", path, perr)
		}
		if hasIgnoreFile(f) {
			return nil
		}
		s, ctors, execs := analyzeFile(fset, f, path)
		structs = append(structs, s...)
		for _, c := range ctors {
			constructors[c] = true
		}
		for _, e := range execs {
			executeRecv[e] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var out []violation
	for _, s := range structs {
		if !constructors["New"+s.name] {
			out = append(out, violation{
				s.file, s.line,
				fmt.Sprintf("%s にはコンストラクタ New%s がありません（CLAUDE.md §2.3）", s.name, s.name),
			})
		}
		if !s.suppressed && !executeRecv[s.name] {
			out = append(out, violation{
				s.file, s.line,
				fmt.Sprintf("%s には Execute メソッドがありません（CLAUDE.md §2.3。集約 usecase 等の例外は //naminglint:allow）", s.name),
			})
		}
	}
	return out, nil
}

// analyzeFile は 1 ファイルから UseCase struct / New コンストラクタ / Execute レシーバを抽出する。
func analyzeFile(fset *token.FileSet, f *ast.File, path string) (structs []ucStruct, constructors, executeRecv []string) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				// 公開 struct `XxxUseCase` のみを対象にする（非公開の補助型は除外）。
				if !ok || !ast.IsExported(ts.Name.Name) || !strings.HasSuffix(ts.Name.Name, "UseCase") {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); !ok {
					continue
				}
				structs = append(structs, ucStruct{
					name:       ts.Name.Name,
					file:       path,
					line:       fset.Position(ts.Pos()).Line,
					suppressed: commentContains(d.Doc, "naminglint:allow") || commentContains(ts.Doc, "naminglint:allow") || commentContains(ts.Comment, "naminglint:allow"),
				})
			}
		case *ast.FuncDecl:
			if d.Recv == nil {
				// 関数: New コンストラクタを集める。
				if strings.HasPrefix(d.Name.Name, "New") {
					constructors = append(constructors, d.Name.Name)
				}
				continue
			}
			// メソッド: Execute を持つレシーバ型名を集める。
			if d.Name.Name == "Execute" {
				if name := receiverTypeName(d.Recv); name != "" {
					executeRecv = append(executeRecv, name)
				}
			}
		}
	}
	return structs, constructors, executeRecv
}

// receiverTypeName はレシーバ（*T または T）から型名 T を取り出す。
func receiverTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	switch e := recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if id, ok := e.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return e.Name
	}
	return ""
}

// hasIgnoreFile はファイル先頭（package 宣言より前）に //naminglint:ignore-file があるかを返す。
func hasIgnoreFile(f *ast.File) bool {
	for _, cg := range f.Comments {
		if cg.End() > f.Package {
			break
		}
		if commentContains(cg, "naminglint:ignore-file") {
			return true
		}
	}
	return false
}

// commentContains はコメント群の生テキストに needle が含まれるかを返す。
// //naminglint:xxx は go ディレクティブ扱いで Text() に出ないため List を直接走査する。
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
