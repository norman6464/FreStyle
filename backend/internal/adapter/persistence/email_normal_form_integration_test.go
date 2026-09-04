//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// emailNormalExprSQL は users の一意索引・検索が使う正規形の式。
// domain.NormalizeEmail の SQL 版で、落とす空白は domain.EmailTrimCutset と同じ集合。
const emailNormalExprSQL = `SELECT lower(btrim($1::text, E'\t\n\x0B\f\r '))`

func TestEmailNormalForm_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()

	truncate := func(t *testing.T) {
		t.Helper()
		testsupport.TruncateAll(t, sqlDB, "users", "user_oidc_identities")
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
			`INSERT INTO users (email, name, is_active, created_at, updated_at)
			 VALUES ('  Pad@Example.com'||chr(9), 'pad', true, NOW(), NOW())`,
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
			`INSERT INTO users (email, name, is_active, created_at, updated_at)
			 VALUES ('  space@example.com  ', 'space', true, NOW(), NOW())`,
		)
		require.NoError(t, err)

		dup := &domain.User{Email: "space@example.com"}
		require.ErrorIs(
			t,
			repo.CreateWithOidcIdentity(ctx, dup, domain.OidcProviderCognito, "space-1"),
			repository.ErrEmailTaken,
		)
	})
}
