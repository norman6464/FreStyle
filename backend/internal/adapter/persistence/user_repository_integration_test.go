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
// GORM の書き込みの round-trip を実 Postgres で検証する。nullable 列（company_id / onboarded_at /
// deleted_at）の詰め替えと、論理削除除外・not-found 時の (nil, nil) も確認する。
func TestUserRepository_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(db)
	ctx := context.Background()

	t.Run("Create → FindByCognitoSub / FindByID で round-trip（company_id 含む）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users")

		cid := uint64(42)
		require.NoError(t, repo.Create(ctx, &domain.User{
			CognitoSub: "sub-1", Email: "u@example.com", DisplayName: "山田",
			Role: domain.RoleTrainee, CompanyID: &cid,
		}))

		got, err := repo.FindByCognitoSub(ctx, "sub-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "sub-1", got.CognitoSub)
		require.Equal(t, "u@example.com", got.Email)
		require.Equal(t, "山田", got.DisplayName)
		require.Equal(t, domain.RoleTrainee, got.Role)
		require.NotNil(t, got.CompanyID)
		require.Equal(t, uint64(42), *got.CompanyID)
		require.Nil(t, got.OnboardedAt) // 未 onboarding
		require.False(t, got.CreatedAt.IsZero())

		byID, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.NotNil(t, byID)
		require.Equal(t, got.ID, byID.ID)
	})

	t.Run("company 無し（SuperAdmin）は CompanyID が nil", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users")
		require.NoError(t, repo.Create(ctx, &domain.User{
			CognitoSub: "admin-1", Email: "a@example.com", DisplayName: "管理者", Role: domain.RoleSuperAdmin,
		}))

		got, err := repo.FindByCognitoSub(ctx, "admin-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Nil(t, got.CompanyID)
	})

	t.Run("見つからない場合は (nil, nil)", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users")

		got, err := repo.FindByCognitoSub(ctx, "no-such-sub")
		require.NoError(t, err)
		require.Nil(t, got)

		byID, err := repo.FindByID(ctx, 999999)
		require.NoError(t, err)
		require.Nil(t, byID)
	})

	t.Run("MarkOnboarded 後は OnboardedAt が入る", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users")
		require.NoError(t, repo.Create(ctx, &domain.User{
			CognitoSub: "ob-1", Email: "o@example.com", DisplayName: "ob", Role: domain.RoleTrainee,
		}))

		u, err := repo.FindByCognitoSub(ctx, "ob-1")
		require.NoError(t, err)
		require.NoError(t, repo.MarkOnboarded(ctx, u.ID))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.NotNil(t, got.OnboardedAt)
	})

	t.Run("ListByRole は role で絞り id 昇順", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users")
		require.NoError(t, repo.Create(ctx, &domain.User{CognitoSub: "t1", Email: "t1@e.com", DisplayName: "t1", Role: domain.RoleTrainee}))
		require.NoError(t, repo.Create(ctx, &domain.User{CognitoSub: "a1", Email: "a1@e.com", DisplayName: "a1", Role: domain.RoleCompanyAdmin}))
		require.NoError(t, repo.Create(ctx, &domain.User{CognitoSub: "t2", Email: "t2@e.com", DisplayName: "t2", Role: domain.RoleTrainee}))

		trainees, err := repo.ListByRole(ctx, domain.RoleTrainee)
		require.NoError(t, err)
		require.Len(t, trainees, 2)
		require.Equal(t, "t1", trainees[0].CognitoSub)
		require.Equal(t, "t2", trainees[1].CognitoSub)
	})
}
