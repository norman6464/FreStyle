//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestUserNormalization_Integration は users 正規化（FRESTYLE-311）の契約を実 Postgres で固定する。
// 旧カラム（users.role / users.cognito_sub）撤去（migrations/0021）後の world を対象にする:
//   - CreateWithOidcIdentity が users 行と identity を単一トランザクションで作る（片方だけ残らない）
//   - role 名は roles.id へ解決して role_id に書き、読み出しは roles を JOIN して name を返す
//   - FindByCognitoSub は user_oidc_identities 経由でのみ解決する
//   - FK / CHECK / 部分 UNIQUE / CASCADE などの DB 制約
func TestUserNormalization_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()

	truncate := func(t *testing.T) {
		t.Helper()
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")
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
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1 AND subject = $2`,
			u.ID, "norm-1",
		).Scan(&count))
		require.Equal(t, int64(1), count)
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
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM users`).Scan(&count))
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
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1`, u.ID,
		).Scan(&count))
		require.Equal(t, int64(1), count)
	})

	t.Run("UpdateRole は role_id を更新する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n4@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-4"))

		require.NoError(t, repo.UpdateRole(ctx, u.ID, domain.RoleSuperAdmin))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, domain.RoleSuperAdmin, got.Role)
		require.Equal(t, domain.RoleIDSuperAdmin, got.RoleID)
	})

	t.Run("UpdateRole は未知の role 名を拒否する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n5@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-5"))
		require.ErrorContains(t, repo.UpdateRole(ctx, u.ID, "no_such_role"), "unknown role")
	})

	t.Run("正規化のバックフィルは起動のたびに冪等に流せる（既存行を壊さない）", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "bf@example.com", Role: domain.RoleCompanyAdmin}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "bf-1"))

		// 何度流しても role_id / identity を壊さない（role_id を旧 role 基準で巻き戻さない）。
		require.NoError(t, database.ApplyCoreSchema(ctx, sqlDB))
		require.NoError(t, database.ApplyCoreSchema(ctx, sqlDB))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, domain.RoleIDCompanyAdmin, got.RoleID)
		var idCount int64
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1`, u.ID,
		).Scan(&idCount))
		require.Equal(t, int64(1), idCount)
	})

	t.Run("DB 制約: 存在しない role_id の INSERT は FK で拒否される", func(t *testing.T) {
		truncate(t)
		_, err := sqlDB.Exec(
			`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
			 VALUES ('fkx@example.com', 'x', 999, true, NOW(), NOW())`,
		)
		require.ErrorContains(t, err, "fk_users_role")
	})

	t.Run("DB 制約: 存在しない user_id の identity は FK で拒否される", func(t *testing.T) {
		truncate(t)
		_, err := sqlDB.Exec(
			`INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
			 VALUES (424242, 'cognito', 'ghost', NOW(), NOW())`,
		)
		require.ErrorContains(t, err, "fk_user_oidc_identities_user")
	})

	t.Run("DB 制約: ユーザーの物理削除で identity が CASCADE 削除される", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "c1@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "cascade-1"))

		_, err := sqlDB.Exec(`DELETE FROM users WHERE id = $1`, u.ID)
		require.NoError(t, err)

		var count int64
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1`, u.ID,
		).Scan(&count))
		require.Equal(t, int64(0), count)
	})

	t.Run("DB 制約: アクティブ行の email 重複は部分 UNIQUE で拒否・論理削除後の再利用は可", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{Email: "dup@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "mail-1"))

		// 同じ email のアクティブ行は作れない（users 行の INSERT が失敗 → トランザクションごと巻き戻る）。
		dup := &domain.User{Email: "dup@example.com", Role: domain.RoleTrainee}
		require.ErrorIs(
			t,
			repo.CreateWithOidcIdentity(ctx, dup, domain.OidcProviderCognito, "mail-2"),
			repository.ErrEmailTaken,
		)
		// identity も巻き戻っている（片方だけ残らない）。
		var count int64
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE subject = $1`, "mail-2",
		).Scan(&count))
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
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM users WHERE email = $1`, "chk@example.com",
		).Scan(&count))
		require.Equal(t, int64(0), count)
	})

	t.Run("DB 制約: role_id は NOT NULL（省略時は DEFAULT trainee が入る）", func(t *testing.T) {
		truncate(t)
		// role_id を省略した INSERT → DEFAULT trainee で通る（NOT NULL + DEFAULT の回帰）。
		_, err := sqlDB.Exec(
			`INSERT INTO users (email, name, is_active, created_at, updated_at)
			 VALUES ('def@example.com', 'd', true, NOW(), NOW())`,
		)
		require.NoError(t, err)
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
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM users WHERE email = $1`, "own2@example.com",
		).Scan(&count))
		require.Equal(t, int64(0), count)
	})

	t.Run("SoftDelete が identity を解放し、同じ subject で再招待できる", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{Email: "re@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "reinvite-sub"))

		require.NoError(t, repo.SoftDelete(ctx, u1.ID))

		// 旧 cognito_sub 列が撤去されたので、同じ email / 同じ subject で新ユーザーを作れる
		// （identity は SoftDelete で解放され、cognito_sub のユニーク衝突も無くなった）。
		u2 := &domain.User{Email: "re@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u2, domain.OidcProviderCognito, "reinvite-sub"))

		got, err := repo.FindByCognitoSub(ctx, "reinvite-sub")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, u2.ID, got.ID)
	})

	t.Run("バックフィルは論理削除ユーザーの残置 identity を掃除する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "dead@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "dead-1"))

		// identity を残したまま論理削除の状態を作る（SoftDelete は identity も消すため、ここでは
		// deleted_at だけを直接立てて「掃除対象の残置 identity」を再現する）。
		_, err := sqlDB.Exec(`UPDATE users SET deleted_at = NOW() WHERE id = $1`, u.ID)
		require.NoError(t, err)

		require.NoError(t, database.ApplyCoreSchema(ctx, sqlDB))

		var count int64
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1`, u.ID,
		).Scan(&count))
		require.Equal(t, int64(0), count)
	})

	t.Run("空 email はアクティブ行でも複数共存できる（部分 UNIQUE の対象外）", func(t *testing.T) {
		truncate(t)
		e1 := &domain.User{Email: "", Role: domain.RoleTrainee}
		e2 := &domain.User{Email: "", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, e1, domain.OidcProviderCognito, "nomail-1"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, e2, domain.OidcProviderCognito, "nomail-2"))
	})

	t.Run("スキーマ DDL を再適用しても NOT_NULL と DEFAULT が剥がれない", func(t *testing.T) {
		// 回帰テスト: 起動のたびに流れる DDL が既存列を作り直したり緩めたりすると、
		// ローリングデプロイ中に安全弁（role_id の NOT NULL / DEFAULT）が消える。
		require.NoError(t, database.ApplyCoreSchema(ctx, sqlDB))

		var isNullable string
		var columnDefault *string
		require.NoError(t, sqlDB.QueryRow(
			`SELECT is_nullable, column_default FROM information_schema.columns
			 WHERE table_name = 'users' AND column_name = 'role_id'`,
		).Scan(&isNullable, &columnDefault))
		require.Equal(t, "NO", isNullable, "role_id の NOT NULL が DDL 再適用で剥がれた")
		require.NotNil(t, columnDefault, "role_id の DEFAULT が DDL 再適用で剥がれた")
		require.Contains(t, *columnDefault, "3")

		var emailNullable string
		require.NoError(t, sqlDB.QueryRow(
			`SELECT is_nullable FROM information_schema.columns
			 WHERE table_name = 'users' AND column_name = 'email'`,
		).Scan(&emailNullable))
		require.Equal(t, "NO", emailNullable, "email の NOT NULL が DDL 再適用で剥がれた")
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
		_, err := sqlDB.Exec(`DROP INDEX IF EXISTS uq_users_email_active`)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, database.ApplyCoreSchema(ctx, sqlDB))
		}()

		for _, sub := range []string{"dup-a", "dup-b"} {
			u := &domain.User{Email: "dup2@example.com", Role: domain.RoleTrainee}
			require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		}

		_, err = repo.FindActiveByEmail(ctx, "dup2@example.com")
		require.ErrorContains(t, err, "重複を解消")
	})

	// 一意索引のキーは domain.NormalizeEmail と同じ正規形 lower(btrim(email, ...))。アプリは
	// 畳んだ値を保存するが、索引が生の byte 一致だと「畳めば同じだがバイトが違う」2 行が
	// 両方作れてしまう。
	t.Run("DB 制約: 大小文字だけ違う email もアクティブ行の重複として拒否する", func(t *testing.T) {
		truncate(t)
		// 直前のサブテストが重複行を残したまま index を落としている場合に備えて張り直す。
		require.NoError(t, database.ApplyCoreSchema(ctx, sqlDB))
		u1 := &domain.User{Email: "case@example.com", Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "case-1"))

		dup := &domain.User{Email: "CASE@Example.com", Role: domain.RoleTrainee}
		require.ErrorIs(
			t,
			repo.CreateWithOidcIdentity(ctx, dup, domain.OidcProviderCognito, "case-2"),
			repository.ErrEmailTaken,
		)
	})

	// FindActiveByEmail の突き合わせも索引と同じ正規形の式。保存値が正規化される前に作られた
	// 大文字混じりの既存行も、同じアドレスとして 1 件に解決できる。
	t.Run("FindActiveByEmail は大小文字を無視して引く", func(t *testing.T) {
		truncate(t)
		// 正規化前の既存行を再現するため、アプリを通さず直接 INSERT する。
		_, err := sqlDB.Exec(
			`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
			 VALUES ('Legacy@Example.com', 'legacy', 3, true, NOW(), NOW())`,
		)
		require.NoError(t, err)

		got, err := repo.FindActiveByEmail(ctx, "legacy@example.com")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "Legacy@Example.com", got.Email)
	})

	// 既存データへの影響: 畳んだ値に重複がある環境では張り替えを見送り、旧索引（生の email）を
	// 残す。先に落とすと「新しい索引を作れない かつ 旧索引も無い」= 無防備な状態になるため。
	t.Run("索引の張り替え: 畳んだ値に重複がある間は旧索引を残し、解消後に張り替える", func(t *testing.T) {
		truncate(t)
		// 張り替え前（生の email キー）の環境を再現する。
		_, err := sqlDB.Exec(`DROP INDEX IF EXISTS uq_users_email_active`)
		require.NoError(t, err)
		_, err = sqlDB.Exec(
			`CREATE UNIQUE INDEX uq_users_email_active ON users (email)
			 WHERE deleted_at IS NULL AND email <> ''`,
		)
		require.NoError(t, err)
		// 旧索引だから作れてしまう「畳めば同じ」2 行（既存本番で起こり得る状態）。
		_, err = sqlDB.Exec(
			`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
			 VALUES ('mix@example.com', 'a', 3, true, NOW(), NOW()),
			        ('MIX@Example.com', 'b', 3, true, NOW(), NOW())`,
		)
		require.NoError(t, err)

		indexdef := func(t *testing.T) string {
			t.Helper()
			var def string
			require.NoError(t, sqlDB.QueryRow(
				`SELECT indexdef FROM pg_indexes WHERE indexname = 'uq_users_email_active'`,
			).Scan(&def))
			return def
		}

		// 起動時の制約適用は落ちず（WARNING のみ）、旧索引がそのまま残る。
		require.NoError(t, database.ApplyCoreSchema(ctx, sqlDB))
		require.NotContains(t, indexdef(t), "btrim")
		require.Contains(t, indexdef(t), "email")

		// 重複を解消すれば、次の起動で正規形（lower(btrim(email, ...))）へ張り替わる。
		_, err = sqlDB.Exec(`DELETE FROM users WHERE email = 'MIX@Example.com'`)
		require.NoError(t, err)
		require.NoError(t, database.ApplyCoreSchema(ctx, sqlDB))
		require.Contains(t, indexdef(t), "btrim")
	})
}
