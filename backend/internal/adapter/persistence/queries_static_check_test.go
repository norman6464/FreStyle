package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// このファイルは queries/*.sql そのものを読んで、SQL の書き方の不変条件を検査する。
//
// なぜ sqlc vet（sqlc.yaml の rules）ではなくここにあるのか:
//
//	vet の CEL 式に渡ってくる query.sql は「書いたとおりの SQL」ではない。sqlc が処理した
//	あとの姿で、(a) スキーマは一切見えず、(b) -- name: 行から最初の SQL キーワードまでの
//	コメントは落ちており、(c) sqlc.arg / sqlc.slice は既に $1, $2 … へ書き換わっている
//	（いずれも sqlc v1.31.1 で実測。sqlc.yaml 冒頭に同じ記載がある）。
//	下の 2 つはそれぞれ「表の列を知る必要がある」「sqlc.slice という文字列を見る必要がある」
//	ため、vet では原理的に書けない。書けないものを無理に近似すると、正当なクエリを
//	弾くか、危ないクエリを見逃すかのどちらかになる。
//
// なぜテストなのか:
//
//	ここが見るのは persistence パッケージ 1 つに閉じた SQL ファイルなので、パッケージに
//	同居させたほうが場所として素直で、`go test ./...`（CI と make verify が既に通す）に
//	そのまま乗る。同じパッケージに sqlcgen_prepared_statements_test.go という先例もある。
//	テストなら、違反する SQL の見本を fixture として常設でき、「本当に弾けるのか」を
//	毎回機械が確かめてくれる（検査そのものが壊れたときに気づける）。

const (
	// queriesDir / schemaDir は、このテストのパッケージから見た相対パス。
	// go test はパッケージのディレクトリを作業ディレクトリにするので固定で良い。
	queriesDir = "queries"
	schemaDir  = "../../infra/database/schema"
)

// ownershipColumns は「その行が誰のものか / どのテナントのものか」を決める列名。
//
// ON CONFLICT の衝突キーにこれらが入っていないと、他人の行に衝突を当てられる（後述）。
// 列名で判断しているのは、スキーマから意味を機械的に読み取る方法が他に無いため。
// 新しく所有・所属を表す列を導入したらここに足すこと。
var ownershipColumns = map[string]string{
	"user_id":             "行の持ち主",
	"owner_id":            "行の持ち主",
	"created_by":          "行を作った人",
	"created_by_user_id":  "行を作った人",
	"workspace_id":        "所属テナント（ワークスペース）",
	"principal_id":        "権限が向けられている主体",
	"group_principal_id":  "権限が向けられている主体",
	"member_principal_id": "権限が向けられている主体",
}

var (
	queryNameRe   = regexp.MustCompile(`(?m)^--[ \t]*name:[ \t]*([A-Za-z0-9_]+)`)
	createTableRe = regexp.MustCompile(`(?is)create\s+table\s+(?:if\s+not\s+exists\s+)?"?([a-z_][a-z0-9_]*)"?\s*\(`)
	insertIntoRe  = regexp.MustCompile(`(?is)\binsert\s+into\s+"?([a-z_][a-z0-9_]*)"?`)
	onConflictRe  = regexp.MustCompile(`(?is)\bon\s+conflict\b`)
	doActionRe    = regexp.MustCompile(`(?is)\bdo\s+(update|nothing)\b`)
	whereRe       = regexp.MustCompile(`(?is)\bwhere\b`)
	identRe       = regexp.MustCompile(`[a-z_][a-z0-9_]*`)
	exemptionRe   = regexp.MustCompile(`(?im)^[ \t]*--[ \t]*upsert-owner-scope:[ \t]*(\S.*?)[ \t]*$`)
	sqlcSliceRe   = regexp.MustCompile(`(?i)sqlc\.slice[ \t]*\(`)
	columnDeclRe  = regexp.MustCompile(`^"?([a-z_][a-z0-9_]*)"?\s+\S`)
)

// namedQuery は queries/*.sql 内の `-- name: Xxx :cmd` から次の name までのひとかたまり。
// raw には先頭の name 行と、その下に書かれた説明コメントも含む。
type namedQuery struct {
	name string
	raw  string
	off  int
}

// upsertSite は 1 つの `INSERT ... ON CONFLICT ... DO UPDATE` の出現。
type upsertSite struct {
	file      string
	line      int
	queryName string
	table     string
	keyCols   []string
	// guard は DO UPDATE のあとに置かれた WHERE 以降の文字列。無ければ空。
	guard string
	// exemption は `-- upsert-owner-scope: <理由>` に書かれた理由。無ければ空。
	exemption string
}

// stripSQLComments は SQL のコメントを空白へ潰す。文字列リテラルの中は触らない。
// 長さと改行位置を保つので、返り値の byte offset は元の SQL の offset と一致する。
func stripSQLComments(s string) string {
	out := []byte(s)
	inString := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inString {
			if c == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				inString = false
			}
			continue
		}
		switch {
		case c == '\'':
			inString = true
		case c == '-' && i+1 < len(s) && s[i+1] == '-':
			for i < len(s) && s[i] != '\n' {
				out[i] = ' '
				i++
			}
			i--
		case c == '/' && i+1 < len(s) && s[i+1] == '*':
			end := len(s)
			if j := strings.Index(s[i+2:], "*/"); j >= 0 {
				end = i + 2 + j + 2
			}
			for ; i < end; i++ {
				if out[i] != '\n' {
					out[i] = ' '
				}
			}
			i--
		}
	}
	return string(out)
}

// parseSchemaTables は CREATE TABLE 群から「表名 → 列名の集合」を作る。
func parseSchemaTables(src string) map[string]map[string]bool {
	stripped := stripSQLComments(src)
	tables := map[string]map[string]bool{}

	for _, m := range createTableRe.FindAllStringSubmatchIndex(stripped, -1) {
		name := stripped[m[2]:m[3]]
		body, ok := balancedParen(stripped, m[1]-1)
		if !ok {
			continue
		}
		cols := map[string]bool{}
		for _, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			cm := columnDeclRe.FindStringSubmatch(line)
			if cm == nil {
				continue
			}
			switch strings.ToUpper(cm[1]) {
			case "PRIMARY", "UNIQUE", "CHECK", "FOREIGN", "CONSTRAINT", "EXCLUDE", "LIKE":
				continue
			}
			cols[cm[1]] = true
		}
		tables[name] = cols
	}
	return tables
}

// balancedParen は s[open] の '(' に対応する ')' までの中身を返す。
func balancedParen(s string, open int) (string, bool) {
	if open < 0 || open >= len(s) || s[open] != '(' {
		return "", false
	}
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return s[open+1 : i], true
			}
		}
	}
	return "", false
}

// splitNamedQueries は 1 ファイルを `-- name:` ごとのかたまりへ割る。
func splitNamedQueries(src string) []namedQuery {
	locs := queryNameRe.FindAllStringSubmatchIndex(src, -1)
	out := make([]namedQuery, 0, len(locs))
	for i, m := range locs {
		end := len(src)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		out = append(out, namedQuery{name: src[m[2]:m[3]], raw: src[m[0]:end], off: m[0]})
	}
	return out
}

func lineOf(src string, off int) int {
	return strings.Count(src[:off], "\n") + 1
}

// scanUpsertSites は 1 ファイルから ON CONFLICT ... DO UPDATE の出現を拾う。
// DO NOTHING は既存行を書き換えないので対象にしない。
func scanUpsertSites(file, src string) []upsertSite {
	var sites []upsertSite

	for _, q := range splitNamedQueries(src) {
		body := stripSQLComments(q.raw)
		exemption := ""
		if m := exemptionRe.FindStringSubmatch(q.raw); m != nil {
			exemption = m[1]
		}

		for _, oc := range onConflictRe.FindAllStringIndex(body, -1) {
			// 直前の INSERT INTO から対象の表を決める。
			table := ""
			for _, ins := range insertIntoRe.FindAllStringSubmatchIndex(body[:oc[0]], -1) {
				table = body[ins[2]:ins[3]]
			}

			// ON CONFLICT の直後に列リストがあれば衝突キー。
			// ON CONSTRAINT <名前> 形式ではキーが読めないので空のままにする（＝ 安全側）。
			p := oc[1]
			for p < len(body) && (body[p] == ' ' || body[p] == '\t' || body[p] == '\n' || body[p] == '\r') {
				p++
			}
			var keyCols []string
			if p < len(body) && body[p] == '(' {
				if inner, ok := balancedParen(body, p); ok {
					keyCols = identRe.FindAllString(strings.ToLower(inner), -1)
					p += len(inner) + 2
				}
			}

			// DO UPDATE / DO NOTHING のどちらか。
			act := doActionRe.FindStringSubmatchIndex(body[p:])
			if act == nil || strings.EqualFold(body[p+act[2]:p+act[3]], "nothing") {
				continue
			}

			guard := ""
			after := body[p+act[1]:]
			if w := whereRe.FindStringIndex(after); w != nil {
				guard = after[w[0]:]
			}

			sites = append(sites, upsertSite{
				file:      file,
				line:      lineOf(src, q.off+oc[0]),
				queryName: q.name,
				table:     table,
				keyCols:   keyCols,
				guard:     guard,
				exemption: exemption,
			})
		}
	}
	return sites
}

// checkUpsertSites は各 upsert を検査し、違反メッセージを返す。
//
// 危ないのは「その表が所有者列を持ち、かつその列が衝突キーに入っていない」とき。
// 攻撃者は自分が渡せる衝突キーの値だけで、他人が持っている行に衝突を当てられる。
// 当たった行は DO UPDATE で書き換わり、所有者列は SET に無ければ被害者のまま残る。
func checkUpsertSites(sites []upsertSite, tables map[string]map[string]bool) []string {
	var out []string
	for _, s := range sites {
		cols, ok := tables[s.table]
		if !ok {
			out = append(out, fmt.Sprintf(
				"%s:%d: %s: INSERT 先の表 %q がスキーマに見つかりません。表名の誤り、またはスキーマ側の追加漏れです",
				s.file, s.line, s.queryName, s.table,
			))
			continue
		}

		inKey := map[string]bool{}
		for _, c := range s.keyCols {
			inKey[c] = true
		}
		var missing []string
		for col := range cols {
			if _, own := ownershipColumns[col]; own && !inKey[col] {
				missing = append(missing, col)
			}
		}
		if len(missing) == 0 {
			continue
		}
		sort.Strings(missing)

		if s.exemption != "" {
			continue
		}

		var unguarded []string
		for _, col := range missing {
			if !mentionsIdent(s.guard, col) {
				unguarded = append(unguarded, col)
			}
		}
		if len(unguarded) == 0 {
			continue
		}

		labelled := make([]string, 0, len(unguarded))
		for _, col := range unguarded {
			labelled = append(labelled, fmt.Sprintf("%s（%s）", col, ownershipColumns[col]))
		}

		out = append(out, fmt.Sprintf(
			"%s:%d: %s: %s への upsert の衝突キー (%s) に所有者列 %s が入っておらず、DO UPDATE の WHERE でも絞られていません。"+
				"他人が同じ衝突キーで書くと、その人の行に当たって中身だけが置き換わります（所有者列は被害者のまま残るので、書いた人も記録に残りません）。"+
				"DO UPDATE に WHERE %s.%s = EXCLUDED.%s のような所有者条件を足してください。"+
				"所有者の概念が無い表（導出キャッシュなど）なら、クエリのコメントに `-- upsert-owner-scope: <理由>` と書いて理由ごと免除してください",
			s.file, s.line, s.queryName, s.table,
			strings.Join(s.keyCols, ", "), strings.Join(labelled, ", "),
			s.table, unguarded[0], unguarded[0],
		))
	}
	return out
}

func mentionsIdent(text, ident string) bool {
	if text == "" {
		return false
	}
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(ident) + `\b`).MatchString(text)
}

func readSQLFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err, "%s を読めません。ディレクトリを移したならこのテストの定数も直してください", dir)

	out := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, rerr := os.ReadFile(path)
		require.NoError(t, rerr, "%s を読めません", path)
		out[path] = string(b)
	}
	require.NotEmpty(t, out, "%s に .sql が 1 つもありません", dir)
	return out
}

// Test_upsertの衝突キーに所有者列が入っていること は、
// `INSERT ... ON CONFLICT (...) DO UPDATE` が他人の行に当たらないことを機械で守る。
//
// 何が起きたか（この検査の出どころ）:
//
//	session_notes は「1 セッション 1 メモ」なので session_id に一意制約を張り、
//	ON CONFLICT (session_id) DO UPDATE で upsert していた。ところが衝突キーに所有者
//	（user_id）が入っていない。他人が同じ session_id へ書くと、衝突先は被害者の行になり、
//	content だけが置き換わる。user_id は SET に無いので被害者のまま残り、結果として
//	「自分のメモが知らないうちに書き換わっているが、誰が書いたのかはどこにも残っていない」
//	という状態になる。今は DO UPDATE に WHERE session_notes.user_id = EXCLUDED.user_id を
//	足して塞いである。
//
// なぜテストでもローカルでも気づけないか:
//
//	SQL としては完全に正しく、sqlc generate も sqlc vet も通る。単体でも結合でも、
//	テストは普通「自分の行を作って自分で更新する」ので衝突先は必ず自分の行になり、
//	条件そのものが起きない。壊れるのは「他人が同じ衝突キーで書いた」ときだけで、
//	しかもエラーではなくデータの上書きとして現れるため、ログにも残らない。
//
// 判定の作り:
//
//	INSERT 先の表の列をスキーマから引き、所有者・所属を表す列（ownershipColumns）が
//	衝突キーに含まれているかを見る。含まれていなければ、DO UPDATE の WHERE でその列が
//	絞られているか、免除コメントがあるかを求める。
//	所有者列を 1 つも持たない表（page_snapshots のような導出キャッシュ）は、そもそも
//	守るべき持ち主がいないので何も要求しない。
//	ON CONFLICT ON CONSTRAINT <名前> 形式は衝突キーが読めないので、キー無しとして
//	扱う（＝ WHERE か免除を求める安全側）。DO NOTHING は既存行を書き換えないので対象外。
//
// 既知の限界:
//
//	WHERE の中身が本当に所有者を絞っているかまでは見ておらず、「その列名に触れていること」
//	しか確かめていない。厳密にやるには述語の意味解析が要る。逆側（列名に触れてもいない
//	WHERE は絶対に所有者を絞っていない）は確実なので、素通りだけは防げる。
func Test_upsertの衝突キーに所有者列が入っていること(t *testing.T) {
	tables := map[string]map[string]bool{}
	for _, src := range readSQLFiles(t, schemaDir) {
		for name, cols := range parseSchemaTables(src) {
			tables[name] = cols
		}
	}
	require.NotEmpty(t, tables, "スキーマから表を 1 つも読み取れませんでした。検査が空振りしています")

	var sites []upsertSite
	files := readSQLFiles(t, queriesDir)
	names := make([]string, 0, len(files))
	for path := range files {
		names = append(names, path)
	}
	sort.Strings(names)
	for _, path := range names {
		sites = append(sites, scanUpsertSites(path, files[path])...)
	}
	require.NotEmpty(t, sites, "ON CONFLICT ... DO UPDATE を 1 つも見つけられませんでした。走査が壊れています")

	violations := checkUpsertSites(sites, tables)
	require.Empty(t, violations, "所有者で絞られていない upsert があります:\n%s", strings.Join(violations, "\n"))
}

// Test_upsert検査が危ない書き方を実際に弾くこと は、上の検査そのものが働いていることを
// 見本の SQL で確かめる。検査が黙って壊れる（何も拾わなくなる）のを防ぐため。
func Test_upsert検査が危ない書き方を実際に弾くこと(t *testing.T) {
	tables := map[string]map[string]bool{
		"session_notes":  {"id": true, "session_id": true, "user_id": true, "content": true},
		"page_snapshots": {"page_id": true, "doc": true, "built_at": true},
		"profiles":       {"user_id": true, "bio": true},
		"space_grants":   {"workspace_id": true, "space_id": true, "principal_id": true, "role": true},
	}

	tests := []struct {
		name    string
		sql     string
		violate bool
	}{
		{
			name: "所有者列が衝突キーに無く WHERE も無い（session_notes で実際に起きた形）",
			sql: `-- name: BadUpsert :one
INSERT INTO session_notes (session_id, user_id, content) VALUES ($1, $2, $3)
ON CONFLICT (session_id) DO UPDATE SET content = EXCLUDED.content
RETURNING id;`,
			violate: true,
		},
		{
			name: "DO UPDATE の WHERE で所有者を絞っている（今の session_note.sql の形）",
			sql: `-- name: GoodUpsertGuarded :one
INSERT INTO session_notes (session_id, user_id, content) VALUES ($1, $2, $3)
ON CONFLICT (session_id) DO UPDATE SET content = EXCLUDED.content
WHERE session_notes.user_id = EXCLUDED.user_id
RETURNING id;`,
			violate: false,
		},
		{
			name: "WHERE はあるが所有者列に触れていない",
			sql: `-- name: BadUpsertUnrelatedWhere :one
INSERT INTO session_notes (session_id, user_id, content) VALUES ($1, $2, $3)
ON CONFLICT (session_id) DO UPDATE SET content = EXCLUDED.content
WHERE session_notes.content <> EXCLUDED.content
RETURNING id;`,
			violate: true,
		},
		{
			name: "衝突キーに所有者列が入っている",
			sql: `-- name: GoodUpsertKeyed :one
INSERT INTO profiles (user_id, bio) VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET bio = EXCLUDED.bio
RETURNING user_id;`,
			violate: false,
		},
		{
			name: "複合キーに所有者列がすべて入っている",
			sql: `-- name: GoodUpsertComposite :one
INSERT INTO space_grants (workspace_id, space_id, principal_id, "role") VALUES ($1, $2, $3, $4)
ON CONFLICT (workspace_id, space_id, principal_id)
DO UPDATE SET "role" = EXCLUDED."role"
RETURNING workspace_id;`,
			violate: false,
		},
		{
			name: "所有者列を持たない表（導出キャッシュ）は要求しない",
			sql: `-- name: GoodSnapshot :exec
INSERT INTO page_snapshots (page_id, doc, built_at) VALUES ($1, $2, now())
ON CONFLICT (page_id) DO UPDATE SET doc = EXCLUDED.doc;`,
			violate: false,
		},
		{
			name: "DO NOTHING は既存行を書き換えないので対象外",
			sql: `-- name: GoodDoNothing :exec
INSERT INTO session_notes (session_id, user_id, content) VALUES ($1, $2, $3)
ON CONFLICT (session_id) DO NOTHING;`,
			violate: false,
		},
		{
			name: "ON CONSTRAINT 形式は衝突キーが読めないので安全側に倒す",
			sql: `-- name: BadUpsertOnConstraint :exec
INSERT INTO session_notes (session_id, user_id, content) VALUES ($1, $2, $3)
ON CONFLICT ON CONSTRAINT uq_session_notes_session_id DO UPDATE SET content = EXCLUDED.content;`,
			violate: true,
		},
		{
			name: "免除コメントがあれば通す",
			sql: `-- name: ExemptedUpsert :exec
-- upsert-owner-scope: 検査そのものの見本。理由を書けば免除される
INSERT INTO session_notes (session_id, user_id, content) VALUES ($1, $2, $3)
ON CONFLICT (session_id) DO UPDATE SET content = EXCLUDED.content;`,
			violate: false,
		},
		{
			name: "コメントの中に書かれた ON CONFLICT は数えない",
			sql: `-- name: CommentOnly :one
-- ON CONFLICT (session_id) DO UPDATE の形は使っていない
SELECT id FROM session_notes WHERE session_id = $1 AND user_id = $2;`,
			violate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkUpsertSites(scanUpsertSites("fixture.sql", tt.sql), tables)
			if tt.violate {
				require.NotEmpty(t, got, "違反として弾かれるはずですが素通りしました")
				return
			}
			require.Empty(t, got, "正当な書き方ですが弾かれました:\n%s", strings.Join(got, "\n"))
		})
	}
}

// Test_クエリがsqlcSliceを使っていないこと は、IN 句のスライス展開 sqlc.slice を禁じる。
//
// 何が起きるか（engine postgresql + sql_package database/sql + sqlc v1.31.1 で実測）:
//
//	sqlc.slice は本来「実行時に ? を要素数ぶんへ展開する」仕組みで、生成コードには
//	strings.Replace(query, "/*SLICE:ids*/?", ...) という書き換えが出る。ところが
//	PostgreSQL 向けの生成では SQL 定数を出したあとにプレースホルダを ? から $N へ
//	振り直す工程が入り、目印の "/*SLICE:ids*/?" ごと消えて IN ($1) になる。
//	実測した生成物がまさにそれで、定数の中に目印が無い。つまり上の Replace は何にも
//	一致せず、書き換えは起きない。結果、SQL のプレースホルダは 1 個のままで、渡す値だけが
//	要素数ぶんになる。
//
// なぜテストでもローカルでも気づけないか:
//
//	生成物は正しい Go なのでビルドも go vet も通り、DB を触らない単体テストでも起きない。
//	さらに要素が 1 個のときだけ IN ($1) と値 1 個で数が偶然合って正しく動くため、手元の
//	動作確認をすり抜けやすい。実際に落ちるのは 2 個以上（または 0 個）を渡して
//	PostgreSQL に実行させたときだけになる。
//
// なぜ sqlc vet に書けないか:
//
//	vet に渡る query.sql では sqlc.slice は既に $N へ書き換わっており、CEL からは
//	sqlc.slice という文字列が見えない（実測。sqlc.yaml 冒頭に記載）。
//
// 代わりの書き方:
//
//	id の集まりは json 配列 1 個のパラメータで渡し、json_array_elements_text で展開する。
//	実例は master_exercise_example.sql の ListMasterExerciseExamplesByExerciseIDs。
func Test_クエリがsqlcSliceを使っていないこと(t *testing.T) {
	var found []string
	for path, src := range readSQLFiles(t, queriesDir) {
		for _, m := range sqlcSliceRe.FindAllStringIndex(stripSQLComments(src), -1) {
			found = append(found, fmt.Sprintf("%s:%d", path, lineOf(src, m[0])))
		}
	}
	sort.Strings(found)
	require.Empty(t, found,
		"sqlc.slice はこの設定（postgresql + database/sql）では展開されず、"+
			"プレースホルダ 1 個に対して要素数ぶんの値を渡す生成になります。"+
			"json 配列 1 個のパラメータで渡し json_array_elements_text で展開してください"+
			"（実例: master_exercise_example.sql の ListMasterExerciseExamplesByExerciseIDs）:\n%s",
		strings.Join(found, "\n"))
}

// Test_sqlcSlice検査がコメントと本文を区別すること は、上の検査が本文だけを見ていることを確かめる。
func Test_sqlcSlice検査がコメントと本文を区別すること(t *testing.T) {
	body := `-- name: Q :many
-- sqlc.slice は使わない（この行はコメント）
SELECT 1;`
	require.Empty(t, sqlcSliceRe.FindAllStringIndex(stripSQLComments(body), -1),
		"コメント内の記述を拾ってしまっています")

	real := `-- name: Q :many
SELECT * FROM t WHERE id IN (sqlc.slice(ids));`
	require.Len(t, sqlcSliceRe.FindAllStringIndex(stripSQLComments(real), -1), 1,
		"本文の sqlc.slice を拾えていません")
}
