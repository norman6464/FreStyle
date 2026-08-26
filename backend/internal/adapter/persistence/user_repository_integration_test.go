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

// TestUserRepository_Integration は sqlc 化した読み取り（FindByCognitoSub / FindByID / ListByRole）と
// GORM の書き込みの round-trip を実 Postgres で検証する。nullable 列（company_id / deleted_at）の
// 詰め替えと、論理削除除外・not-found 時の (nil, nil) も確認する。
func TestUserRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()

	t.Run("CreateWithOidcIdentity → FindByCognitoSub / FindByID で round-trip（company_id 含む）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")

		cid := uint64(42)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "u@example.com", Name: "山田",
			Role: domain.RoleTrainee, CompanyID: &cid,
		}, domain.OidcProviderCognito, "sub-1"))

		got, err := repo.FindByCognitoSub(ctx, "sub-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "u@example.com", got.Email)
		require.Equal(t, "山田", got.Name)
		require.Equal(t, domain.RoleTrainee, got.Role)
		require.NotNil(t, got.CompanyID)
		require.Equal(t, uint64(42), *got.CompanyID)
		require.False(t, got.CreatedAt.IsZero())

		byID, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.NotNil(t, byID)
		require.Equal(t, got.ID, byID.ID)
	})

	t.Run("company 無し（SuperAdmin）は CompanyID が nil", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "a@example.com", Name: "管理者", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "admin-1"))

		got, err := repo.FindByCognitoSub(ctx, "admin-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Nil(t, got.CompanyID)
	})

	t.Run("見つからない場合は (nil, nil)", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")

		got, err := repo.FindByCognitoSub(ctx, "no-such-sub")
		require.NoError(t, err)
		require.Nil(t, got)

		byID, err := repo.FindByID(ctx, 999999)
		require.NoError(t, err)
		require.Nil(t, byID)
	})

	t.Run("ListByRole は role で絞り id 昇順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{Email: "t1@e.com", Name: "t1", Role: domain.RoleTrainee}, domain.OidcProviderCognito, "t1"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{Email: "a1@e.com", Name: "a1", Role: domain.RoleCompanyAdmin}, domain.OidcProviderCognito, "a1"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{Email: "t2@e.com", Name: "t2", Role: domain.RoleTrainee}, domain.OidcProviderCognito, "t2"))

		trainees, err := repo.ListByRole(ctx, domain.RoleTrainee)
		require.NoError(t, err)
		require.Len(t, trainees, 2)
		require.Equal(t, "t1", trainees[0].Name)
		require.Equal(t, "t2", trainees[1].Name)
	})
}
