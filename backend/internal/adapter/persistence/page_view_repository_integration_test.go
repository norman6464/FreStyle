//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPageViewRepository_Integration は page_views の upsert 型の設計（段2・簡略版）を
// 実 PostgreSQL で固定する: 同じ人が同じページを何度見ても行は 1 つのまま・
// 閲覧数は行数（見た人数）・ページ/ユーザー削除で CASCADE・可視判定より前の候補取得は
// アーカイブ済みを除く、の 4 点。
func TestPageViewRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewPageViewRepository(sqlDB)
	ctx := context.Background()

	t.Run("同じ人が同じページを何度見ても行は1つのまま_viewed_atだけ進む", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-a")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		user := createUser(t, sqlDB, "viewer")

		require.NoError(t, repo.RecordView(ctx, ws, page, user))
		first := currentViewedAt(t, sqlDB, user, page)

		time.Sleep(10 * time.Millisecond)
		require.NoError(t, repo.RecordView(ctx, ws, page, user))
		second := currentViewedAt(t, sqlDB, user, page)

		assert.True(t, second.After(first), "2回目の viewed_at は1回目より新しいはず")

		n := countPageViewRows(t, sqlDB, user, page)
		assert.Equal(t, 1, n, "同じ(user,page)の行は upsert で1つのまま")
	})

	t.Run("閲覧数は見た人数_延べ回数ではない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-b")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		alice := createUser(t, sqlDB, "alice")
		bob := createUser(t, sqlDB, "bob")

		require.NoError(t, repo.RecordView(ctx, ws, page, alice))
		require.NoError(t, repo.RecordView(ctx, ws, page, alice)) // 延べでは2回だが人数は変わらない
		require.NoError(t, repo.RecordView(ctx, ws, page, bob))

		count, err := repo.CountViews(ctx, page)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("別ワークスペースのページIDを渡すとErrPageNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		wsA := createWorkspace(t, sqlDB, "ws-c")
		wsB := createWorkspace(t, sqlDB, "ws-d")
		spaceA := createSpace(t, sqlDB, wsA, "s1")
		page := createPage(t, sqlDB, wsA, spaceA, nil, "a0")
		user := createUser(t, sqlDB, "mismatched")

		err := repo.RecordView(ctx, wsB, page, user)
		assert.ErrorIs(t, err, repository.ErrPageNotFound)
		assert.Equal(t, 0, countPageViewRows(t, sqlDB, user, page), "行が増えていないこと")
	})

	t.Run("ページ削除で閲覧記録も消える", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-e")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		user := createUser(t, sqlDB, "viewer2")
		require.NoError(t, repo.RecordView(ctx, ws, page, user))
		require.Equal(t, 1, countPageViewRows(t, sqlDB, user, page))

		_, err := sqlDB.Exec(`DELETE FROM pages WHERE id = $1`, page)
		require.NoError(t, err)
		assert.Equal(t, 0, countPageViewRows(t, sqlDB, user, page))
	})

	t.Run("ユーザー削除で閲覧記録も消える", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-f")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		user := createUser(t, sqlDB, "viewer3")
		require.NoError(t, repo.RecordView(ctx, ws, page, user))
		require.Equal(t, 1, countPageViewRows(t, sqlDB, user, page))

		_, err := sqlDB.Exec(`DELETE FROM users WHERE id = $1`, user)
		require.NoError(t, err)
		assert.Equal(t, 0, countPageViewRows(t, sqlDB, user, page))
	})

	t.Run("最近見たページの候補はviewed_atの新しい順_アーカイブ済みは除く", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-g")
		space := createSpace(t, sqlDB, ws, "s1")
		older := createPage(t, sqlDB, ws, space, nil, "a0")
		newer := createPage(t, sqlDB, ws, space, nil, "a1")
		archived := createPage(t, sqlDB, ws, space, nil, "a2")
		user := createUser(t, sqlDB, "recent")

		require.NoError(t, repo.RecordView(ctx, ws, older, user))
		time.Sleep(10 * time.Millisecond)
		require.NoError(t, repo.RecordView(ctx, ws, newer, user))
		time.Sleep(10 * time.Millisecond)
		require.NoError(t, repo.RecordView(ctx, ws, archived, user))
		_, err := sqlDB.Exec(`UPDATE pages SET archived_at = now() WHERE id = $1`, archived)
		require.NoError(t, err)

		out, err := repo.ListRecentPageViewCandidates(ctx, user)
		require.NoError(t, err)
		ids := make([]string, 0, len(out))
		for _, rp := range out {
			ids = append(ids, rp.PageID)
		}
		assert.NotContains(t, ids, archived, "アーカイブ済みは候補から除く")
		require.Contains(t, ids, newer)
		require.Contains(t, ids, older)
		// newer の方が viewed_at が新しいので先に来る。
		newerIdx, olderIdx := -1, -1
		for i, id := range ids {
			if id == newer {
				newerIdx = i
			}
			if id == older {
				olderIdx = i
			}
		}
		assert.Less(t, newerIdx, olderIdx, "新しい順であること")
	})

	t.Run("別ワークスペースの候補は混ざらない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		wsA := createWorkspace(t, sqlDB, "ws-h")
		wsB := createWorkspace(t, sqlDB, "ws-i")
		spaceA := createSpace(t, sqlDB, wsA, "s1")
		spaceB := createSpace(t, sqlDB, wsB, "s1")
		pageA := createPage(t, sqlDB, wsA, spaceA, nil, "a0")
		pageB := createPage(t, sqlDB, wsB, spaceB, nil, "a0")
		user := createUser(t, sqlDB, "cross-ws")

		require.NoError(t, repo.RecordView(ctx, wsA, pageA, user))
		require.NoError(t, repo.RecordView(ctx, wsB, pageB, user))

		out, err := repo.ListRecentPageViewCandidates(ctx, user)
		require.NoError(t, err)
		require.Len(t, out, 2)
		slugs := map[string]bool{}
		for _, rp := range out {
			slugs[rp.WorkspaceSlug] = true
			assert.NotEmpty(t, rp.WorkspaceID)
		}
		assert.Len(t, slugs, 2, "両方のワークスペースが別々に出ること")
	})
}

func currentViewedAt(t *testing.T, db *sql.DB, userID uint64, pageID string) time.Time {
	t.Helper()
	var viewedAt time.Time
	require.NoError(t, db.QueryRow(
		`SELECT viewed_at FROM page_views WHERE user_id = $1 AND page_id = $2`, userID, pageID,
	).Scan(&viewedAt))
	return viewedAt
}

func countPageViewRows(t *testing.T, db *sql.DB, userID uint64, pageID string) int {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM page_views WHERE user_id = $1 AND page_id = $2`, userID, pageID,
	).Scan(&n))
	return n
}
