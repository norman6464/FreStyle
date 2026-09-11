//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noSuchWorkspaceID はどのワークスペースにも該当しない UUID。
const noSuchWorkspaceID = "0198a000-0000-7000-8000-0000000000ff"

// noSuchPageID はどのページにも該当しない UUID。
const noSuchPageID = "0198a000-0000-7000-8000-0000000000fe"

// noSuchThreadID はどのコメントスレッドにも該当しない UUID。
const noSuchThreadID = "0198a000-0000-7000-8000-0000000000fd"

// listCase は「一覧を返す repository メソッド」1 件分の検証定義。
//
// call は any のスライスを返す形に揃える（要素型が異なるため）。
type listCase struct {
	name string
	call func(ctx context.Context, db *sql.DB) (any, error)
}

func listCases() []listCase {
	return []listCase{
		{
			name: "ページのコメントスレッド一覧",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewCommentRepository(db).
					ListCommentThreadsByPage(ctx, noSuchWorkspaceID, noSuchPageID)
			},
		},
		{
			name: "スレッドの返信一覧",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewCommentRepository(db).
					ListCommentsByThreads(ctx, []string{noSuchThreadID})
			},
		},
	}
}

// TestPersistence_一覧が0件でもnullではなく空配列を返すこと_Integration は、
// 一覧を返す repository メソッドが「該当行なし」で nil を返さないことを実 DB で検証する。
//
// nil スライスは encoding/json で null になり、フロントの map / filter / for-of が
// TypeError で落ちる（staging 実機で観測）。新規ワークスペース・空のページという、
// 新メンバーが最初に踏む動線で発生するため影響が大きい。
func TestPersistence_一覧が0件でもnullではなく空配列を返すこと_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	for _, tc := range listCases() {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.call(ctx, sqlDB)

			require.NoError(t, err)
			assert.NotNil(t, got, "0 件でも nil スライスを返さないこと")
			encoded, marshalErr := json.Marshal(got)
			require.NoError(t, marshalErr)
			assert.Equal(t, "[]", string(encoded), "JSON が null ではなく [] になること")
		})
	}
}

// TestPersistence_一覧取得の失敗はエラーとして返ること_Integration は、
// 空配列の契約が「エラーを握り潰して空配列を返す」ことにならないよう異常系を固定する。
//
// 取得できなかったことを空配列として返すと、利用者には「0 件」と区別がつかず
// 障害に気づけなくなる。context を中断した状態で必ずエラーが返ることを確認する。
func TestPersistence_一覧取得の失敗はエラーとして返ること_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	canceled, cancel := context.WithCancel(context.Background())
	cancel() // 実行前に中断しておく

	for _, tc := range listCases() {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.call(canceled, sqlDB)

			assert.Error(t, err, "取得に失敗したらエラーを返すこと（空配列で握り潰さない）")
		})
	}
}
