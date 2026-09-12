//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
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
			`INSERT INTO profiles (user_id, bio, avatar_url, status_text, updated_at)
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
		require.Equal(t, "active", got.StatusText)
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
	p := &domain.Profile{UserID: 42, Bio: "v1", AvatarURL: "a1", StatusText: "s1"}
	require.NoError(t, repo.Upsert(ctx, p))
	require.False(t, p.UpdatedAt.IsZero(), "updated_at が書き戻される")

	got, err := repo.FindByUserID(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "v1", got.Bio)

	// 2) 既存 → 同じ user_id で更新（重複行を作らない）。
	p2 := &domain.Profile{UserID: 42, Bio: "v2", AvatarURL: "a2", StatusText: "s2"}
	require.NoError(t, repo.Upsert(ctx, p2))

	got, err = repo.FindByUserID(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, uint64(42), got.UserID, "user_id は保持される")
	require.Equal(t, "v2", got.Bio, "既存行が更新される")
	require.Equal(t, "a2", got.AvatarURL)
	require.Equal(t, "s2", got.StatusText)

	var cnt int64
	require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM profiles WHERE user_id = $1`, 42).Scan(&cnt))
	require.Equal(t, int64(1), cnt, "user_id 単位の upsert なので行は 1 つ")
}

// TestProfileRepository_UpsertProfile_ステータス列には触れない は、一般更新（Upsert）が
// status_emoji / status_expires_at を書き換えないこと、および新規行では既定値
// （空文字 / NULL）が入ることを確認する（段 14。UpdateStatus との担当分離）。
func TestProfileRepository_UpsertProfile_ステータス列には触れない(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewProfileRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "profiles")

	require.NoError(t, repo.Upsert(ctx, &domain.Profile{UserID: 1, Bio: "hi"}))
	got, err := repo.FindByUserID(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "", got.StatusEmoji, "新規行は既定値の空文字のまま")
	require.Nil(t, got.StatusExpiresAt, "新規行は既定値の NULL のまま")

	_, err = repo.UpdateStatus(ctx, 1, "🎉", "休暇中", nil)
	require.NoError(t, err)

	require.NoError(t, repo.Upsert(ctx, &domain.Profile{UserID: 1, Bio: "hi2"}))
	got, err = repo.FindByUserID(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "hi2", got.Bio, "bio は更新される")
	require.Equal(t, "🎉", got.StatusEmoji, "一般更新（Upsert）はステータスに触れない")
	require.Equal(t, "休暇中", got.StatusText)
}

// TestProfileRepository_UpdateStatus_Integration は PUT /me/status が使う UpdateStatus の
// upsert 意味論（bio / avatar_url には触れない）と、失効時刻を過ぎたステータスが
// FindByUserID の応答からだけ空になる（DB の値は残る）ことを実 Postgres で確認する（段 14）。
func TestProfileRepository_UpdateStatus_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewProfileRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "profiles")

	require.NoError(t, repo.Upsert(ctx, &domain.Profile{UserID: 9, Bio: "自己紹介", AvatarURL: "a.png"}))

	future := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	updated, err := repo.UpdateStatus(ctx, 9, "🏖️", "休暇中", &future)
	require.NoError(t, err)
	require.Equal(t, "自己紹介", updated.Bio, "UpdateStatus は bio / avatar_url に触れない")
	require.Equal(t, "a.png", updated.AvatarURL)
	require.Equal(t, "🏖️", updated.StatusEmoji)
	require.Equal(t, "休暇中", updated.StatusText)
	require.NotNil(t, updated.StatusExpiresAt)
	require.WithinDuration(t, future, *updated.StatusExpiresAt, time.Second)

	got, err := repo.FindByUserID(ctx, 9)
	require.NoError(t, err)
	require.Equal(t, "🏖️", got.StatusEmoji, "失効前は見える")
	require.Equal(t, "休暇中", got.StatusText)

	// 失効時刻を過ぎさせる（DB の値はそのまま、応答からだけ空になることを見る）。
	past := time.Now().Add(-time.Hour)
	_, err = sqlDB.ExecContext(ctx, `UPDATE profiles SET status_expires_at = $1 WHERE user_id = $2`, past, 9)
	require.NoError(t, err)

	got, err = repo.FindByUserID(ctx, 9)
	require.NoError(t, err)
	require.Equal(t, "", got.StatusEmoji, "失効していれば応答は空文字")
	require.Equal(t, "", got.StatusText)

	var rawEmoji string
	require.NoError(t, sqlDB.QueryRow(`SELECT status_emoji FROM profiles WHERE user_id = $1`, 9).Scan(&rawEmoji))
	require.Equal(t, "🏖️", rawEmoji, "DB の値そのものは消えていない")

	// UpdateStatus が最初の書き込み（profiles 行が無い状態）でも、bio / avatar_url は
	// 既定値の空文字のまま作られる（INSERT の列挙から外しているため）。
	fresh, err := repo.UpdateStatus(ctx, 10, "😀", "in a meeting", nil)
	require.NoError(t, err)
	require.Equal(t, "", fresh.Bio)
	require.Equal(t, "", fresh.AvatarURL)
	require.Nil(t, fresh.StatusExpiresAt, "expiresAt=nil なら無期限")
}
