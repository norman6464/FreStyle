package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppEnv     string
	ServerPort string

	// DatabaseURL は Supabase 等のマネージド Postgres の完全接続文字列。
	// セットされていると DB_HOST 等より優先される。
	DatabaseURL string

	// 個別接続設定（DATABASE_URL 未設定時のフォールバック / ローカル開発用）。
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// AppBaseURL はフロントエンドの絶対 URL。
	// 例: https://frestyle.dev (末尾スラッシュ無し / 有り どちらも可)
	AppBaseURL string

	OIDC   OIDCConfig
	Images ImagesConfig
}

// ImagesConfig は profile / リッチテキスト画像 / KB ページ画像 upload の presign 発行に必要な
// 設定。バケットは Cloud Storage（internal/infra/gcs.Presigner）。GCS のバケット名は
// プロジェクト内で一意なグローバル名前空間なので、クライアント側の呼び出しにリージョン
// 指定は要らない。
type ImagesConfig struct {
	Bucket string
}

// OIDCConfig は Bearer の ID トークンを検証するために要る設定。
//
// GCIP（Google Cloud Identity Platform）はクライアント SDK でサインインして ID トークンを
// 直接受け取る設計で、`/authorize` `/token` を提供する認可サーバーとしては振る舞わない。
// 認可コード交換・リフレッシュ・クライアントシークレットが無いため、従来の OpenID Connect
// （認可コード + PKCE）向けの項目（AuthorizeURI / TokenURI / EndSessionURI / ClientID /
// ClientSecret / RedirectURI）は持たない。
type OIDCConfig struct {
	// Issuer は発行者の識別子。トークンの iss と完全一致する値。
	Issuer string
	// JWKSURI は署名鍵の取得先。
	JWKSURI string
	// Audiences は ID トークンの aud に含まれていることを要求する値（カンマ区切り）。
	// GCIP は aud に client_id ではなく GCP のプロジェクト ID を入れる。
	Audiences []string
}

// Configured は認証に必要な設定が揃っているかを返す。
func (c OIDCConfig) Configured() bool {
	return c.Issuer != "" && c.JWKSURI != "" && len(c.Audiences) > 0
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:      getEnvOrDefault("APP_ENV", "local"),
		ServerPort:  getEnvOrDefault("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      getEnvOrDefault("DB_PORT", "5432"),
		DBUser:      getEnvOrDefault("DB_USER", "postgres"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      getEnvOrDefault("DB_NAME", "fre_style"),
		DBSSLMode:   getEnvOrDefault("DB_SSLMODE", "require"),
		AppBaseURL:  getEnvOrDefault("APP_BASE_URL", ""),
		OIDC: OIDCConfig{
			Issuer:    os.Getenv("OIDC_ISSUER"),
			JWKSURI:   os.Getenv("OIDC_JWKS_URI"),
			Audiences: splitAndTrim(os.Getenv("OIDC_AUDIENCES")),
		},
		Images: ImagesConfig{
			Bucket: os.Getenv("IMAGES_BUCKET"),
		},
	}
	if cfg.DatabaseURL == "" && cfg.DBHost == "" {
		return nil, fmt.Errorf("DATABASE_URL or DB_HOST is required")
	}

	// 認証設定は揃っているか揃っていないかのどちらかにする。**足りないまま起動しない。**
	// 発行者を通さない逃げ道は作らない — APP_ENV は未設定でも既定値 local に解決されるため、
	// そのような逃げ道があると環境変数を注入し忘れた環境がそのまま「署名を検証しない本番」に
	// なり得る。通す側に倒すと誰も気づけないので、起動時に止める側に倒す。
	if !cfg.OIDC.Configured() {
		return nil, fmt.Errorf(
			"OIDC の設定が足りません（OIDC_ISSUER / OIDC_JWKS_URI / OIDC_AUDIENCES は必須）: "+
				"issuer=%t jwks=%t audiences=%t",
			cfg.OIDC.Issuer != "", cfg.OIDC.JWKSURI != "", len(cfg.OIDC.Audiences) > 0,
		)
	}

	return cfg, nil
}

// PostgresDSN は GORM に渡す DSN を返す。DATABASE_URL があればそのまま、
// 無ければ個別設定から key=value 形式の DSN を組み立てる。
func (c *Config) PostgresDSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// splitAndTrim はカンマ区切りの設定値を、空要素を落として配列にする。
// 打ち間違いの空白で値が一致しなくなるのを避けるため前後の空白を落とす。
func splitAndTrim(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
