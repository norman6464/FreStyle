package config

import (
	"os"
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
		{"空文字なら空(免除なし)", "", ""},
		{"前後の空白は落とす", "  ops@example.com\t", "ops@example.com"},
		{"そのまま保持", "ops@example.com", "ops@example.com"},
		// 小文字化は config の責務ではない（正規化は domain.NormalizeEmail が一手に引き受け、
		// 免除の比較はそこを通した値どうしで行う）。config が畳み始めたらここが落ちる。
		{"大小文字は config では畳まない", "OPS@Example.com", "OPS@Example.com"},
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

// 環境変数そのものが存在しない経路も固定する。上のテーブルの「空文字」は
// 「変数はあるが値が空」であって「変数が無い」とは別の状態で、os.LookupEnv で
// 存在を分岐する実装に変えると通る経路が変わる。この免除は「明示的に設定したときだけ効く」
// ことが安全性の根拠なので、既定（未設定）で免除が付かないことを直接押さえる。
func Test_BootstrapSuperAdminEmail_環境変数が無ければ免除しない(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x/y")
	// t.Setenv を先に呼んでおくと、テスト終了時に元の値へ復元される（Unsetenv しても戻る）。
	t.Setenv("BOOTSTRAP_SUPER_ADMIN_EMAIL", "ops@example.com")
	require.NoError(t, os.Unsetenv("BOOTSTRAP_SUPER_ADMIN_EMAIL"))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Empty(t, cfg.BootstrapSuperAdminEmail)
}
