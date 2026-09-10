//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestUserRepository_Integration は sqlc 化した読み取り（FindByOidcSubject / FindByID）と
// 書き込みの round-trip を実 Postgres で検証する。nullable 列（workspace_id / deleted_at）の
// 詰め替えと、論理削除除外・not-found 時の (nil, nil) も確認する。
func TestUserRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	oidcRepo := persistence.NewUserOidcIdentityRepository(sqlDB)
	ctx := context.Background()

	t.Run("Create + EnsureIdentity → FindByOidcSubject / FindByID で round-trip（workspace_id 含む）", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, append([]string{"user_oidc_identities"}, workspaceWriteTables...)...)
		ws := uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws, "ワークスペース 42", true)
		wid := ws.String()

		u := &domain.User{
			Email: "u@example.com", Name: "山田",
			WorkspaceID: &wid,
		}
		require.NoError(t, repo.Create(ctx, u))
		require.NoError(t, oidcRepo.EnsureIdentity(ctx, u.ID, domain.OidcProviderDefault, "sub-1"))

		got, err := repo.FindByOidcSubject(ctx, "sub-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "u@example.com", got.Email)
		require.Equal(t, "山田", got.Name)
		require.NotNil(t, got.WorkspaceID)
		require.Equal(t, wid, *got.WorkspaceID)
		require.False(t, got.CreatedAt.IsZero())

		byID, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.NotNil(t, byID)
		require.Equal(t, got.ID, byID.ID)
	})

	t.Run("ワークスペース無しは WorkspaceID が nil", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")
		u := &domain.User{Email: "a@example.com", Name: "管理者"}
		require.NoError(t, repo.Create(ctx, u))
		require.NoError(t, oidcRepo.EnsureIdentity(ctx, u.ID, domain.OidcProviderDefault, "admin-1"))

		got, err := repo.FindByOidcSubject(ctx, "admin-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Nil(t, got.WorkspaceID)
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
}
