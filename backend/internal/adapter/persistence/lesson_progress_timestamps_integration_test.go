//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestLessonProgressRepository_Timestamps_Integration は完了記録の作成時に
// completed_at / created_at が非ゼロで書かれることを固定する（NOT NULL 列に時刻が入る）。
func TestLessonProgressRepository_Timestamps_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewLessonProgressRepository(db)
	ctx := context.Background()

	testsupport.TruncateAll(t, db, "user_chapter_progress")

	changed, err := repo.MarkCompleted(ctx, 1, 10, 100)
	require.NoError(t, err)
	require.True(t, changed)

	rows, err := repo.ListByUser(ctx, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.False(t, rows[0].CompletedAt.IsZero(), "completed_at が入る")
	require.False(t, rows[0].CreatedAt.IsZero(), "created_at が入る")
	require.NotZero(t, rows[0].ID, "id が採番される")
	require.Equal(t, uint64(100), rows[0].CourseID)
}
