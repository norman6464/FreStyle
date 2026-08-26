//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestSessionNoteRepository_Integration は sqlc 化した FindBySessionID（round-trip / not-found）を実 Postgres で検証する。
func TestSessionNoteRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewSessionNoteRepository(db)
	ctx := context.Background()

	t.Run("Upsert → FindBySessionID で round-trip", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "session_notes")
		require.NoError(t, repo.Upsert(ctx, &domain.SessionNote{SessionID: 55, UserID: 7, Content: "メモ本文"}))

		got, err := repo.FindBySessionID(ctx, 55)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, uint64(55), got.SessionID)
		require.Equal(t, uint64(7), got.UserID)
		require.Equal(t, "メモ本文", got.Content)
	})

	t.Run("未作成は (nil, nil)", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "session_notes")
		got, err := repo.FindBySessionID(ctx, 999)
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

// TestSessionNoteRepository_Upsert_WriteBack_Integration は Upsert が id / created_at /
// updated_at を書き戻すこと（GORM Save 相当）を実 Postgres で固定する。
// session_id には一意制約が無く、GORM Save(ID=0) は常に INSERT する（現行挙動）。
// sqlc 化後もこの挙動と書き戻しを保つ。
func TestSessionNoteRepository_Upsert_WriteBack_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewSessionNoteRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "session_notes")

	n := &domain.SessionNote{SessionID: 77, UserID: 3, Content: "本文"}
	require.NoError(t, repo.Upsert(ctx, n))
	require.NotZero(t, n.ID, "id が採番され書き戻される")
	require.False(t, n.CreatedAt.IsZero(), "created_at が書き戻される")
	require.False(t, n.UpdatedAt.IsZero(), "updated_at が書き戻される")

	got, err := repo.FindBySessionID(ctx, 77)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, n.ID, got.ID)
	require.Equal(t, "本文", got.Content)
}
