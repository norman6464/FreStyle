package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

// applyEmbeddedSchema は go:embed した DDL を 1 トランザクションで適用する（冪等）。
//
// 索引は「まだ存在しないものだけ」を実行する。CREATE INDEX IF NOT EXISTS は、既存の索引が
// あってスキップする場合でも対象テーブルの ShareLock を取り、トランザクションが終わるまで
// 手放さない（CREATE TABLE IF NOT EXISTS は取らない）。全 DDL を 1 つのトランザクションへ
// まとめて流すこの経路では、索引を持つ中核テーブルすべての ShareLock を同時に握ったまま
// 起動時マイグレーションが走ることになり、その間そのテーブルへの INSERT / UPDATE / DELETE が
// すべて詰まる。索引が出揃っている通常の起動では CREATE INDEX を 1 本も発行しないようにして、
// ロックを取らずに通り抜けさせる（カタログを先に引いてから必要な文だけ流す形）。
//
// 索引が本当に足りないときだけロックを取る。その待ち時間は withMigrateTx の lock_timeout で
// 有限に抑える。
func applyEmbeddedSchema(ctx context.Context, db *sql.DB, label string, ddls ...string) error {
	return withMigrateTx(ctx, db, label, func(tx *sql.Tx) error {
		existing, err := existingIndexNames(ctx, tx)
		if err != nil {
			return fmt.Errorf("既存索引の一覧取得に失敗: %w", err)
		}
		for _, ddl := range ddls {
			for _, stmt := range splitSQLStatements(ddl) {
				if name := createIndexName(stmt); name != "" && existing[name] {
					continue
				}
				if _, err := tx.ExecContext(ctx, stmt); err != nil {
					return fmt.Errorf("DDL の適用に失敗（%s）: %w", summarizeStatement(stmt), err)
				}
			}
		}
		return nil
	})
}

// existingIndexNames は search_path 上に既に在る索引名の集合を返す。
// 参照するのはカタログだけなので、対象テーブルのロックは取らない。
func existingIndexNames(ctx context.Context, tx *sql.Tx) (map[string]bool, error) {
	rows, err := tx.QueryContext(
		ctx,
		`SELECT indexname FROM pg_indexes WHERE schemaname = ANY (current_schemas(false))`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	names := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

// indexExists は指定名の索引が search_path 上に在るかを返す。
// CREATE INDEX を出す前の事前チェック用（実行してしまうとスキップされてもロックを取るため）。
func indexExists(ctx context.Context, db Executor, name string) (bool, error) {
	var n int64
	if err := db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM pg_indexes
		  WHERE schemaname = ANY (current_schemas(false)) AND indexname = $1`,
		name,
	).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// createIndexRe は CREATE INDEX 文から索引名を取り出す。
var createIndexRe = regexp.MustCompile(`(?is)^CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:CONCURRENTLY\s+)?(?:IF\s+NOT\s+EXISTS\s+)?("[^"]+"|[A-Za-z_][A-Za-z0-9_$]*)`)

// createIndexName は CREATE INDEX 文の索引名を返す。CREATE INDEX 文でなければ空文字を返す。
// 引用符なしの識別子は PostgreSQL が小文字へ畳むので、カタログの値と突き合わせられるよう
// ここでも小文字にする。
func createIndexName(stmt string) string {
	m := createIndexRe.FindStringSubmatch(stripLeadingComments(stmt))
	if m == nil {
		return ""
	}
	if name := m[1]; strings.HasPrefix(name, `"`) {
		return strings.Trim(name, `"`)
	}
	return strings.ToLower(m[1])
}

// splitSQLStatements は DDL 本文をセミコロン区切りの文へ分割する。
// 行コメント（--）/ ブロックコメント（/* */・入れ子可）/ 引用符（' " $tag$）の内側の
// セミコロンでは切らない。schema/*.sql を 1 文ずつ実行するための最小限の分割器であって、
// 汎用の SQL パーサではない。
func splitSQLStatements(ddl string) []string {
	var out []string
	start, i, n := 0, 0, len(ddl)
	for i < n {
		switch {
		case strings.HasPrefix(ddl[i:], "--"):
			i = skipLineComment(ddl, i)
		case strings.HasPrefix(ddl[i:], "/*"):
			i = skipBlockComment(ddl, i)
		case ddl[i] == '\'' || ddl[i] == '"':
			i = skipQuoted(ddl, i)
		case ddl[i] == '$':
			i = skipDollarQuoted(ddl, i)
		case ddl[i] == ';':
			out = appendStatement(out, ddl[start:i])
			i++
			start = i
		default:
			i++
		}
	}
	return appendStatement(out, ddl[start:])
}

// skipLineComment は行コメントの終端（改行の次）まで進めた位置を返す。
func skipLineComment(s string, i int) int {
	if j := strings.IndexByte(s[i:], '\n'); j >= 0 {
		return i + j + 1
	}
	return len(s)
}

// skipBlockComment はブロックコメントの終端まで進めた位置を返す（入れ子に対応）。
func skipBlockComment(s string, i int) int {
	depth := 1
	i += 2
	for i < len(s) && depth > 0 {
		switch {
		case strings.HasPrefix(s[i:], "/*"):
			depth++
			i += 2
		case strings.HasPrefix(s[i:], "*/"):
			depth--
			i += 2
		default:
			i++
		}
	}
	return i
}

// skipQuoted は ' または " で囲まれた範囲の終端まで進めた位置を返す（” / "" の二重化に対応）。
func skipQuoted(s string, i int) int {
	q := s[i]
	i++
	for i < len(s) {
		if s[i] != q {
			i++
			continue
		}
		if i+1 < len(s) && s[i+1] == q {
			i += 2
			continue
		}
		return i + 1
	}
	return i
}

// skipDollarQuoted は $tag$ … $tag$ の終端まで進めた位置を返す。
// $ が dollar quote の開始でなければ 1 文字だけ進める。
func skipDollarQuoted(s string, i int) int {
	tag, ok := dollarQuoteTag(s[i:])
	if !ok {
		return i + 1
	}
	rest := i + len(tag)
	if j := strings.Index(s[rest:], tag); j >= 0 {
		return rest + j + len(tag)
	}
	return len(s)
}

// dollarQuoteTag は先頭が dollar quote の開始なら区切り記号（$tag$）を返す。
func dollarQuoteTag(s string) (string, bool) {
	if s == "" || s[0] != '$' {
		return "", false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c == '$' {
			return s[:i+1], true
		}
		if c != '_' && !isASCIILetter(c) && !isASCIIDigit(c) {
			return "", false
		}
		// タグの 1 文字目は数字にできない（$1 のような位置パラメータと区別するため）。
		if i == 1 && isASCIIDigit(c) {
			return "", false
		}
	}
	return "", false
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isASCIIDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// appendStatement は空白とコメントだけの断片を捨てて文を積む。
func appendStatement(out []string, raw string) []string {
	stmt := strings.TrimSpace(raw)
	if stripLeadingComments(stmt) == "" {
		return out
	}
	return append(out, stmt)
}

// stripLeadingComments は先頭に並ぶ空白とコメントを落とした残りを返す。
// 文の直前に置かれた説明コメントを跨いで先頭キーワードを見るために使う。
func stripLeadingComments(s string) string {
	for {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "--"):
			j := strings.IndexByte(s, '\n')
			if j < 0 {
				return ""
			}
			s = s[j+1:]
		case strings.HasPrefix(s, "/*"):
			j := strings.Index(s, "*/")
			if j < 0 {
				return ""
			}
			s = s[j+2:]
		default:
			return s
		}
	}
}

// summarizeStatement はエラーメッセージ用に文の先頭 1 行を短く整える。
func summarizeStatement(stmt string) string {
	s := stripLeadingComments(stmt)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	const limit = 80
	if r := []rune(s); len(r) > limit {
		return string(r[:limit]) + "…"
	}
	return s
}
