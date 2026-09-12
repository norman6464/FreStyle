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
	SES    SESConfig
	SMTP   SMTPConfig
}

// ImagesConfig は profile / リッチテキスト画像 / KB ページ画像 upload の presign 発行に必要な
// 設定。バケットは Cloud Storage（internal/infra/gcs.Presigner）。リージョンは持たない
// —— GCS のバケット名はプロジェクト内で一意なグローバル名前空間で、クライアント側の呼び出しに
// リージョン指定は要らない（AWS S3 の Region とは異なる）。
type ImagesConfig struct {
	Bucket string
}

// SESConfig は招待マジックリンクメール送信用の SES v2 設定。FromAddress は SES で検証済の
// 送信元。未設定（空文字）のときは送信スキップ → token をログに残してフォールバック。
type SESConfig struct {
	Region      string
	FromAddress string
}

// SMTPConfig は SES を使わない環境（staging）向けのメール送信設定。Host が設定されていれば
// SES より優先して SMTP で送信する（staging の box 上メールキャッチャー宛。認証・TLS なしの
// 内部ネットワーク前提）。
type SMTPConfig struct {
	Host        string
	Port        string
	FromAddress string
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
		// SESConfig.Region は SES を実際に呼び出すコードが無いため（招待メール送信は
		// toC 化で無くなった）、AWS_REGION には連鎖させず固定の既定値にする。
		SES: SESConfig{
			Region:      getEnvOrDefault("SES_REGION", "ap-northeast-1"),
			FromAddress: os.Getenv("SES_FROM_ADDRESS"),
		},
		SMTP: SMTPConfig{
			Host:        os.Getenv("MAIL_SMTP_HOST"),
			Port:        getEnvOrDefault("MAIL_SMTP_PORT", "1025"),
			FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
		},
	}
	if cfg.DatabaseURL == "" && cfg.DBHost == "" {
		return nil, fmt.Errorf("DATABASE_URL or DB_HOST is required")
	}

	// 認証設定は揃っているか揃っていないかのどちらかにする。**足りないまま起動しない。**
	// 以前は「JWKS が無く APP_ENV が local なら署名検証をしない」という逃げ道があった。
	// APP_ENV は未設定でも既定値 local に解決されるため、環境変数を注入し忘れた環境が
	// そのまま「署名を検証しない本番」になり得た。通す側に倒すと誰も気づけないので、
	// 起動時に止める側に倒す。
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
