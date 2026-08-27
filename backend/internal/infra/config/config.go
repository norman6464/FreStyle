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

	// LocalPasswordAuth はローカル開発専用のパスワードログイン（infra/localauth）が有効かの
	// 解決済み判定。有効になるのは次の 2 条件を両方満たすときだけ（FRESTYLE-311）:
	//   1. LOCAL_PASSWORD_AUTH が truthy（利用者の明示 opt-in）
	//   2. APP_ENV が「明示的に」local（既定値 "local" ではなく生の env を見る。APP_ENV を
	//      注入しない環境＝staging 等を local 扱いしない）
	// 以前は「三重ガード」と説明していたが、3 か所とも同じ cfg.AppEnv（既定 local）を見るため
	// 実質 1 条件に縮退していた（多角レビューで指摘）。生の APP_ENV を明示要求することで
	// staging（APP_ENV 未設定）では確実に無効化する。加えて localauth の署名鍵はプロセス毎の
	// ランダム値（token.go）なので、固定鍵による偽造も成立しない。
	// JWKS 設定の有無は条件に含めない（ローカルで実 Cognito プールと併用したい構成があるため）。
	LocalPasswordAuth bool

	// BootstrapSuperAdminEmail は「最初の運営管理者」を招待なしで受け入れるための
	// ブートストラップ用アドレス（BOOTSTRAP_SUPER_ADMIN_EMAIL）。
	// 招待は FreStyle 唯一のアカウント発行統制なので、Cognito の admin グループに属している
	// だけで招待を迂回できてはいけない。一方でその免除は「まだ super_admin が 1 人も居ない
	// 環境で最初の 1 人を作る」唯一の経路でもある。そこで免除の条件を
	// 「ここで明示した 1 アドレス」「Cognito admin グループ所属」「super_admin が 0 人」の
	// 3 つ揃いに絞る（判定は usecase 側）。未設定（既定）なら免除は一切効かない。
	BootstrapSuperAdminEmail string

	// CodeRunnerURL はコード実行サイドカー（cmd/coderunner）の baseURL。
	// セットされていると backend 本体は os/exec せず HTTP 越しに runner へ委譲する
	// （例: http://127.0.0.1:9000）。未設定なら in-process サンドボックスで実行する。
	CodeRunnerURL string

	Cognito CognitoConfig
	S3      S3Config
	SES     SESConfig
	SMTP    SMTPConfig
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

// CognitoConfig は Cognito Hosted UI / OAuth2 token endpoint との通信に必要な設定。
// SES マジックリンク方式に切り替えたため AdminCreateUser API は使わない。
type CognitoConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	TokenURI     string
	JwkSetURI    string
	// Region は USER_PASSWORD_AUTH の InitiateAuth を呼ぶ cognitoidp クライアント用。
	Region string
	// UserPoolID は AdminCreateUser（招待の初期パスワード方式・FRESTYLE-313）に使う。
	// 未設定なら初期パスワード方式は無効（マジックリンクは影響なし）。
	UserPoolID string
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
		Cognito: CognitoConfig{
			ClientID:     os.Getenv("COGNITO_CLIENT_ID"),
			ClientSecret: os.Getenv("COGNITO_CLIENT_SECRET"),
			RedirectURI:  os.Getenv("COGNITO_REDIRECT_URI"),
			TokenURI:     os.Getenv("COGNITO_TOKEN_URI"),
			JwkSetURI:    os.Getenv("COGNITO_JWK_SET_URI"),
			// Cognito が別リージョンの構成も表現できるよう COGNITO_REGION を優先し、未設定時のみ AWS_REGION。
			Region:     getEnvOrDefault("COGNITO_REGION", getEnvOrDefault("AWS_REGION", "ap-northeast-1")),
			UserPoolID: os.Getenv("COGNITO_USER_POOL_ID"),
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
	// DATABASE_URL か DB_HOST の少なくとも一方は必須。
	if cfg.DatabaseURL == "" && cfg.DBHost == "" {
		return nil, fmt.Errorf("DATABASE_URL or DB_HOST is required")
	}

	// ローカル専用パスワードログインの有効化判定（上の LocalPasswordAuth の 2 条件）。
	// 生の APP_ENV を見る（既定値 "local" に依存しない）ことが staging での誤有効化を防ぐ肝。
	localPwRequested := os.Getenv("LOCAL_PASSWORD_AUTH") == "1" || os.Getenv("LOCAL_PASSWORD_AUTH") == "true"
	appEnvExplicitLocal := os.Getenv("APP_ENV") == "local"
	cfg.LocalPasswordAuth = localPwRequested && appEnvExplicitLocal
	if localPwRequested && !cfg.LocalPasswordAuth {
		fmt.Fprintf(os.Stderr,
			"WARN: LOCAL_PASSWORD_AUTH は無効化されました（APP_ENV を明示的に local にしてください。現在 APP_ENV=%q）\n",
			os.Getenv("APP_ENV"))
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
