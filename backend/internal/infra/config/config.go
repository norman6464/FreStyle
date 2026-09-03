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

	// AppBaseURL は招待マジックリンクの組み立てに使う、フロントエンドの絶対 URL。
	// 例: https://frestyle.jp (末尾スラッシュ無し / 有り どちらも可)
	AppBaseURL string

	// BootstrapSuperAdminEmail は「最初の運営管理者」を招待なしで受け入れるための
	// ブートストラップ用アドレス（BOOTSTRAP_SUPER_ADMIN_EMAIL）。
	// 招待は FreStyle 唯一のアカウント発行統制なので、発行者側で管理者の役割を持っている
	// だけで招待を迂回できてはいけない。一方でその免除は「まだ super_admin が 1 人も居ない
	// 環境で最初の 1 人を作る」唯一の経路でもある。そこで免除の条件を
	// 「ここで明示した 1 アドレス」「発行者側の管理者ロール所持」「super_admin が 0 人」の
	// 3 つ揃いに絞る（判定は usecase 側）。未設定（既定）なら免除は一切効かない。
	BootstrapSuperAdminEmail string

	// CodeRunnerURL はコード実行サイドカー（cmd/coderunner）の baseURL。
	// セットされていると backend 本体は os/exec せず HTTP 越しに runner へ委譲する
	// （例: http://127.0.0.1:9000）。未設定なら in-process サンドボックスで実行する。
	CodeRunnerURL string

	OIDC OIDCConfig
	S3   S3Config
	SES  SESConfig
	SMTP SMTPConfig
}

// S3Config は profile / note 画像 upload の presign 発行に必要な設定。
type S3Config struct {
	Region           string
	NoteImagesBucket string
}

// SESConfig は招待マジックリンクメール送信用の SES v2 設定。
// FromAddress は SES で検証済の送信元（例: "FreStyle <noreply@frestyle.jp>"）。
// 未設定（空文字）のときは送信スキップ → token をログに残してフォールバック。
type SESConfig struct {
	Region      string
	FromAddress string
}

// SMTPConfig は SES を使わない環境（staging）向けのメール送信設定。
// Host が設定されているときは SES より優先して SMTP で送信する
// （staging の box 上メールキャッチャーが宛先。認証・TLS なしの内部ネットワーク前提）。
type SMTPConfig struct {
	Host        string
	Port        string
	FromAddress string
}

// OIDCConfig は OpenID Connect の発行者と話すために要る設定。
//
// 以前は COGNITO_* という名前で、issuer を持っていなかった（JWKS の URL から
// 文字列を削って推測していた）。発行者の URL の形に依存する推測なので、
// 発行者を替えると黙って壊れる。issuer は必ず明示する。
type OIDCConfig struct {
	// Issuer は発行者の識別子。トークンの iss と完全一致する値。
	Issuer string
	// AuthorizeURI はログイン画面へ送る認可要求の宛先。フロントエンドが使う。
	AuthorizeURI string
	// TokenURI は認可コードとリフレッシュトークンの交換先。
	TokenURI string
	// JWKSURI は署名鍵の取得先。
	JWKSURI string
	// EndSessionURI はログアウトのとき発行者側のセッションも終わらせる宛先。
	// 空なら発行者側のセッションは残る（同じ端末で再ログインが素通りになる）。
	EndSessionURI string
	ClientID      string
	// ClientSecret は機密クライアントのときだけ設定する。
	// 空なら公開クライアント（PKCE）として扱う。ブラウザで動くアプリは秘密を
	// 持てないので、こちらが既定の形。
	ClientSecret string
	RedirectURI  string
	// Audiences は access_token の aud に含まれていることを要求する値（カンマ区切り）。
	// 空なら ClientID を要求する。発行者によっては aud にプロジェクトの識別子を
	// 入れるので、その差を推測ではなく設定で吸収する。
	Audiences []string
	// AdminRoleClaim は役割の一覧が入っているクレーム名。
	// AdminRole はそのうち運営管理者を表す役割名。
	AdminRoleClaim string
	AdminRole      string
}

// Configured は認証に必要な設定が揃っているかを返す。
func (c OIDCConfig) Configured() bool {
	return c.Issuer != "" && c.JWKSURI != "" && c.TokenURI != "" &&
		c.ClientID != "" && c.RedirectURI != ""
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:        getEnvOrDefault("APP_ENV", "local"),
		ServerPort:    getEnvOrDefault("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        getEnvOrDefault("DB_PORT", "5432"),
		DBUser:        getEnvOrDefault("DB_USER", "postgres"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        getEnvOrDefault("DB_NAME", "fre_style"),
		DBSSLMode:     getEnvOrDefault("DB_SSLMODE", "require"),
		AppBaseURL:    getEnvOrDefault("APP_BASE_URL", ""),
		CodeRunnerURL: os.Getenv("CODE_RUNNER_URL"),
		// 前後の空白は運用者の打ち間違いで免除が黙って効かなくなるのを避けるため落とす。
		BootstrapSuperAdminEmail: strings.TrimSpace(os.Getenv("BOOTSTRAP_SUPER_ADMIN_EMAIL")),
		OIDC: OIDCConfig{
			Issuer:        os.Getenv("OIDC_ISSUER"),
			AuthorizeURI:  os.Getenv("OIDC_AUTHORIZE_URI"),
			TokenURI:      os.Getenv("OIDC_TOKEN_URI"),
			JWKSURI:       os.Getenv("OIDC_JWKS_URI"),
			EndSessionURI: os.Getenv("OIDC_END_SESSION_URI"),
			ClientID:      os.Getenv("OIDC_CLIENT_ID"),
			ClientSecret:  os.Getenv("OIDC_CLIENT_SECRET"),
			RedirectURI:   os.Getenv("OIDC_REDIRECT_URI"),
			Audiences:     splitAndTrim(os.Getenv("OIDC_AUDIENCES")),
			// 役割の在り処は発行者ごとに違うので設定で指す。既定は Zitadel の
			// プロジェクトロール（役割名を鍵にした表で入る）。
			AdminRoleClaim: getEnvOrDefault("OIDC_ROLES_CLAIM", "urn:zitadel:iam:org:project:roles"),
			AdminRole:      getEnvOrDefault("OIDC_ADMIN_ROLE", "admin"),
		},
		S3: S3Config{
			Region:           getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			NoteImagesBucket: os.Getenv("NOTE_IMAGES_BUCKET"),
		},
		SES: SESConfig{
			Region:      getEnvOrDefault("SES_REGION", getEnvOrDefault("AWS_REGION", "ap-northeast-1")),
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

	// 認証の設定は揃っているか揃っていないかのどちらかにする。**足りないまま起動しない。**
	//
	// 以前は「JWKS が無く、かつ APP_ENV が local なら署名検証をしない」という逃げ道があった。
	// APP_ENV は未設定でも既定値 local に解決されるので、環境変数を注入し忘れた環境は
	// そのまま「署名を検証しない本番」になり得た。落ちる側に倒しても気づけるが、
	// 通す側に倒すと誰も気づかない。だから起動時に止める。
	if !cfg.OIDC.Configured() {
		return nil, fmt.Errorf(
			"OIDC の設定が足りません（OIDC_ISSUER / OIDC_JWKS_URI / OIDC_TOKEN_URI / OIDC_CLIENT_ID / OIDC_REDIRECT_URI は必須）: "+
				"issuer=%t jwks=%t token=%t client_id=%t redirect_uri=%t",
			cfg.OIDC.Issuer != "", cfg.OIDC.JWKSURI != "", cfg.OIDC.TokenURI != "",
			cfg.OIDC.ClientID != "", cfg.OIDC.RedirectURI != "",
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
