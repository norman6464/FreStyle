//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// emailNormalExprSQL は users の一意索引・検索・招待の照会が使う正規形の式。
// domain.NormalizeEmail の SQL 版で、落とす空白は domain.EmailTrimCutset と同じ集合。
const emailNormalExprSQL = `SELECT lower(btrim($1::text, E'\t\n\x0B\f\r '))`

func TestEmailNormalForm_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	invRepo := persistence.NewAdminInvitationRepository(sqlDB)
	ctx := context.Background()

	truncate := func(t *testing.T) {
		t.Helper()
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities", "invitations")
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
			require.NoError(t, sqlDB.QueryRow(emailNormalExprSQL, in).Scan(&got))
			require.Equalf(t, domain.NormalizeEmail(in), got,
				"入力 %q: SQL の正規形が domain.NormalizeEmail と一致しません", in)
		}
	})

	// 正規化が入る前に保存された「前後に空白の付いた」行。式が lower(email) だけだと引けない。
	t.Run("FindActiveByEmail は前後空白付きの既存行を正規形で引く", func(t *testing.T) {
		truncate(t)
		_, err := sqlDB.Exec(
			`INSERT INTO users (email, name, role, is_active, created_at, updated_at)
			 VALUES ('  Pad@Example.com'||chr(9), 'pad', 'trainee', true, NOW(), NOW())`,
		)
		require.NoError(t, err)

		got, err := repo.FindActiveByEmail(ctx, "pad@example.com")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "  Pad@Example.com\t", got.Email)
	})

	// 一意索引のキーも同じ正規形。空白だけ違う 2 行を別キーとして通してはいけない。
	t.Run("DB 制約: 前後空白だけ違う email もアクティブ行の重複として拒否する", func(t *testing.T) {
		truncate(t)
		_, err := sqlDB.Exec(
			`INSERT INTO users (email, name, role, is_active, created_at, updated_at)
			 VALUES ('  space@example.com  ', 'space', 'trainee', true, NOW(), NOW())`,
		)
		require.NoError(t, err)

		dup := &domain.User{Email: "space@example.com", Role: domain.RoleTrainee}
		require.ErrorIs(
			t,
			repo.CreateWithOidcIdentity(ctx, dup, domain.OidcProviderCognito, "space-1"),
			repository.ErrEmailTaken,
		)
	})

	// 招待の照会も同じ正規形。ログイン時の招待ゲートは正規形の OIDC メールで引くため、
	// 空白付き・大文字混じりのまま入った pending 行が引けないと、招待した相手が拒否される。
	t.Run("FindPendingByEmail は前後空白・大文字混じりの既存行を正規形で引く", func(t *testing.T) {
		truncate(t)
		_, err := sqlDB.Exec(
			`INSERT INTO invitations (email, role, name, status, token, expires_at, created_at)
			 VALUES ('  Invited@Example.com  ', $1, 'inv', $2, 'tok-pad', $3, NOW())`,
			domain.RoleCompanyAdmin, domain.InvitationStatusPending,
			time.Now().UTC().Add(time.Hour),
		)
		require.NoError(t, err)

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
			Email:     "  New@Example.com\t",
			Role:      domain.RoleCompanyAdmin,
			Name:      "new",
			Status:    domain.InvitationStatusPending,
			Token:     &token,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		require.NoError(t, invRepo.Create(ctx, inv))

		var stored string
		require.NoError(t, sqlDB.QueryRow(`SELECT email FROM invitations WHERE token = $1`, token).Scan(&stored))
		require.Equal(t, "new@example.com", stored)
	})
}
