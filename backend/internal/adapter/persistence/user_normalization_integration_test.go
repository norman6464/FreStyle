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

// TestUserNormalization_Integration は users 正規化（FRESTYLE-311）の契約を実 Postgres で固定する。
// expand-contract の読み替えフェーズ（旧カラムは残し dual-write を維持、物理撤去は後続 PR）を対象にする:
//   - CreateWithOidcIdentity が users 行と identity を単一トランザクションで作る（片方だけ残らない）
//   - 旧 role 列 / cognito_sub 列へも併記する（dual-write。ローリング中に旧タスクが読むため必須の契約）
//   - role 名は roles.id へ解決して role_id に書き、読み出しは roles を JOIN して name を返す
//   - FindByCognitoSub は user_oidc_identities 経由で解決する（cognito_sub フォールバックは後続 PR で撤去）
//   - FK / CHECK / 部分 UNIQUE / CASCADE などの DB 制約
func TestUserNormalization_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(db)
	ctx := context.Background()

	truncate := func(t *testing.T) {
		t.Helper()
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")
	}

	t.Run("CreateWithOidcIdentity は role 名を role_id に解決し identity も作る", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n1@example.com", Name: "n1", Role: domain.RoleCompanyAdmin}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-1"))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, domain.RoleCompanyAdmin, got.Role)
		require.Equal(t, domain.RoleIDCompanyAdmin, got.RoleID)

		// identity が対で作られている。
		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("user_id = ? AND subject = ?", u.ID, "norm-1").Count(&count).Error)
		require.Equal(t, int64(1), count)

		// 旧 cognito_sub 列へも subject を併記する（ローリングデプロイ中に旧タスクが NULL を
		// Scan して 500 になるのを防ぐ dual-write。撤去は後続 PR）。
		var legacySub string
		require.NoError(t, db.Raw(`SELECT cognito_sub FROM users WHERE id = ?`, u.ID).Scan(&legacySub).Error)
		require.Equal(t, "norm-1", legacySub)
	})

	t.Run("未知の role 名の作成はエラー（黙って別ロールにしない・行も残さない）", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "x@example.com", Role: "no_such_role"}
		require.ErrorContains(
			t,
			repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-x"),
			"unknown role",
		)
		// role 解決はトランザクション前に失敗するので users 行は作られない。
		var count int64
		require.NoError(t, db.Model(&domain.User{}).Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("FindByCognitoSub は identities 経由で解決する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n2@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-2"))

		got, err := repo.FindByCognitoSub(ctx, "norm-2")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, u.ID, got.ID)
		require.Equal(t, domain.RoleTrainee, got.Role)
	})

	t.Run("EnsureOidcIdentity は冪等（同一 provider+subject を重複して作らない）", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n3@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-3"))

		// 作成時に張られた identity と同じものを張り直しても増えない。
		require.NoError(t, repo.EnsureOidcIdentity(ctx, u.ID, domain.OidcProviderCognito, "norm-3"))

		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("user_id = ?", u.ID).Count(&count).Error)
		require.Equal(t, int64(1), count)
	})

	t.Run("UpdateRole は role_id を更新し、旧 role 列へも併記する（dual-write 維持）", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n4@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-4"))

		require.NoError(t, repo.UpdateRole(ctx, u.ID, domain.RoleSuperAdmin))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, domain.RoleSuperAdmin, got.Role)
		require.Equal(t, domain.RoleIDSuperAdmin, got.RoleID)

		// ローリングデプロイ中に旧タスクが読む旧 role 列への併記を守る（撤去は後続 PR）。
		// これが欠けると、昇格が role_id だけ更新され旧タスクに trainee と誤認される。
		var legacyRole string
		require.NoError(t, db.Raw(`SELECT role FROM users WHERE id = ?`, u.ID).Scan(&legacyRole).Error)
		require.Equal(t, string(domain.RoleSuperAdmin), legacyRole)
	})

	t.Run("UpdateRole は未知の role 名を拒否する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n5@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-5"))
		require.ErrorContains(t, repo.UpdateRole(ctx, u.ID, "no_such_role"), "unknown role")
	})

	t.Run("BackfillUserNormalization は起動のたびに冪等に流せる（既存行を壊さない）", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "bf@example.com", Role: domain.RoleCompanyAdmin}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "bf-1"))

		// 何度流しても role_id / identity を壊さない（role_id を旧 role 基準で巻き戻さない）。
		require.NoError(t, database.BackfillUserNormalization(db))
		require.NoError(t, database.BackfillUserNormalization(db))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, domain.RoleIDCompanyAdmin, got.RoleID)
		var idCount int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).Where("user_id = ?", u.ID).Count(&idCount).Error)
		require.Equal(t, int64(1), idCount)
	})

	t.Run("DB 制約: 存在しない role_id の INSERT は FK で拒否される", func(t *testing.T) {
		truncate(t)
		err := db.Exec(
			`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
			 VALUES ('fkx@example.com', 'x', 999, true, NOW(), NOW())`,
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
		u := &domain.User{Email: "c1@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "cascade-1"))

		require.NoError(t, db.Exec(`DELETE FROM users WHERE id = ?`, u.ID).Error)

		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("user_id = ?", u.ID).Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("DB 制約: アクティブ行の email 重複は部分 UNIQUE で拒否・論理削除後の再利用は可", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{Email: "dup@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "mail-1"))

		// 同じ email のアクティブ行は作れない（users 行の INSERT が失敗 → トランザクションごと巻き戻る）。
		dup := &domain.User{Email: "dup@example.com", Role: domain.RoleTrainee}
		require.ErrorContains(
			t,
			repo.CreateWithOidcIdentity(ctx, dup, domain.OidcProviderCognito, "mail-2"),
			"uq_users_email_active",
		)
		// identity も巻き戻っている（片方だけ残らない）。
		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("subject = ?", "mail-2").Count(&count).Error)
		require.Equal(t, int64(0), count)

		// 論理削除すればアクティブ行が消えるので同じ email で再登録できる（再招待のシナリオ）。
		require.NoError(t, repo.SoftDelete(ctx, u1.ID))
		dup2 := &domain.User{Email: "dup@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, dup2, domain.OidcProviderCognito, "mail-3"))
	})

	t.Run("DB 制約: identity の空 subject は CHECK で拒否され、users 行も巻き戻る", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "chk@example.com", Role: domain.RoleTrainee}
		require.ErrorContains(
			t,
			repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, ""),
			"ck_user_oidc_identities_not_empty",
		)
		var count int64
		require.NoError(t, db.Model(&domain.User{}).Where("email = ?", "chk@example.com").Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("DB 制約: role_id は NOT NULL（省略時は DEFAULT trainee が入る）", func(t *testing.T) {
		truncate(t)
		// role_id を省略した INSERT → DEFAULT trainee で通る（NOT NULL + DEFAULT の回帰）。
		require.NoError(t, db.Exec(
			`INSERT INTO users (email, name, is_active, created_at, updated_at)
			 VALUES ('def@example.com', 'd', true, NOW(), NOW())`,
		).Error)
		got, err := repo.FindActiveByEmail(ctx, "def@example.com")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, domain.RoleIDTrainee, got.RoleID)
		// 読み出しは roles を JOIN するので role 名も trainee で返る。
		require.Equal(t, domain.RoleTrainee, got.Role)
	})

	t.Run("作成時に別ユーザーの subject 占有はトランザクションごとエラーになる", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{Email: "own1@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "shared-sub"))

		// 同じ subject を別ユーザーに割り当てようとすると identity の一意制約で失敗し、
		// users 行ごと巻き戻る（孤児ユーザーを作らない）。
		u2 := &domain.User{Email: "own2@example.com", Role: domain.RoleTrainee}
		require.Error(t, repo.CreateWithOidcIdentity(ctx, u2, domain.OidcProviderCognito, "shared-sub"))
		var count int64
		require.NoError(t, db.Model(&domain.User{}).Where("email = ?", "own2@example.com").Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("SoftDelete は identity を削除して subject の占有を解く", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{Email: "re@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "reinvite-sub"))

		require.NoError(t, repo.SoftDelete(ctx, u1.ID))

		// identity（正規化後の突き合わせの正）は削除され、subject の占有が解ける。
		// これにより cognito_sub 撤去後（FRESTYLE-311 後続 PR）は同一 OIDC アカウントの再招待が成立する。
		var idCount int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("subject = ?", "reinvite-sub").Count(&idCount).Error)
		require.Equal(t, int64(0), idCount)

		// 論理削除された行は identity 経由の解決から外れる。
		got, err := repo.FindByCognitoSub(ctx, "reinvite-sub")
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("バックフィルは論理削除ユーザーの残置 identity を掃除する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "dead@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "dead-1"))

		// identity を残したまま論理削除の状態を作る（SoftDelete は identity も消すため、ここでは
		// deleted_at だけを直接立てて「掃除対象の残置 identity」を再現する）。
		require.NoError(t, db.Exec(`UPDATE users SET deleted_at = NOW() WHERE id = ?`, u.ID).Error)

		require.NoError(t, database.BackfillUserNormalization(db))

		var count int64
		require.NoError(t, db.Model(&domain.UserOidcIdentity{}).
			Where("user_id = ?", u.ID).Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("空 email はアクティブ行でも複数共存できる（部分 UNIQUE の対象外）", func(t *testing.T) {
		truncate(t)
		e1 := &domain.User{Email: "", Role: domain.RoleTrainee}
		e2 := &domain.User{Email: "", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, e1, domain.OidcProviderCognito, "nomail-1"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, e2, domain.OidcProviderCognito, "nomail-2"))
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

	t.Run("FindActiveByEmail はハッシュ込みで 1 件返し、無効・削除行は除外する", func(t *testing.T) {
		truncate(t)
		hash := "$2a$10$Xgxiol1/CKW0E2qp4P3JOO/fZp3dcDmXxMHk76rHrOLRec8RIaqEm"
		u := &domain.User{Email: "find@example.com", Role: domain.RoleTrainee, PasswordHash: &hash}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "mail-find-1"))

		got, err := repo.FindActiveByEmail(ctx, "find@example.com")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.PasswordHash)
		require.Equal(t, hash, *got.PasswordHash)

		// 無効化すると引けない。
		require.NoError(t, repo.UpdateActive(ctx, u.ID, false))
		got, err = repo.FindActiveByEmail(ctx, "find@example.com")
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("FindActiveByEmail は email 重複（index 未作成環境）で曖昧ログインを拒否する", func(t *testing.T) {
		truncate(t)
		// uq_users_email_active が作れない既存環境を再現するため一時的に index を落とす。
		require.NoError(t, db.Exec(`DROP INDEX IF EXISTS uq_users_email_active`).Error)
		defer func() {
			require.NoError(t, database.ApplyUserNormalizationConstraints(db))
		}()

		for _, sub := range []string{"dup-a", "dup-b"} {
			u := &domain.User{Email: "dup2@example.com", Role: domain.RoleTrainee}
			require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		}

		_, err := repo.FindActiveByEmail(ctx, "dup2@example.com")
		require.ErrorContains(t, err, "重複を解消")
	})
}
