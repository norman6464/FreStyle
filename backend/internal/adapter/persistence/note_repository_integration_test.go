//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestNoteRepository_Integration は NoteRepository の所有権スコープと並び順を実 Postgres で検証する。
func TestNoteRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewNoteRepository(db)
	ctx := context.Background()

	t.Run("ListByUserID は自分の note だけを updated_at DESC で返す", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "notes")

		older := &domain.Note{UserID: 7, Title: "older", UpdatedAt: time.Now().Add(-time.Hour)}
		newer := &domain.Note{UserID: 7, Title: "newer", UpdatedAt: time.Now()}
		other := &domain.Note{UserID: 8, Title: "someone-else"}
		require.NoError(t, repo.Create(ctx, older))
		require.NoError(t, repo.Create(ctx, newer))
		require.NoError(t, repo.Create(ctx, other))

		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		require.Len(t, rows, 2, "user 8 の note は WHERE user_id で除外される")
		require.Equal(t, "newer", rows[0].Title, "updated_at DESC で新しい順")
		require.Equal(t, "older", rows[1].Title)
	})

	t.Run("Delete は user_id スコープで他人の note を消さない", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "notes")

		mine := &domain.Note{UserID: 7, Title: "mine"}
		theirs := &domain.Note{UserID: 8, Title: "theirs"}
		require.NoError(t, repo.Create(ctx, mine))
		require.NoError(t, repo.Create(ctx, theirs))

		// user 7 が user 8 の note を消そうとしても WHERE user_id=7 で no-op。
		require.NoError(t, repo.Delete(ctx, 7, theirs.ID))
		got, err := repo.FindByID(ctx, theirs.ID)
		require.NoError(t, err, "他人の note は残る")
		require.Equal(t, "theirs", got.Title)

		// 自分の note は消せる。
		require.NoError(t, repo.Delete(ctx, 7, mine.ID))
		_, err = repo.FindByID(ctx, mine.ID)
		require.Error(t, err, "自分の note は削除済み")
	})

	t.Run("Update は内容を保存する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "notes")

		n := &domain.Note{UserID: 7, Title: "before", Content: "x"}
		require.NoError(t, repo.Create(ctx, n))
		n.Title = "after"
		require.NoError(t, repo.Update(ctx, n))

		got, err := repo.FindByID(ctx, n.ID)
		require.NoError(t, err)
		require.Equal(t, "after", got.Title)
	})
}

// TestNoteRepository_Timestamps_Integration は Create が created_at / updated_at を now で埋め、
// Update が updated_at を進めつつ created_at を保つこと（GORM autoTime 相当）を固定する。
func TestNoteRepository_Timestamps_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewNoteRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "notes")

	before := time.Now().Add(-time.Second)
	n := &domain.Note{UserID: 7, Title: "before", Content: "x"}
	require.NoError(t, repo.Create(ctx, n))
	require.False(t, n.CreatedAt.IsZero(), "created_at が書き戻される")
	require.False(t, n.UpdatedAt.IsZero(), "updated_at が書き戻される")
	require.True(t, n.CreatedAt.After(before), "created_at は now 付近")
	createdAt := n.CreatedAt

	time.Sleep(5 * time.Millisecond)
	n.Title = "after"
	require.NoError(t, repo.Update(ctx, n))

	got, err := repo.FindByID(ctx, n.ID)
	require.NoError(t, err)
	require.Equal(t, "after", got.Title)
	require.WithinDuration(t, createdAt, got.CreatedAt, time.Millisecond, "created_at は更新で変わらない")
	require.True(t, got.UpdatedAt.After(createdAt), "updated_at は更新で進む")
}
