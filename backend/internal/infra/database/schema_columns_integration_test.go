//go:build integration

package database_test

import (
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestCoreSchema_列の型が本番の実列と一致する_Integration は、schema.hcl から生成した DDL を
// 流した直後のスキーマの列型を固定する。
//
// ここに並ぶ列は、宣言（schema.hcl）と本番の実列がずれていたものを本番側へ合わせた
// 結果そのもの。型が合っていないと、sqlc の生成型と実列の型が食い違う。
// 手で 1 回突き合わせた事実を CI に残すためのテストなので、緩めない。
func TestCoreSchema_列の型が本番の実列と一致する_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)

	tests := []struct {
		table  string
		column string
		typ    string
		maxLen int64 // varchar(n) の n。0 なら長さを持たない型
		why    string
	}{
		{"users", "role", "text", 0, "role はロール名そのものを持つ列（かつては roles マスタ表への FK だった）"},
		// migration 0011 が ALTER ADD COLUMN で作った列。AutoMigrate は fresh DB に bigint を作っていた。
		{"master_exercises", "sort_order", "integer", 0, "本番の master_exercises.sort_order は integer"},
	}
	for _, tt := range tests {
		t.Run(tt.table+"."+tt.column, func(t *testing.T) {
			typ, maxLen := columnType(t, db, tt.table, tt.column)
			require.Equal(t, tt.typ, typ, tt.why)
			require.Equal(t, tt.maxLen, maxLen, tt.why)
		})
	}
}

// columnType は列の SQL 型と varchar の最大長を返す（長さを持たない型は 0）。
func columnType(t *testing.T, db *sql.DB, table, column string) (string, int64) {
	t.Helper()
	var typ string
	var maxLen sql.NullInt64
	err := db.QueryRowContext(
		t.Context(),
		`SELECT data_type, character_maximum_length
		   FROM information_schema.columns
		  WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2`,
		table, column,
	).Scan(&typ, &maxLen)
	require.NoError(t, err, "%s.%s の列情報を引けない", table, column)
	return typ, maxLen.Int64
}
