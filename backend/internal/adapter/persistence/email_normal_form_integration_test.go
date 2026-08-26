//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// emailNormalExprSQL は users の一意索引・検索・招待の照会が使う正規形の式。
// domain.NormalizeEmail の SQL 版で、落とす空白は domain.EmailTrimCutset と同じ集合。
const emailNormalExprSQL = `SELECT lower(btrim(?::text, E'\t\n\x0B\f\r '))`

// TestEmailNormalForm_Integration は「アプリの正規形（domain.NormalizeEmail）」と
// 「DB 側の正規形（lower(btrim(email, EmailTrimCutset))）」が同じ 1 つの形であることを
// 実 PostgreSQL で固定する。
//
// 2 つがずれると、前後に空白の付いた既存行が正規形では引けない（招待したのに招待が無い・
// ログインできない）一方で、一意索引はその行と正規形の行を別キーとして両方通してしまい、
// 同じ人の行が 2 つできる。式を揃えるだけでは不十分で、保存側（招待の作成）と
// 既存行のバックフィルも同じ形に揃っている必要がある。
func TestEmailNormalForm_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewUserRepository(sqlDB)
	invRepo := persistence.NewAdminInvitationRepository(sqlDB)
	ctx := context.Background()

	truncate := func(t *testing.T) {
		t.Helper()
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities", "invitations")
	}

	// Go と SQL が同じ入力を同じ値へ畳むこと。片方だけを直したときにここが落ちる。
	// 大小文字の畳み方そのもの（U+212A / U+017F など）は lower() のロケール実装に依存するため
	// ここでは扱わない（Go 側の畳み方は domain の単体テストが固定している）。ここで固定したいのは
	// 今回ずれていた軸、つまり「前後の空白として何を落とすか」。
	t.Run("正規形は Go と SQL で一致する", func(t *testing.T) {
		inputs := []string{
			"ops@example.com",
			"OPS@Example.com",
			"  ops@example.com\t",
			"\r\n\v\fops@example.com\v\f\r\n",
			"",
			"   ",
			// ASCII 空白以外の Unicode 空白はどちらも落とさない（btrim の文字集合に無く、
			// Go 側も EmailTrimCutset しか落とさない）。落とさないこと自体が一致の条件で、
			// Go 側を strings.TrimSpace（Unicode 空白まで落とす）へ戻すとここが落ちる。
			"\u00A0ops@example.com",
			"ops@example.com\u00A0",
			"\u0085ops@example.com",
			"\u3000ops@example.com",
		}
		for _, in := range inputs {
			var got string
			require.NoError(t, db.Raw(emailNormalExprSQL, in).Scan(&got).Error)
			require.Equalf(t, domain.NormalizeEmail(in), got,
				"入力 %q: SQL の正規形が domain.NormalizeEmail と一致しません", in)
		}
	})

	// 正規化が入る前に保存された「前後に空白の付いた」行。式が lower(email) だけだと引けない。
	t.Run("FindActiveByEmail は前後空白付きの既存行を正規形で引く", func(t *testing.T) {
		truncate(t)
		require.NoError(t, db.Exec(
			`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
			 VALUES ('  Pad@Example.com'||chr(9), 'pad', 3, true, NOW(), NOW())`,
		).Error)

		got, err := repo.FindActiveByEmail(ctx, "pad@example.com")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "  Pad@Example.com\t", got.Email)
	})

	// 一意索引のキーも同じ正規形。空白だけ違う 2 行を別キーとして通してはいけない。
	t.Run("DB 制約: 前後空白だけ違う email もアクティブ行の重複として拒否する", func(t *testing.T) {
		truncate(t)
		require.NoError(t, database.ApplyUserNormalizationConstraints(db))
		require.NoError(t, db.Exec(
			`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
			 VALUES ('  space@example.com  ', 'space', 3, true, NOW(), NOW())`,
		).Error)

		dup := &domain.User{Email: "space@example.com", Role: domain.RoleTrainee}
		require.ErrorContains(
			t,
			repo.CreateWithOidcIdentity(ctx, dup, domain.OidcProviderCognito, "space-1"),
			"uq_users_email_active",
		)
	})

	// 既存行そのものを正規形へ畳む。式を揃えるだけだと、旧索引しか張れていない環境では
	// 「' a@x.com' が在るのに 'a@x.com' も作れる」状態が残る。
	t.Run("BackfillUserNormalization は既存行の email を正規形へ畳む", func(t *testing.T) {
		truncate(t)
		require.NoError(t, db.Exec(
			`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
			 VALUES ('  Fold@Example.com  ', 'fold', 3, true, NOW(), NOW())`,
		).Error)

		require.NoError(t, database.BackfillUserNormalization(db))

		var email string
		require.NoError(t, db.Raw(`SELECT email FROM users WHERE name = 'fold'`).Scan(&email).Error)
		require.Equal(t, "fold@example.com", email)
	})

	// 畳むと別の行と同じアドレスになる行は触らない（別人かもしれない 2 行を勝手に寄せない）。
	// 残った衝突は索引の張り替えを見送らせ、ログインは曖昧として拒否される（fail closed）。
	t.Run("BackfillUserNormalization は畳むと衝突する行を触らず、索引の張り替えも見送る", func(t *testing.T) {
		truncate(t)
		// 正規形の索引がある限りこの 2 行は作れない。張り替え前の環境を再現するため一時的に落とす。
		require.NoError(t, db.Exec(`DROP INDEX IF EXISTS uq_users_email_active`).Error)
		defer func() {
			require.NoError(t, db.Exec(`DELETE FROM users`).Error)
			require.NoError(t, database.ApplyUserNormalizationConstraints(db))
		}()
		require.NoError(t, db.Exec(
			`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
			 VALUES (' clash@example.com ', 'a', 3, true, NOW(), NOW()),
			        ('Clash@Example.com', 'b', 3, true, NOW(), NOW())`,
		).Error)

		require.NoError(t, database.BackfillUserNormalization(db))

		var emails []string
		require.NoError(t, db.Raw(`SELECT email FROM users ORDER BY name`).Scan(&emails).Error)
		require.Equal(t, []string{" clash@example.com ", "Clash@Example.com"}, emails)

		// 衝突が残っている間は索引を作らない（起動は落とさず WARNING）。
		require.NoError(t, database.ApplyUserNormalizationConstraints(db))
		var indexdef string
		require.NoError(t, db.Raw(
			`SELECT COALESCE(max(indexdef), '') FROM pg_indexes WHERE indexname = 'uq_users_email_active'`,
		).Scan(&indexdef).Error)
		require.Empty(t, indexdef)

		// 衝突を解消すれば畳めるようになり、索引も張れる。
		require.NoError(t, db.Exec(`DELETE FROM users WHERE name = 'b'`).Error)
		require.NoError(t, database.BackfillUserNormalization(db))
		var email string
		require.NoError(t, db.Raw(`SELECT email FROM users WHERE name = 'a'`).Scan(&email).Error)
		require.Equal(t, "clash@example.com", email)
		require.NoError(t, database.ApplyUserNormalizationConstraints(db))
		require.NoError(t, db.Raw(
			`SELECT COALESCE(max(indexdef), '') FROM pg_indexes WHERE indexname = 'uq_users_email_active'`,
		).Scan(&indexdef).Error)
		require.Contains(t, indexdef, "btrim")
	})

	// 招待の照会も同じ正規形。ログイン時の招待ゲートは正規形の OIDC メールで引くため、
	// 空白付き・大文字混じりのまま入った pending 行が引けないと、招待した相手が拒否される。
	t.Run("FindPendingByEmail は前後空白・大文字混じりの既存行を正規形で引く", func(t *testing.T) {
		truncate(t)
		require.NoError(t, db.Exec(
			`INSERT INTO invitations (company_id, email, role, name, status, token, expires_at, created_at)
			 VALUES (1, '  Invited@Example.com  ', ?, 'inv', ?, 'tok-pad', ?, NOW())`,
			domain.RoleCompanyAdmin, domain.InvitationStatusPending,
			time.Now().UTC().Add(time.Hour),
		).Error)

		got, err := invRepo.FindPendingByEmail(ctx, "invited@example.com")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "  Invited@Example.com  ", got.Email)
	})

	// 保存側も正規形。ここが生のままだと、照会だけ揃えても「どちらの表現で入っているか」が
	// 行ごとにばらつき、同じアドレスの解釈が 2 つある状態が続く。
	t.Run("Create は招待の email を正規形で保存する", func(t *testing.T) {
		truncate(t)
		token := "tok-normalize"
		inv := &domain.AdminInvitation{
			CompanyID: 1,
			Email:     "  New@Example.com\t",
			Role:      domain.RoleCompanyAdmin,
			Name:      "new",
			Status:    domain.InvitationStatusPending,
			Token:     &token,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		require.NoError(t, invRepo.Create(ctx, inv))

		var stored string
		require.NoError(t, db.Raw(`SELECT email FROM invitations WHERE token = ?`, token).Scan(&stored).Error)
		require.Equal(t, "new@example.com", stored)
	})

	// 既存の pending 行も起動時に畳む（invitations に一意制約は無いので衝突判定は要らない）。
	t.Run("BackfillUserNormalization は保留中の招待の email も畳む", func(t *testing.T) {
		truncate(t)
		require.NoError(t, db.Exec(
			`INSERT INTO invitations (company_id, email, role, name, status, token, expires_at, created_at)
			 VALUES (1, ' Legacy@Example.com ', ?, 'legacy', ?, 'tok-legacy', ?, NOW()),
			        (1, ' Done@Example.com ', ?, 'done', ?, 'tok-done', ?, NOW())`,
			domain.RoleCompanyAdmin, domain.InvitationStatusPending, time.Now().UTC().Add(time.Hour),
			domain.RoleCompanyAdmin, domain.InvitationStatusAccepted, time.Now().UTC().Add(time.Hour),
		).Error)

		require.NoError(t, database.BackfillUserNormalization(db))

		var pending string
		require.NoError(t, db.Raw(`SELECT email FROM invitations WHERE token = 'tok-legacy'`).Scan(&pending).Error)
		require.Equal(t, "legacy@example.com", pending)
		// 受理済み（履歴）は触らない。ログインの突き合わせ対象は pending だけ。
		var accepted string
		require.NoError(t, db.Raw(`SELECT email FROM invitations WHERE token = 'tok-done'`).Scan(&accepted).Error)
		require.Equal(t, " Done@Example.com ", accepted)
	})
}
