package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 認証の設定が欠けているときは起動を止める。
//
// 以前は「JWKS が無く、かつ APP_ENV が local なら署名検証をしない」という逃げ道があった。
// APP_ENV は未設定でも既定値 local に解決されるため、環境変数を注入し忘れた環境は
// そのまま「署名を検証しないまま動く」状態になり得た。落ちる側に倒せば必ず気づくが、
// 通す側に倒すと誰も気づかない。
func Test_設定_OIDCが欠けていたら起動を止める(t *testing.T) {
	required := []string{
		"OIDC_ISSUER",
		"OIDC_JWKS_URI",
		"OIDC_TOKEN_URI",
		"OIDC_CLIENT_ID",
		"OIDC_REDIRECT_URI",
	}
	for _, missing := range required {
		t.Run(missing+" が無い", func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv(missing, "")

			_, err := Load()
			require.Error(t, err, "%s が無いのに起動できてしまった", missing)
			assert.Contains(t, err.Error(), "OIDC")
		})
	}
}

// APP_ENV を明示しなくても、認証の設定が欠けていれば止まる。
// 「local だから許す」という抜け道を残さないことの確認。
func Test_設定_APP_ENV未設定でも認証設定は必須(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x/y")
	// APP_ENV は設定しない（未設定なら既定の local に解決される）
	t.Setenv("APP_ENV", "")
	// 手元のシェルに OIDC_* が残っていると、このテストが「別の理由で」通ってしまう。
	// 見たいのは「認証設定が無いこと」なので、明示的に空にする。
	for _, k := range []string{
		"OIDC_ISSUER", "OIDC_JWKS_URI", "OIDC_TOKEN_URI",
		"OIDC_CLIENT_ID", "OIDC_REDIRECT_URI",
	} {
		t.Setenv(k, "")
	}

	cfg, err := Load()
	require.Error(t, err, "APP_ENV 未設定でも認証設定は必須のはず")
	assert.Nil(t, cfg)
}

func Test_設定_揃っていれば読み込める(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OIDC_CLIENT_SECRET", "")
	t.Setenv("OIDC_AUDIENCES", " project-1 , client-id ")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "https://issuer.test", cfg.OIDC.Issuer)
	assert.Equal(t, "client-id", cfg.OIDC.ClientID)
	// 秘密が空 = 公開クライアント（PKCE）として扱う。
	assert.Empty(t, cfg.OIDC.ClientSecret)
	// カンマ区切りは前後の空白を落として読む（打ち間違いで一致しなくなるのを避ける）。
	assert.Equal(t, []string{"project-1", "client-id"}, cfg.OIDC.Audiences)
	// 役割の在り処は既定を持つ（発行者を替えるときは設定で差し替える）。
	assert.Equal(t, "urn:zitadel:iam:org:project:roles", cfg.OIDC.AdminRoleClaim)
	assert.Equal(t, "admin", cfg.OIDC.AdminRole)
}

func Test_設定_役割の在り処は設定で差し替えられる(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OIDC_ROLES_CLAIM", "groups")
	t.Setenv("OIDC_ADMIN_ROLE", "operators")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "groups", cfg.OIDC.AdminRoleClaim)
	assert.Equal(t, "operators", cfg.OIDC.AdminRole)
}
