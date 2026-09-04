//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestUserNormalization_Integration は users 正規化（FRESTYLE-311）の契約を実 Postgres で固定する。
// 旧カラム（users.role / users.cognito_sub）撤去（migrations/0021、および users.role 列自体の
// 撤去）後の world を対象にする:
//   - CreateWithOidcIdentity が users 行と identity を単一トランザクションで作る（片方だけ残らない）
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

	// recreateUniqueEmailActiveIndex は uq_users_email_active を schema.hcl と同じ定義で張り直す。
	// 索引を意図的に落とすテスト（重複が index 未作成環境を再現する）専用の後始末で、
	// database.ApplySchema は「まだ何も無い空の DB」専用（IF NOT EXISTS を持たない）なので
	// ここでは使えない。
	recreateUniqueEmailActiveIndex := func(t *testing.T, db *sql.DB) {
		t.Helper()
		_, err := db.Exec(
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email_active
			   ON users (lower(btrim(email, E'\t\n\x0B\f\r ')))
			 WHERE deleted_at IS NULL AND btrim(email, E'\t\n\x0B\f\r ') <> ''`,
		)
		require.NoError(t, err)
	}

	t.Run("CreateWithOidcIdentity は users 行と identity を対で作る", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n1@example.com", Name: "n1"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-1"))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "n1", got.Name)

		// identity が対で作られている。
		var count int64
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1 AND subject = $2`,
			u.ID, "norm-1",
		).Scan(&count))
		require.Equal(t, int64(1), count)
	})

	t.Run("FindByCognitoSub は identities 経由で解決する", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n2@example.com"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-2"))

		got, err := repo.FindByCognitoSub(ctx, "norm-2")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, u.ID, got.ID)
	})

	t.Run("EnsureOidcIdentity は冪等（同一 provider+subject を重複して作らない）", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "n3@example.com"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "norm-3"))

		// 作成時に張られた identity と同じものを張り直しても増えない。
		require.NoError(t, repo.EnsureOidcIdentity(ctx, u.ID, domain.OidcProviderCognito, "norm-3"))

		var count int64
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1`, u.ID,
		).Scan(&count))
		require.Equal(t, int64(1), count)
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
		u := &domain.User{Email: "c1@example.com"}
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
		u1 := &domain.User{Email: "dup@example.com"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "mail-1"))

		// 同じ email のアクティブ行は作れない（users 行の INSERT が失敗 → トランザクションごと巻き戻る）。
		dup := &domain.User{Email: "dup@example.com"}
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
		dup2 := &domain.User{Email: "dup@example.com"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, dup2, domain.OidcProviderCognito, "mail-3"))
	})

	t.Run("DB 制約: identity の空 subject は CHECK で拒否され、users 行も巻き戻る", func(t *testing.T) {
		truncate(t)
		u := &domain.User{Email: "chk@example.com"}
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

	t.Run("作成時に別ユーザーの subject 占有はトランザクションごとエラーになる", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{Email: "own1@example.com"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "shared-sub"))

		// 同じ subject を別ユーザーに割り当てようとすると identity の一意制約で失敗し、
		// users 行ごと巻き戻る（孤児ユーザーを作らない）。
		u2 := &domain.User{Email: "own2@example.com"}
		require.Error(t, repo.CreateWithOidcIdentity(ctx, u2, domain.OidcProviderCognito, "shared-sub"))
		var count int64
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM users WHERE email = $1`, "own2@example.com",
		).Scan(&count))
		require.Equal(t, int64(0), count)
	})

	t.Run("SoftDelete が identity を解放し、同じ subject で再招待できる", func(t *testing.T) {
		truncate(t)
		u1 := &domain.User{Email: "re@example.com"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "reinvite-sub"))

		require.NoError(t, repo.SoftDelete(ctx, u1.ID))

		// 旧 cognito_sub 列が撤去されたので、同じ email / 同じ subject で新ユーザーを作れる
		// （identity は SoftDelete で解放され、cognito_sub のユニーク衝突も無くなった）。
		u2 := &domain.User{Email: "re@example.com"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u2, domain.OidcProviderCognito, "reinvite-sub"))

		got, err := repo.FindByCognitoSub(ctx, "reinvite-sub")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, u2.ID, got.ID)
	})

	t.Run("空 email はアクティブ行でも複数共存できる（部分 UNIQUE の対象外）", func(t *testing.T) {
		truncate(t)
		e1 := &domain.User{Email: ""}
		e2 := &domain.User{Email: ""}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, e1, domain.OidcProviderCognito, "nomail-1"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, e2, domain.OidcProviderCognito, "nomail-2"))
	})

	t.Run("FindActiveByEmail はハッシュ込みで 1 件返し、無効・削除行は除外する", func(t *testing.T) {
		truncate(t)
		hash := "$2a$10$Xgxiol1/CKW0E2qp4P3JOO/fZp3dcDmXxMHk76rHrOLRec8RIaqEm"
		u := &domain.User{Email: "find@example.com", PasswordHash: &hash}
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
			// このサブテストが残す重複行を消してから張り直す（重複が残ったままだと
			// CREATE UNIQUE INDEX 自体が失敗する。宣言的スキーマは黙って作らず失敗を選ぶ）。
			truncate(t)
			recreateUniqueEmailActiveIndex(t, sqlDB)
		}()

		for _, sub := range []string{"dup-a", "dup-b"} {
			u := &domain.User{Email: "dup2@example.com"}
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
		recreateUniqueEmailActiveIndex(t, sqlDB)
		u1 := &domain.User{Email: "case@example.com"}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u1, domain.OidcProviderCognito, "case-1"))

		dup := &domain.User{Email: "CASE@Example.com"}
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
			`INSERT INTO users (email, name, is_active, created_at, updated_at)
			 VALUES ('Legacy@Example.com', 'legacy', true, NOW(), NOW())`,
		)
		require.NoError(t, err)

		got, err := repo.FindActiveByEmail(ctx, "legacy@example.com")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "Legacy@Example.com", got.Email)
	})
}
