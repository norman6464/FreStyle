package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_LocalPasswordAuth_有効化は3条件すべてを要求する は、ローカル専用パスワードログインの
// 有効化が「LOCAL_PASSWORD_AUTH=truthy かつ 明示 APP_ENV=local かつ JWKS 未設定」の
// AND であることを固定する（staging=APP_ENV 未設定 で無効になること。JWKS は条件に含めない）。
func Test_LocalPasswordAuth_有効化は3条件すべてを要求する(t *testing.T) {
	// Load が DB 必須で落ちないよう最低限の DSN を与える。
	base := func(t *testing.T) {
		t.Helper()
		t.Setenv("DATABASE_URL", "postgres://x/y")
	}

	cases := []struct {
		name   string
		flag   string
		appEnv string
		jwks   string
		want   bool
	}{
		{"2条件そろえば有効", "1", "local", "", true},
		{"true でも有効", "true", "local", "", true},
		{"フラグ未設定なら無効", "", "local", "", false},
		{"APP_ENV 未設定(staging相当)は無効", "1", "", "", false},
		{"APP_ENV=prod は無効", "1", "prod", "", false},
		{"JWKS 設定済みでも有効(ローカルCognito併用)", "1", "local", "https://example/.well-known/jwks.json", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			base(t)
			t.Setenv("LOCAL_PASSWORD_AUTH", c.flag)
			t.Setenv("APP_ENV", c.appEnv)
			t.Setenv("COGNITO_JWK_SET_URI", c.jwks)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, c.want, cfg.LocalPasswordAuth,
				"flag=%q appEnv=%q jwks=%q", c.flag, c.appEnv, c.jwks)
		})
	}
}
