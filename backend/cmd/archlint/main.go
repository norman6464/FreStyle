// Command archlint は FreStyle のクリーンアーキテクチャ依存方向ルール（CLAUDE.md §2）を
// 静的に検証する。go/parser で各ファイルの import を解析し、層をまたぐ禁止依存を検出する。
//
// 依存方向: handler → usecase → repository(port) / infra → domain
// （実装は adapter/persistence、port は usecase/repository に分離）
//
// 使い方:
//
//	go run ./cmd/archlint           # backend/ 直下で実行（カレントを module root とみなす）
//	go run ./cmd/archlint <root>    # 別ディレクトリを指定
//
// 違反があれば `path:line: メッセージ` 形式で出力し exit code 1 を返す。CI の品質ゲートに組み込む。
//
// 抑制:
//   - import 行末に `//archlint:allow` を付けるとその 1 行を無視する
//   - ファイル先頭コメントに `//archlint:ignore-file` を含めるとそのファイル全体を無視する
package main

import (
	"bufio"
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

// 層・依存先の分類キー。import path / ファイルパスの両方をこの値に正規化して照合する。
const (
	layerDomain      = "domain"
	layerUsecase     = "usecase"
	layerUsecaseRepo = "usecase/repository"
	layerHandler     = "handler"
	layerPersistence = "adapter/persistence"
	layerInfra       = "infra"

	targetGin     = "gin"
	targetNetHTTP = "net/http"
)

// rules は「ソース層 → 禁止する依存先 → 違反メッセージ」。ここに無い組み合わせは許容。
// handler→persistence だけは wiring ファイル（router.go / routes_*.go）で例外扱いする（§2.6）。
var rules = map[string]map[string]string{
	layerDomain: {
		layerHandler:     "domain は handler を import できません（domain は他層に依存しない）",
		layerUsecase:     "domain は usecase を import できません（domain は他層に依存しない）",
		layerUsecaseRepo: "domain は usecase/repository を import できません（domain は他層に依存しない）",
		layerPersistence: "domain は adapter/persistence を import できません（domain は他層に依存しない）",
		layerInfra:       "domain は infra を import できません（domain は他層に依存しない）",
		targetGin:        "domain は gin を import できません（domain は標準ライブラリ + GORM tag のみ）",
		targetNetHTTP:    "domain は net/http を import できません（domain は標準ライブラリ + GORM tag のみ）",
	},
	layerUsecaseRepo: {
		layerHandler:     "usecase/repository(port) は handler を import できません",
		layerUsecase:     "usecase/repository(port) は usecase 本体を import できません（port は domain のみに依存）",
		layerPersistence: "usecase/repository(port) は adapter/persistence を import できません（DIP 違反）",
		layerInfra:       "usecase/repository(port) は infra を import できません",
		targetGin:        "usecase/repository(port) は gin を import できません",
		targetNetHTTP:    "usecase/repository(port) は net/http を import できません",
	},
	layerUsecase: {
		layerHandler:     "usecase は handler を import できません（依存方向違反）",
		layerPersistence: "usecase は adapter/persistence を import できません（repository interface に依存すること: DIP）",
		targetGin:        "usecase は gin を import できません（*gin.Context など HTTP 層の型を参照しない）",
		targetNetHTTP:    "usecase は net/http を import できません（HTTP 層の型を参照しない）",
	},
	layerPersistence: {
		layerHandler: "adapter/persistence は handler を import できません",
		layerUsecase: "adapter/persistence は usecase 本体を import できません（依存先は port = usecase/repository だけ）",
		targetGin:    "adapter/persistence は gin を import できません",
	},
	layerInfra: {
		layerHandler:     "infra は handler を import できません（infra は usecase / handler を参照しない）",
		layerUsecase:     "infra は usecase を import できません（infra は usecase / handler を参照しない）",
		layerUsecaseRepo: "infra は usecase/repository を import できません",
		layerPersistence: "infra は adapter/persistence を import できません",
		targetGin:        "infra は gin を import できません",
	},
	layerHandler: {
		layerPersistence: "handler は adapter/persistence を直接 import できません（usecase 経由にする。wiring は router.go / routes_*.go に限定）",
	},
}

// violation は 1 件の依存方向違反。
type violation struct {
	file string
	line int
	imp  string
	msg  string
}

// importRef は解析済みの 1 つの import。
type importRef struct {
	path       string
	line       int
	target     string // 分類済みの依存先キー（rules の照合に使う）
	suppressed bool   // //archlint:allow が付いていれば true
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

// runCLI は CLI 本体。exit code を返す（0=OK / 1=違反あり / 2=実行エラー）。
// os.Exit を呼ばないことで end-to-end テストを可能にする。
func runCLI(args []string, stdout, stderr io.Writer) int {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	modulePath, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		fmt.Fprintf(stderr, "archlint: go.mod を読めません: %v\n", err)
		return 2
	}
	internalPrefix := modulePath + "/internal/"
	internalRoot := filepath.Join(root, "internal")

	violations, err := run(internalRoot, internalPrefix)
	if err != nil {
		fmt.Fprintf(stderr, "archlint: %v\n", err)
		return 2
	}

	if len(violations) == 0 {
		fmt.Fprintln(stdout, "archlint: OK — クリーンアーキテクチャ依存方向ルール違反なし")
		return 0
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].file != violations[j].file {
			return violations[i].file < violations[j].file
		}
		return violations[i].line < violations[j].line
	})
	for _, v := range violations {
		fmt.Fprintf(stdout, "%s:%d: %s（import %q）\n", v.file, v.line, v.msg, v.imp)
	}
	fmt.Fprintf(stderr, "\narchlint: %d 件の依存方向違反が見つかりました\n", len(violations))
	return 1
}

// run は internalRoot 配下の本番 .go を走査し、依存方向違反を集める。
func run(internalRoot, internalPrefix string) ([]violation, error) {
	var out []violation
	err := filepath.WalkDir(internalRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relDir, err := filepath.Rel(internalRoot, filepath.Dir(path))
		if err != nil {
			return err
		}
		src := classifyRel(filepath.ToSlash(relDir))
		if _, ok := rules[src]; !ok {
			return nil // ルール対象外の層（cmd 等）はスキップ
		}

		imports, ignoreFile, err := parseImports(path, internalPrefix)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if ignoreFile {
			return nil
		}
		out = append(out, violationsFor(src, filepath.Base(path), path, imports)...)
		return nil
	})
	return out, err
}

// violationsFor は 1 ファイル分の import 列からルール違反を抽出する（純粋関数: テスト容易）。
func violationsFor(src, filename, fullpath string, imports []importRef) []violation {
	forbidden := rules[src]
	wiring := isWiringFile(filename)
	var vs []violation
	for _, imp := range imports {
		if imp.suppressed || imp.target == "" {
			continue
		}
		// handler→persistence は wiring ファイルでのみ許容。
		if src == layerHandler && imp.target == layerPersistence && wiring {
			continue
		}
		if msg, bad := forbidden[imp.target]; bad {
			vs = append(vs, violation{file: fullpath, line: imp.line, imp: imp.path, msg: msg})
		}
	}
	return vs
}

// parseImports は import 文だけを解析し、各 import を分類して返す。
// ファイル先頭コメントに archlint:ignore-file があれば ignoreFile=true。
func parseImports(path, internalPrefix string) (imports []importRef, ignoreFile bool, err error) {
	fset := token.NewFileSet()
	// ImportsOnly は import の trailing コメント関連付けを欠落させるため使わない（抑制判定に必要）。
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	for _, cg := range f.Comments {
		if commentContains(cg, "archlint:ignore-file") {
			return nil, true, nil
		}
	}

	for _, spec := range f.Imports {
		p := strings.Trim(spec.Path.Value, `"`)
		ref := importRef{
			path:   p,
			line:   fset.Position(spec.Path.Pos()).Line,
			target: classifyImport(p, internalPrefix),
		}
		if commentHasAllow(spec.Comment) || commentHasAllow(spec.Doc) {
			ref.suppressed = true
		}
		imports = append(imports, ref)
	}
	return imports, false, nil
}

// commentHasAllow は import に付いた抑制コメント //archlint:allow の有無を返す。
func commentHasAllow(cg *ast.CommentGroup) bool {
	return commentContains(cg, "archlint:allow")
}

// commentContains はコメント群の「生テキスト」に needle が含まれるかを返す。
// CommentGroup.Text() は //archlint:xxx を go ディレクティブ扱いで除去するため、
// 各 Comment.Text（// を含む生文字列）を直接走査する。
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

// classifyImport は import path を rules 照合用キーに正規化する。対象外は "" を返す。
func classifyImport(importPath, internalPrefix string) string {
	if importPath == targetNetHTTP {
		return targetNetHTTP
	}
	if importPath == "github.com/gin-gonic/gin" || strings.HasPrefix(importPath, "github.com/gin-gonic/gin/") {
		return targetGin
	}
	if strings.HasPrefix(importPath, internalPrefix) {
		return classifyRel(strings.TrimPrefix(importPath, internalPrefix))
	}
	return ""
}

// classifyRel は internal/ からの相対パス（import / ファイル共通）を層キーに分類する。
func classifyRel(rel string) string {
	switch {
	case rel == layerDomain || strings.HasPrefix(rel, layerDomain+"/"):
		return layerDomain
	case rel == layerUsecaseRepo || strings.HasPrefix(rel, layerUsecaseRepo+"/"):
		return layerUsecaseRepo
	case rel == layerUsecase || strings.HasPrefix(rel, layerUsecase+"/"):
		return layerUsecase
	case rel == layerHandler || strings.HasPrefix(rel, layerHandler+"/"):
		return layerHandler
	case rel == layerPersistence || strings.HasPrefix(rel, layerPersistence+"/"):
		return layerPersistence
	case rel == "adapter" || strings.HasPrefix(rel, "adapter/"):
		return layerPersistence
	case rel == layerInfra || strings.HasPrefix(rel, layerInfra+"/"):
		return layerInfra
	default:
		return ""
	}
}

// isWiringFile は handler の組み立て担当ファイル（router.go / routes_*.go）かを判定する。
func isWiringFile(filename string) bool {
	return filename == "router.go" ||
		(strings.HasPrefix(filename, "routes_") && strings.HasSuffix(filename, ".go"))
}

// readModulePath は go.mod の先頭 module 行から module path を取り出す。
func readModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module 行が見つかりません")
}
