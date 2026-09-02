//go:build integration

package persistence_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ensureCourses は章をぶら下げる先のコースを用意する。
//
// # なぜ要るようになったか
//
// course_chapters には courses への外部キーが無く、実在しないコース ID を書けた。
// 対象ごとの権限を張れるようにするにあたって FK を張ったので、章は必ず実在するコースに
// ぶら下がる（本番でもそうなっている形を、テストでも作る）。
//
// workspaceID は nil を渡せる。**この FK はコース ID しか見ない**ので、章とコースで
// 所属が揃っている必要は無い（テナントの一致は権限の表から張る複合 FK が受け持つ）。
// 所属を検証しないテストは nil でよい。
//
// 既にあるコースには触らない。同じテスト内で 2 回呼んでも、別のテストが先に同じ id を
// 作っていても、そのまま通る。**章を空にするたびに呼ぶ**こと — ワークスペースごと
// 空にする経路では、courses も FK の連鎖で一緒に消える。
func ensureCourses(t *testing.T, db *sql.DB, workspaceID *string, ids ...uint64) {
	t.Helper()
	for _, id := range ids {
		_, err := db.Exec(
			`INSERT INTO courses (id, workspace_id, created_by_user_id, title, created_at, updated_at)
			 VALUES ($1, $2, 1, $3, now(), now())
			 ON CONFLICT (id) DO NOTHING`,
			id, workspaceID, fmt.Sprintf("コース %d", id),
		)
		require.NoError(t, err)
	}
}
