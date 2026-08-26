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

// TestProfileRepository_Integration は sqlc 化した FindByUserID（round-trip / not-found）を実 Postgres で検証する。
func TestProfileRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewProfileRepository(sqlDB)
	ctx := context.Background()

	t.Run("FindByUserID は profile を返す", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "profiles")
		_, err := sqlDB.ExecContext(
			ctx,
			`INSERT INTO profiles (user_id, bio, avatar_url, status_message, updated_at)
			 VALUES ($1, $2, $3, $4, now())`,
			7, "自己紹介", "https://example.com/a.png", "active",
		)
		require.NoError(t, err)

		got, err := repo.FindByUserID(ctx, 7)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, uint64(7), got.UserID)
		require.Equal(t, "自己紹介", got.Bio)
		require.Equal(t, "https://example.com/a.png", got.AvatarURL)
		require.Equal(t, "active", got.StatusMessage)
	})

	t.Run("未作成は (nil, nil)", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "profiles")
		got, err := repo.FindByUserID(ctx, 999)
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

// TestProfileRepository_Upsert_Integration は Upsert の意味論を実 Postgres で固定する。
// 未作成なら作成し、既存なら user_id 単位で 1 行を更新する（重複行を作らない）。
// updated_at が書き戻されることも確認する（GORM Save 相当）。
func TestProfileRepository_Upsert_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewProfileRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "profiles")

	// 1) 未作成 → 作成される。
	p := &domain.Profile{UserID: 42, Bio: "v1", AvatarURL: "a1", StatusMessage: "s1"}
	require.NoError(t, repo.Upsert(ctx, p))
	require.False(t, p.UpdatedAt.IsZero(), "updated_at が書き戻される")

	got, err := repo.FindByUserID(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "v1", got.Bio)

	// 2) 既存 → 同じ user_id で更新（重複行を作らない）。
	p2 := &domain.Profile{UserID: 42, Bio: "v2", AvatarURL: "a2", StatusMessage: "s2"}
	require.NoError(t, repo.Upsert(ctx, p2))

	got, err = repo.FindByUserID(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, uint64(42), got.UserID, "user_id は保持される")
	require.Equal(t, "v2", got.Bio, "既存行が更新される")
	require.Equal(t, "a2", got.AvatarURL)
	require.Equal(t, "s2", got.StatusMessage)

	var cnt int64
	require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM profiles WHERE user_id = $1`, 42).Scan(&cnt))
	require.Equal(t, int64(1), cnt, "user_id 単位の upsert なので行は 1 つ")
}
