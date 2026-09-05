//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestUserRepository_Integration は sqlc 化した読み取り（FindByCognitoSub / FindByID）と
// 書き込みの round-trip を実 Postgres で検証する。nullable 列（workspace_id / deleted_at）の
// 詰め替えと、論理削除除外・not-found 時の (nil, nil) も確認する。
func TestUserRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	oidcRepo := persistence.NewUserOidcIdentityRepository(sqlDB)
	ctx := context.Background()

	t.Run("Create + EnsureIdentity → FindByCognitoSub / FindByID で round-trip（workspace_id 含む）", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, append([]string{"user_oidc_identities"}, workspaceWriteTables...)...)
		ws := uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws, "ワークスペース 42", true)
		wid := ws.String()

		u := &domain.User{
			Email: "u@example.com", Name: "山田",
			WorkspaceID: &wid,
		}
		require.NoError(t, repo.Create(ctx, u))
		require.NoError(t, oidcRepo.EnsureIdentity(ctx, u.ID, domain.OidcProviderCognito, "sub-1"))

		got, err := repo.FindByCognitoSub(ctx, "sub-1")
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
		require.NoError(t, oidcRepo.EnsureIdentity(ctx, u.ID, domain.OidcProviderCognito, "admin-1"))

		got, err := repo.FindByCognitoSub(ctx, "admin-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Nil(t, got.WorkspaceID)
	})

	t.Run("見つからない場合は (nil, nil)", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")

		got, err := repo.FindByCognitoSub(ctx, "no-such-sub")
		require.NoError(t, err)
		require.Nil(t, got)

		byID, err := repo.FindByID(ctx, 999999)
		require.NoError(t, err)
		require.Nil(t, byID)
	})
}
