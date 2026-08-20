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
		require.Equal(t, string(domain.RoleSuperAdmin), legacyRole)
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

	t.Run("DB 制約: 存在しない role_id の INSERT は FK で拒否される", func(t *testing.T) {
		truncate(t)
		err := db.Exec(
			`INSERT INTO users (cognito_sub, email, name, role, role_id, is_active, created_at, updated_at)
			 VALUES ('fk-x', 'fkx@example.com', 'x', 'trainee', 999, true, NOW(), NOW())`,
		).Error
		require.ErrorContains(t, err, "fk_users_role")
	})

	t.Run("DB 制約: 存在しない user_id の identity は FK で拒否される", func(t *testing.T) {
		truncate(t)
		err := db.Exec(
			`INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
			 VALUES (424242, 'cognito', 'ghost', NOW(), NOW())`,
		).Error
		require.ErrorContains(t, err, "fk_user_oidc_identities_user")
	})

	t.Run("DB 制約: ユーザーの物理削除で identity が CASCADE 削除される", func(t *testing.T) {
		truncate(t)
		u := &domain.User{CognitoSub: "cascade-1", Email: "c1@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u))
		require.NoError(t, repo.EnsureOidcIdentity(ctx, u.ID, domain.OidcProviderCognito, "cascade-1"))

		require.NoError(t, db.Exec(`DELETE FROM users WHERE id = ?`, u.ID).Error)

		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("user_id = ?", u.ID).Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("DB 制約: アクティブ行の email 重複は部分 UNIQUE で拒否・論理削除後の再利用は可", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{CognitoSub: "mail-1", Email: "dup@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u1))

		dup := &domain.User{CognitoSub: "mail-2", Email: "dup@example.com", Role: domain.RoleTrainee}
		require.ErrorContains(t, repo.Create(ctx, dup), "uq_users_email_active")

		// 論理削除すればアクティブ行が消えるので同じ email で再登録できる（再招待のシナリオ）。
		require.NoError(t, repo.SoftDelete(ctx, u1.ID))
		dup2 := &domain.User{CognitoSub: "mail-3", Email: "dup@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, dup2))
	})

	t.Run("DB 制約: identity の空 subject は CHECK で拒否される", func(t *testing.T) {
		truncate(t)
		u := &domain.User{CognitoSub: "chk-1", Email: "chk@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u))
		err := repo.EnsureOidcIdentity(ctx, u.ID, domain.OidcProviderCognito, "")
		require.ErrorContains(t, err, "ck_user_oidc_identities_not_empty")
	})

	t.Run("DB 制約: role_id は NOT NULL（省略時は DEFAULT trainee が入る）", func(t *testing.T) {
		truncate(t)
		// 旧コード相当の INSERT（role_id 省略）→ DEFAULT trainee で通る（ローリングデプロイ保全）。
		require.NoError(t, db.Exec(
			`INSERT INTO users (cognito_sub, email, name, role, is_active, created_at, updated_at)
			 VALUES ('def-1', 'def@example.com', 'd', 'company_admin', true, NOW(), NOW())`,
		).Error)
		got, err := repo.FindByCognitoSub(ctx, "def-1")
		require.NoError(t, err)
		require.Equal(t, domain.RoleIDTrainee, got.RoleID)
		// 読み出しは移行期間中「旧カラムが正」なので、role_id が既定値でも正しいロールが返る
		// （混在ウィンドウで旧コードが作った company_admin が trainee 扱いにならない）。
		require.Equal(t, domain.RoleCompanyAdmin, got.Role)

		// バックフィル（起動時に毎回実行）が role 文字列との不一致を修復する。
		require.NoError(t, database.BackfillUserNormalization(db))
		got, err = repo.FindByCognitoSub(ctx, "def-1")
		require.NoError(t, err)
		require.Equal(t, domain.RoleIDCompanyAdmin, got.RoleID)
		require.Equal(t, domain.RoleCompanyAdmin, got.Role)
	})

	t.Run("EnsureOidcIdentity は別ユーザーの subject 占有をエラーにする", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{CognitoSub: "own-1", Email: "own1@example.com", Role: domain.RoleTrainee}
		u2 := &domain.User{CognitoSub: "own-2", Email: "own2@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u1))
		require.NoError(t, repo.Create(ctx, u2))
		require.NoError(t, repo.EnsureOidcIdentity(ctx, u1.ID, domain.OidcProviderCognito, "shared-sub"))

		err := repo.EnsureOidcIdentity(ctx, u2.ID, domain.OidcProviderCognito, "shared-sub")
		require.ErrorContains(t, err, "oidc identity conflict")
	})

	t.Run("SoftDelete が identity を解放し、同じ subject で再招待できる", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{CognitoSub: "reinvite-old", Email: "re@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u1))
		require.NoError(t, repo.EnsureOidcIdentity(ctx, u1.ID, domain.OidcProviderCognito, "reinvite-sub"))

		require.NoError(t, repo.SoftDelete(ctx, u1.ID))

		// identity が消え、別ユーザーが同じ subject を取れる（同一 OIDC アカウントの再招待）。
		u2 := &domain.User{CognitoSub: "reinvite-new", Email: "re@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, u2))
		require.NoError(t, repo.EnsureOidcIdentity(ctx, u2.ID, domain.OidcProviderCognito, "reinvite-sub"))

		got, err := repo.FindByCognitoSub(ctx, "reinvite-sub")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, u2.ID, got.ID)
	})

	t.Run("バックフィルは論理削除ユーザーの identity を作らず、残置も掃除する", func(t *testing.T) {
		truncate(t)
		// 旧コード相当: 論理削除済みの行（cognito_sub あり）
		require.NoError(t, db.Exec(
			`INSERT INTO users (cognito_sub, email, name, role, is_active, created_at, updated_at, deleted_at)
			 VALUES ('dead-1', 'dead@example.com', 'dead', 'trainee', true, NOW(), NOW(), NOW())`,
		).Error)
		require.NoError(t, database.BackfillUserNormalization(db))
		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("subject = ?", "dead-1").Count(&count).Error)
		require.Equal(t, int64(0), count)

		// 残置 identity（過去データ相当）も掃除される。
		var deadID uint64
		require.NoError(t, db.Raw(`SELECT id FROM users WHERE cognito_sub = 'dead-1'`).Scan(&deadID).Error)
		require.NoError(t, db.Exec(
			`INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
			 VALUES (?, 'cognito', 'dead-1', NOW(), NOW())`, deadID,
		).Error)
		require.NoError(t, database.BackfillUserNormalization(db))
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("subject = ?", "dead-1").Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("空 email はアクティブ行でも複数共存できる（部分 UNIQUE の対象外）", func(t *testing.T) {
		truncate(t)
		e1 := &domain.User{CognitoSub: "nomail-1", Email: "", Role: domain.RoleTrainee}
		e2 := &domain.User{CognitoSub: "nomail-2", Email: "", Role: domain.RoleTrainee}
		require.NoError(t, repo.Create(ctx, e1))
		require.NoError(t, repo.Create(ctx, e2))
	})

	t.Run("AutoMigrate を再実行しても NOT_NULL と DEFAULT が剥がれない", func(t *testing.T) {
		// 回帰テスト: gorm タグと DB 状態が一致していないと、AutoMigrate が毎起動
		// DROP NOT NULL / DROP DEFAULT を発行し、ローリングデプロイ中に安全弁が消える。
		require.NoError(t, database.AutoMigrateAll(db))

		var row struct {
			IsNullable    string
			ColumnDefault *string
		}
		require.NoError(t, db.Raw(
			`SELECT is_nullable, column_default FROM information_schema.columns
			 WHERE table_name = 'users' AND column_name = 'role_id'`,
		).Scan(&row).Error)
		require.Equal(t, "NO", row.IsNullable, "role_id の NOT NULL が AutoMigrate 再実行で剥がれた")
		require.NotNil(t, row.ColumnDefault, "role_id の DEFAULT が AutoMigrate 再実行で剥がれた")
		require.Contains(t, *row.ColumnDefault, "3")

		var emailNullable string
		require.NoError(t, db.Raw(
			`SELECT is_nullable FROM information_schema.columns
			 WHERE table_name = 'users' AND column_name = 'email'`,
		).Scan(&emailNullable).Error)
		require.Equal(t, "NO", emailNullable, "email の NOT NULL が AutoMigrate 再実行で剥がれた")
	})
}
