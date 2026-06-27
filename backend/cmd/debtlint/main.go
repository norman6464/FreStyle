// Command debtlint は「期限付きで明示された技術的負債」を CI で見張る自作リンター。
// コード中の `TODO(debt: YYYY-MM-DD)` コメントを集め、(1) 期限切れ (2) 期限未設定 を検出する。
// クリーンコース「技術的負債」の『負債は自覚して・期限を切って借りる』を機械化したもの。
// archlint / naminglint / apispec-lint と同じ枠組み（go/ast のみ・外部依存なし）。
//
// 検出ルール:
//   - `// TODO(debt: 2026-09-01): ...` で 期限が過去 → 期限切れの負債（返済すべき）として違反
//   - `// TODO(debt): ...` のように 期限が無い → 期限未設定の負債として違反（負債には締切を）
//   - `// TODO(debt: 2099-01-01): ...` のように 期限が未来 → OK（自覚され・期限が切られた負債）
//
// 検査対象は internal/ と cmd/ 配下の .go（コメントのみを見るので文字列リテラルは対象外）。
//
// 使い方:
//
//	go run ./cmd/debtlint            # backend/ 直下で実行
//	go run ./cmd/debtlint <root>    # 別ディレクトリを指定
//
// 違反は `path:line: メッセージ` 形式で出力し exit 1。
//
// 抑制（正当に期限を延ばす / 例外扱いする場合）:
//   - 同じコメント行に `//debtlint:allow <理由>` を併記するとその 1 件を無視
//   - ファイル先頭コメントに `//debtlint:ignore-file <理由>` でファイル全体を無視
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// debtRe は `TODO(debt)` / `TODO(debt: 2026-09-01)` を捉える。日付は任意。
var debtRe = regexp.MustCompile(`TODO\(debt(?::?\s*(\d{4}-\d{2}-\d{2}))?`)

const dateLayout = "2006-01-02"

// violation は 1 件の負債コメント違反。
type violation struct {
	file string
	line int
	msg  string
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr, time.Now()))
}

// runCLI は CLI 本体。now を引数で受けることで「期限切れ判定」をテスト可能にする。
// exit code: 0=OK / 1=違反あり / 2=実行エラー。
func runCLI(args []string, stdout, stderr io.Writer, now time.Time) int {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var violations []violation
	for _, sub := range []string{"internal", "cmd"} {
		dir := filepath.Join(root, sub)
		if _, err := os.Stat(dir); err != nil {
			continue // そのディレクトリが無ければスキップ
		}
		vs, err := scanDir(dir, today)
		if err != nil {
			fmt.Fprintf(stderr, "debtlint: %v\n", err)
			return 2
		}
		violations = append(violations, vs...)
	}

	if len(violations) == 0 {
		fmt.Fprintln(stdout, "debtlint: OK — 期限切れ / 期限未設定の技術的負債コメントはありません")
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
	fmt.Fprintf(stderr, "\ndebtlint: %d 件の要返済 / 要期限設定の技術的負債が見つかりました\n", len(violations))
	return 1
}

// scanDir は dir 配下の .go を再帰的に走査し、負債コメント違反を集める。
func scanDir(dir string, today time.Time) ([]violation, error) {
	var violations []violation
	fset := token.NewFileSet()

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return fmt.Errorf("%s の parse に失敗: %w", path, perr)
		}
		violations = append(violations, scanFile(fset, f, today)...)
		return nil
	})
	return violations, err
}

// scanFile は 1 ファイルのコメントから負債違反を集める（純粋関数: テスト容易）。
func scanFile(fset *token.FileSet, f *ast.File, today time.Time) []violation {
	// ファイル全体抑制。CommentGroup.Text() はディレクティブ風コメントを落とすため、
	// 生の *ast.Comment.Text を直接見る。
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "debtlint:ignore-file") {
				return nil
			}
		}
	}

	var violations []violation
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			text := c.Text
			m := debtRe.FindStringSubmatch(text)
			if m == nil {
				continue
			}
			if strings.Contains(text, "debtlint:allow") {
				continue // 行内抑制
			}
			pos := fset.Position(c.Slash)
			file := relpath(pos.Filename)

			if m[1] == "" {
				violations = append(violations, violation{
					file: file, line: pos.Line,
					msg: "技術的負債 TODO に期限がありません。`TODO(debt: YYYY-MM-DD)` で締切を付けてください（クリーンコース『技術的負債』）",
				})
				continue
			}
			deadline, err := time.Parse(dateLayout, m[1])
			if err != nil {
				violations = append(violations, violation{
					file: file, line: pos.Line,
					msg: fmt.Sprintf("技術的負債 TODO の期限 %q が不正です（YYYY-MM-DD 形式で書いてください）", m[1]),
				})
				continue
			}
			if deadline.Before(today) {
				overdue := int(today.Sub(deadline).Hours() / 24)
				violations = append(violations, violation{
					file: file, line: pos.Line,
					msg: fmt.Sprintf("期限切れの技術的負債です（期限 %s・%d 日超過）。返済するか期限を更新してください", m[1], overdue),
				})
			}
		}
	}
	return violations
}

// relpath は出力用に絶対パスを相対パスへ寄せる（カレント基準。失敗時は元のまま）。
func relpath(p string) string {
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, p); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return p
}
