package repository

import "context"

// SQLQueryResult は read-only SQL コンソールの実行結果。列名と行（各セルは任意の型）を持つ。
// MaxRows を超えた場合は Truncated=true で打ち切る。
type SQLQueryResult struct {
	Columns   []string `json:"columns"`
	Rows      [][]any  `json:"rows"`
	Truncated bool     `json:"truncated"`
}

// SQLConsoleRepository は super_admin 向け read-only SQL コンソールの実行ポート。
// 実装は read-only トランザクション内で 1 クエリのみ実行し、行数上限とタイムアウトを課す
// （万一バリデーションをすり抜けた write も Postgres の read-only トランザクションが拒否する）。
type SQLConsoleRepository interface {
	RunReadOnly(ctx context.Context, query string, maxRows int) (*SQLQueryResult, error)
}
