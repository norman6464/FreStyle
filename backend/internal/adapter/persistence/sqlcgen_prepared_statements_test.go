package persistence

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// sqlcgenDir は sqlc の生成先（このテストのパッケージから見た相対パス）。
// go test はパッケージのディレクトリを作業ディレクトリにするので、ここは常に
// internal/adapter/persistence/sqlcgen を指す。
const sqlcgenDir = "sqlcgen"

// Test_sqlc生成物が準備済みステートメントを使っていないこと は、生成コードが
// DBTX.PrepareContext を「呼んで」いないことを構文解析で確かめる。
//
// なぜこれを機械で守るのか:
//
//	sqlc.yaml で emit_prepared_queries: true にすると、sqlc は起動時に全クエリを
//	PrepareContext して *sql.Stmt を Queries に抱え込み、以後その Stmt を使い回す
//	コードを生成する。ローカルの直結 PostgreSQL では速くて正しい。
//	しかし本番の接続は Supabase の transaction pooler（pgbouncer）越しで、pooler は
//	文と文のあいだで裏の PostgreSQL 接続を別クライアントへ回す。準備したときの接続と
//	実行するときの接続が同じである保証がないので、実行時に
//	"prepared statement does not exist"（SQLSTATE 26000）で落ちる。
//	infra/database/postgres.go が pooler の DSN を見て simple query protocol を
//	強制しているのも同じ理由である。
//
//	たちが悪いのは、この壊れ方がローカルでも CI でも再現しないこと。どちらも
//	PostgreSQL への直結で、接続が生きているあいだ prepared statement は残るため、
//	テストは全部緑のまま通る。壊れるのはデプロイして pooler 越しに叩いた瞬間だけになる。
//
// なぜ「設定ファイルの文字列」ではなく「生成結果」を見るのか:
//
//	sqlc.yaml に emit_prepared_queries が書かれていないことを検査しても、設定の
//	書き方（別ファイルへの分割・将来のキー名変更・上書き）が変わると素通りする。
//	見るべきなのは結果、すなわち生成コードが実際に Prepare しているかどうか。
//
// 宣言と呼び出しを取り違えないこと:
//
//	DBTX interface の PrepareContext(context.Context, string) (*sql.Stmt, error) は
//	emit_prepared_queries の値に関係なく sqlc が必ず「宣言」する。宣言を違反として
//	数えるとこの検査は常に落ち、逆に呼び出しではなく宣言の有無だけを見ると常に通る。
//	ここでは interface の宣言は違反にせず、CallExpr（実際の呼び出し）だけを数える。
//	そのうえで「宣言が確かに存在すること」も確認し、検査が正しい生成物を見ていることを
//	担保する。
func Test_sqlc生成物が準備済みステートメントを使っていないこと(t *testing.T) {
	fset := token.NewFileSet()

	entries, err := os.ReadDir(sqlcgenDir)
	require.NoError(t, err, "sqlc の生成先 %s を読めません。生成先を移したならこのテストの sqlcgenDir も直してください", sqlcgenDir)

	var (
		scanned          []string
		prepareCalls     []string
		declaredInDBTX   bool
		declarationFiles []string
	)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(sqlcgenDir, e.Name())

		f, perr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		require.NoError(t, perr, "%s を構文解析できません", path)
		scanned = append(scanned, path)

		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.InterfaceType:
				// interface のメソッド「宣言」。DBTX は設定に関わらず PrepareContext を
				// 宣言するので、ここに現れても違反ではない。
				if node.Methods == nil {
					return true
				}
				for _, m := range node.Methods.List {
					for _, name := range m.Names {
						if name.Name == "PrepareContext" {
							declaredInDBTX = true
							declarationFiles = append(declarationFiles, fset.Position(name.Pos()).String())
						}
					}
				}
			case *ast.CallExpr:
				// 実際の「呼び出し」。db.PrepareContext(...) / q.db.PrepareContext(...) など
				// セレクタ経由の呼び出しだけが prepared statement を作る。
				sel, ok := node.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if sel.Sel.Name == "PrepareContext" || sel.Sel.Name == "Prepare" {
					prepareCalls = append(prepareCalls, fmt.Sprintf("%s: %s(...)", fset.Position(node.Pos()), sel.Sel.Name))
				}
			}
			return true
		})
	}

	// 生成先を移した・空になった等で 1 ファイルも読めていないと、この検査は
	// 何も見ないまま緑になる。空振りを違反として扱う。
	require.NotEmpty(t, scanned, "%s に .go が 1 つもありません（`make sqlc` を実行してください）", sqlcgenDir)

	// 宣言が消えていたら、生成物の形が変わったか別の場所を見ている。どちらでも
	// この検査の前提が崩れているので、黙って通さず気づけるようにする。
	require.True(t, declaredInDBTX,
		"DBTX interface に PrepareContext の宣言が見つかりません。sqlc の生成物の形が変わった可能性があります"+
			"（読んだファイル: %s）", strings.Join(scanned, ", "))

	if len(prepareCalls) > 0 {
		require.Failf(t, "sqlc の生成物が準備済みステートメントを使っています",
			"PrepareContext の呼び出しが %d 箇所あります:\n%s\n\n"+
				"emit_prepared_queries: true にすると sqlc は *sql.Stmt を抱え込む Prepare() を生成します。"+
				"本番は Supabase transaction pooler（pgbouncer）越しで prepared statement が"+
				"セッションをまたいで残らないため、実行時に \"prepared statement does not exist\" で落ちます。"+
				"ローカル直結と CI（直結）では再現しないので、テストが緑のままデプロイ先だけが壊れます。"+
				"sqlc.yaml の emit_prepared_queries を外して `make sqlc` をやり直してください。\n"+
				"（宣言側の DBTX.PrepareContext は違反ではありません: %s）",
			len(prepareCalls), formatSample(prepareCalls, 5), strings.Join(declarationFiles, ", "))
	}
}

// formatSample は違反一覧の先頭 limit 件だけを整形する。
// 生成物は 1 クエリにつき 1 回 Prepare するので、違反時は数百件並びうる。
// 全部出すと失敗の理由がスクロールで流れてしまうため、件数は上の %d に任せて抜粋だけ見せる。
func formatSample(items []string, limit int) string {
	shown := items
	suffix := ""
	if len(items) > limit {
		shown = items[:limit]
		suffix = fmt.Sprintf("\n  ... 他 %d 件", len(items)-limit)
	}
	return "  " + strings.Join(shown, "\n  ") + suffix
}
