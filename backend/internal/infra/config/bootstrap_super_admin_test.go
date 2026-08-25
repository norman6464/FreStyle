package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_BootstrapSuperAdminEmail_既定は空で前後空白を落とす は、招待免除のブートストラップが
// 「明示的に環境変数を設定したときだけ効く」ことと、打ち間違いで黙って無効化されないよう
// 前後の空白を落とすことを固定する。
func Test_BootstrapSuperAdminEmail_既定は空で前後空白を落とす(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want string
	}{
		{"未設定なら空(免除なし)", "", ""},
		{"前後の空白は落とす", "  ops@example.com\t", "ops@example.com"},
		{"そのまま保持", "ops@example.com", "ops@example.com"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://x/y")
			t.Setenv("BOOTSTRAP_SUPER_ADMIN_EMAIL", c.env)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, c.want, cfg.BootstrapSuperAdminEmail)
		})
	}
}
