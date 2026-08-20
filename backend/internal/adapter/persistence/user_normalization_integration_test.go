//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestUserNormalization_Integration は users 正規化（FRESTYLE-311 PR1）の契約を実 Postgres で固定する。
//   - Create が role 名を roles.id に解決して role_id を書く（旧カラム role へも併記 = ロールバック保全）
//   - FindByCognitoSub が user_oidc_identities 経由で解決できる（旧カラム無しでも到達）
//   - EnsureOidcIdentity の冪等性
//   - BackfillUserNormalization が旧カラムだけの行を正規化テーブルへ埋める
func TestUserNormalization_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(db)
	ctx := context.Background()

	truncate := func(t *testing.T) {
		t.Helper()
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")
	}

	t.Run("Create は role 名を role_id に解決して書く", func(t *testing.T) {
		truncate(t)
		u := &domain.User{CognitoSub: "norm-1", Email: "n1@example.com", Name: "n1", Role: domain.RoleCompanyAdmin}
		require.NoError(t, repo.Create(ctx, u))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, domain.RoleCompanyAdmin, got.Role)
		require.Equal(t, domain.RoleIDCompanyAdmin, got.RoleID)
	})

	t.Run("未知の role 名の Create はエラー（黙って別ロールにしない）", func(t *testing.T) {
		truncate(t)
		err := repo.Create(ctx, &domain.User{CognitoSub: "norm-x", Email: "x@example.com", Role: "no_such_role"})
		require.ErrorContains(t, err, "unknown role")
	})

	t.Run("FindByCognitoSub は identities 経由で解決できる（旧カラム無しでも）", func(t *testing.T) {
		truncate(t)
		u := &domain.User{CognitoSub: "norm-2", Email: "n2@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u))
		require.NoError(t, repo.EnsureOidcIdentity(ctx, u.ID, domain.OidcProviderCognito, "norm-2"))

		// 旧カラムを空にして「identities だけが正」の状態を作る（PR3 後の世界の先取り検証）。
		require.NoError(t, db.Exec(`UPDATE users SET cognito_sub = '' WHERE id = ?`, u.ID).Error)

		got, err := repo.FindByCognitoSub(ctx, "norm-2")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, u.ID, got.ID)
		require.Equal(t, domain.RoleTrainee, got.Role)
	})

	t.Run("EnsureOidcIdentity は冪等", func(t *testing.T) {
		truncate(t)
		u := &domain.User{CognitoSub: "norm-3", Email: "n3@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u))

		require.NoError(t, repo.EnsureOidcIdentity(ctx, u.ID, domain.OidcProviderCognito, "norm-3"))
		require.NoError(t, repo.EnsureOidcIdentity(ctx, u.ID, domain.OidcProviderCognito, "norm-3"))

		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("user_id = ?", u.ID).Count(&count).Error)
		require.Equal(t, int64(1), count)
	})

	t.Run("UpdateRole は role_id と旧カラムの両方を更新する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{CognitoSub: "norm-4", Email: "n4@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u))

		require.NoError(t, repo.UpdateRole(ctx, u.ID, domain.RoleSuperAdmin))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, domain.RoleSuperAdmin, got.Role)
		require.Equal(t, domain.RoleIDSuperAdmin, got.RoleID)

		// 旧カラムにも併記されている（ロールバック保全・PR3 で撤去）。
		var legacyRole string
		require.NoError(t, db.Raw(`SELECT role FROM users WHERE id = ?`, u.ID).Scan(&legacyRole).Error)
		require.Equal(t, domain.RoleSuperAdmin, legacyRole)
	})

	t.Run("UpdateRole は未知の role 名を拒否する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{CognitoSub: "norm-5", Email: "n5@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u))
		require.ErrorContains(t, repo.UpdateRole(ctx, u.ID, "no_such_role"), "unknown role")
	})

	t.Run("BackfillUserNormalization が旧カラムだけの行を埋める", func(t *testing.T) {
		truncate(t)
		// 旧コードが書いた形の行を再現（role 文字列 + cognito_sub のみ / role_id・identity 無し）。
		require.NoError(t, db.Exec(
			`INSERT INTO users (cognito_sub, email, name, role, is_active, created_at, updated_at)
			 VALUES ('legacy-1', 'l1@example.com', 'legacy', ?, true, NOW(), NOW())`,
			domain.RoleCompanyAdmin,
		).Error)
		// 未知 role の行は trainee に倒れることも検証する。
		require.NoError(t, db.Exec(
			`INSERT INTO users (cognito_sub, email, name, role, is_active, created_at, updated_at)
			 VALUES ('legacy-2', 'l2@example.com', 'legacy2', 'weird_role', true, NOW(), NOW())`,
		).Error)

		require.NoError(t, database.BackfillUserNormalization(db))

		got, err := repo.FindByCognitoSub(ctx, "legacy-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, domain.RoleCompanyAdmin, got.Role)
		require.Equal(t, domain.RoleIDCompanyAdmin, got.RoleID)

		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("subject = ?", "legacy-1").Count(&count).Error)
		require.Equal(t, int64(1), count)

		got2, err := repo.FindByCognitoSub(ctx, "legacy-2")
		require.NoError(t, err)
		require.NotNil(t, got2)
		require.Equal(t, domain.RoleIDTrainee, got2.RoleID)

		// 冪等: もう一度流しても identity が増えない。
		require.NoError(t, database.BackfillUserNormalization(db))
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("subject = ?", "legacy-1").Count(&count).Error)
		require.Equal(t, int64(1), count)
	})
}
