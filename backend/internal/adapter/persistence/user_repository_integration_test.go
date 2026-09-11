//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestUserRepository_Integration は sqlc 化した読み取り（FindByOidcSubject / FindByID）と
// 書き込みの round-trip を実 Postgres で検証する。nullable 列（deleted_at）の詰め替えと、
// 論理削除除外・not-found 時の (nil, nil) も確認する。
func TestUserRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	oidcRepo := persistence.NewUserOidcIdentityRepository(sqlDB)
	ctx := context.Background()

	t.Run("Create + EnsureIdentity → FindByOidcSubject / FindByID で round-trip", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")
		u := &domain.User{Email: "u@example.com", Name: "山田"}
		require.NoError(t, repo.Create(ctx, u))
		require.NoError(t, oidcRepo.EnsureIdentity(ctx, u.ID, domain.OidcProviderDefault, "sub-1"))

		got, err := repo.FindByOidcSubject(ctx, "sub-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "u@example.com", got.Email)
		require.Equal(t, "山田", got.Name)
		require.False(t, got.CreatedAt.IsZero())

		byID, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.NotNil(t, byID)
		require.Equal(t, got.ID, byID.ID)
	})

	t.Run("UpdateEmail は email だけを更新する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")
		u := &domain.User{Email: "", Name: "未検証だった人"}
		require.NoError(t, repo.Create(ctx, u))

		require.NoError(t, repo.UpdateEmail(ctx, u.ID, "verified@example.com"))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "verified@example.com", got.Email)
		require.Equal(t, "未検証だった人", got.Name, "name は変わらないこと")
	})

	t.Run("UpdateEmail は既に別のアクティブユーザーが使っている値だとErrEmailTakenを返す", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")
		taken := &domain.User{Email: "taken@example.com", Name: "先に取った人"}
		require.NoError(t, repo.Create(ctx, taken))
		u := &domain.User{Email: "", Name: "後から検証した人"}
		require.NoError(t, repo.Create(ctx, u))

		err := repo.UpdateEmail(ctx, u.ID, "taken@example.com")
		require.ErrorIs(t, err, repository.ErrEmailTaken)

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, "", got.Email, "失敗したので email は空のままのはず")
	})

	t.Run("UpdateEmail は存在しないユーザーにはErrNotFoundを返す", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")
		err := repo.UpdateEmail(ctx, 999999, "nobody@example.com")
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("見つからない場合は (nil, nil)", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")

		got, err := repo.FindByOidcSubject(ctx, "no-such-sub")
		require.NoError(t, err)
		require.Nil(t, got)

		byID, err := repo.FindByID(ctx, 999999)
		require.NoError(t, err)
		require.Nil(t, byID)
	})

	// FindDisplayByID は段 5 の表示統合の要。FindByID と違い退会・停止でも解決できることが
	// 本体（コメント・変更履歴の投稿者表示）を壊さない条件そのものなので、ここで固定する。
	t.Run("FindDisplayByIDはFindByIDと違い退会・停止していても解決する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities", "profiles")
		u := &domain.User{Email: "gone@example.com", Name: "退会した人"}
		require.NoError(t, repo.Create(ctx, u))
		_, err := sqlDB.ExecContext(ctx,
			`INSERT INTO profiles (user_id, bio, avatar_url, status_message, updated_at)
			 VALUES ($1, '', $2, $3, now())`,
			u.ID, "https://example.test/gone.png", "退会済み",
		)
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx,
			`UPDATE users SET status = 'deactivated', deleted_at = now() WHERE id = $1`, u.ID,
		)
		require.NoError(t, err)

		// FindByID（退会は除外する経路）は解決しない。
		byID, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Nil(t, byID, "退会済みは FindByID では除外される")

		// FindDisplayByID は解決する — 過去のコメント・変更履歴の投稿者表示を壊さないため。
		display, err := repo.FindDisplayByID(ctx, u.ID)
		require.NoError(t, err)
		require.NotNil(t, display)
		require.Equal(t, "退会した人", display.Name)
		require.Equal(t, "https://example.test/gone.png", display.AvatarURL)
		require.Equal(t, "退会済み", display.StatusMessage)

		display, err = repo.FindDisplayByID(ctx, 999999)
		require.NoError(t, err)
		require.Nil(t, display, "実在しない id は (nil, nil)")
	})
}
