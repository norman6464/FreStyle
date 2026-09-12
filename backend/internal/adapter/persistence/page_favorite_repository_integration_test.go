//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPageFavoriteRepository_Integration は page_favorites（段7）を実 PostgreSQL で固定する:
// 付ける/外すの冪等性・ワークスペース分離・ページ/ユーザー削除での CASCADE の 3 点。
func TestPageFavoriteRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewPageFavoriteRepository(sqlDB)
	ctx := context.Background()

	t.Run("付けるのは冪等_2回目はcreated falseで行は増えない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-a")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		user := createUser(t, sqlDB, "fav-a")

		created1, err := repo.Add(ctx, ws, page, user)
		require.NoError(t, err)
		assert.True(t, created1)

		created2, err := repo.Add(ctx, ws, page, user)
		require.NoError(t, err)
		assert.False(t, created2, "2回目は既に付いているので新規ではない")

		assert.Equal(t, 1, countPageFavoriteRows(t, sqlDB, user, page))
	})

	t.Run("外すのは冪等_付いていなくてもエラーにしない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-b")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		user := createUser(t, sqlDB, "fav-b")

		require.NoError(t, repo.Remove(ctx, page, user))

		_, err := repo.Add(ctx, ws, page, user)
		require.NoError(t, err)
		require.NoError(t, repo.Remove(ctx, page, user))
		assert.Equal(t, 0, countPageFavoriteRows(t, sqlDB, user, page))
	})

	t.Run("IsFavoriteは付け外しの実際の状態を返す", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-c")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		user := createUser(t, sqlDB, "fav-c")

		before, err := repo.IsFavorite(ctx, page, user)
		require.NoError(t, err)
		assert.False(t, before)

		_, err = repo.Add(ctx, ws, page, user)
		require.NoError(t, err)
		after, err := repo.IsFavorite(ctx, page, user)
		require.NoError(t, err)
		assert.True(t, after)
	})

	t.Run("別ワークスペースのページIDを渡すとErrPageNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		wsA := createWorkspace(t, sqlDB, "ws-d")
		wsB := createWorkspace(t, sqlDB, "ws-e")
		spaceA := createSpace(t, sqlDB, wsA, "s1")
		page := createPage(t, sqlDB, wsA, spaceA, nil, "a0")
		user := createUser(t, sqlDB, "fav-d")

		_, err := repo.Add(ctx, wsB, page, user)
		assert.ErrorIs(t, err, repository.ErrPageNotFound)
		assert.Equal(t, 0, countPageFavoriteRows(t, sqlDB, user, page))
	})

	t.Run("一覧はワークスペース単位で別ワークスペースは混ざらない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		wsA := createWorkspace(t, sqlDB, "ws-f")
		wsB := createWorkspace(t, sqlDB, "ws-g")
		spaceA := createSpace(t, sqlDB, wsA, "s1")
		spaceB := createSpace(t, sqlDB, wsB, "s1")
		pageA := createPage(t, sqlDB, wsA, spaceA, nil, "a0")
		pageB := createPage(t, sqlDB, wsB, spaceB, nil, "a0")
		user := createUser(t, sqlDB, "fav-e")

		_, err := repo.Add(ctx, wsA, pageA, user)
		require.NoError(t, err)
		_, err = repo.Add(ctx, wsB, pageB, user)
		require.NoError(t, err)

		outA, err := repo.ListFavorites(ctx, wsA, user)
		require.NoError(t, err)
		require.Len(t, outA, 1)
		assert.Equal(t, pageA, outA[0].PageID)
	})

	t.Run("ページ削除で行が消える", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-h")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		user := createUser(t, sqlDB, "fav-f")
		_, err := repo.Add(ctx, ws, page, user)
		require.NoError(t, err)

		_, err = sqlDB.Exec(`DELETE FROM pages WHERE id = $1`, page)
		require.NoError(t, err)
		assert.Equal(t, 0, countPageFavoriteRows(t, sqlDB, user, page))
	})

	t.Run("ユーザー削除で行が消える", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, kbTables...)
		ws := createWorkspace(t, sqlDB, "ws-i")
		space := createSpace(t, sqlDB, ws, "s1")
		page := createPage(t, sqlDB, ws, space, nil, "a0")
		user := createUser(t, sqlDB, "fav-g")
		_, err := repo.Add(ctx, ws, page, user)
		require.NoError(t, err)

		_, err = sqlDB.Exec(`DELETE FROM users WHERE id = $1`, user)
		require.NoError(t, err)
		assert.Equal(t, 0, countPageFavoriteRows(t, sqlDB, user, page))
	})
}

func countPageFavoriteRows(t *testing.T, db *sql.DB, userID uint64, pageID string) int {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM page_favorites WHERE user_id = $1 AND page_id = $2`, userID, pageID,
	).Scan(&n))
	return n
}
