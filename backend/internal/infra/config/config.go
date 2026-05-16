package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv     string
	ServerPort string

	// DatabaseURL は Supabase 等 の マネージド Postgres 用 の 完全 接続 文字 列。
	// 例: "postgresql://postgres.xxxx:pwd@aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres?pgbouncer=true&sslmode=require"
	// 環境 変数 DATABASE_URL がセットされていると こちらを 優先 し、 個別 の DB_HOST 等 は 無視 する。
	// RDS から Supabase へ の 段階 移行 用 (= URL を 切り替える だけ で 接続 先 を 変えられる)。
	DatabaseURL string

	// 個別 接続 設定 (DATABASE_URL 未 セット 時 の フォールバック / ローカル 開発 用)
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// AppBaseURL は招待マジックリンクの組み立てに使う、フロントエンドの絶対 URL。
	// 例: https://normanblog.com (末尾スラッシュ無し / 有り どちらも可)
	AppBaseURL string

	Cognito  CognitoConfig
	S3       S3Config
	Bedrock  BedrockConfig
	DynamoDB DynamoDBConfig
	SES      SESConfig
}

// S3Config は profile / note 画像 upload の presign 発行に必要な設定。
type S3Config struct {
	Region            string
	NoteImagesBucket  string
	NoteImagesCDNBase string // 配信用 CDN URL (CloudFront / virtual-hosted)
}

// BedrockConfig は AWS Bedrock Converse API 呼び出しに必要な設定。
type BedrockConfig struct {
	Region  string
	ModelID string
}

// DynamoDBConfig は AI チャットメッセージを保存する DynamoDB の設定。
type DynamoDBConfig struct {
	Region      string
	AiChatTable string
}

// SESConfig は招待マジックリンクメール送信用の SES v2 設定。
// FromAddress は SES で検証済の送信元（例: "FreStyle <noreply@normanblog.com>"）。
// 未設定（空文字）のときは送信スキップ → token をログに残してフォールバック。
type SESConfig struct {
	Region      string
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
		Cognito: CognitoConfig{
			ClientID:     os.Getenv("COGNITO_CLIENT_ID"),
			ClientSecret: os.Getenv("COGNITO_CLIENT_SECRET"),
			RedirectURI:  os.Getenv("COGNITO_REDIRECT_URI"),
			TokenURI:     os.Getenv("COGNITO_TOKEN_URI"),
			JwkSetURI:    os.Getenv("COGNITO_JWK_SET_URI"),
		},
		S3: S3Config{
			Region:           getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			NoteImagesBucket: os.Getenv("NOTE_IMAGES_BUCKET"),
			// ecs.yml が渡す環境変数名は NOTE_IMAGES_CDN_URL。
			NoteImagesCDNBase: getEnvOrDefault("NOTE_IMAGES_CDN_URL", ""),
		},
		Bedrock: BedrockConfig{
			Region: getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			// Claude 4 系（haiku-4-5 / sonnet-4-x / opus-4-x）は on-demand throughput では呼べず、
			// Inference Profile 経由必須。foundation-model ARN を直接指定すると ConverseStream が
			// ValidationException で 400 を返す（2026-05-08 本番でユーザーから AI チャット返信なし
			// の報告で発覚）。
			//
			// `jp.anthropic.claude-sonnet-4-5-20250929-v1:0` が ap-northeast-1 用の Japan inference
			// profile（sonnet-4.5）。IAM 側 (frestyle-infrastructure: BedrockInvokeAccess) でも
			// `arn:aws:bedrock:${region}:${account}:inference-profile/jp.anthropic.claude-sonnet-4-*`
			// を許可しておく必要がある。
			//
			// モデル選定: 旧 default の haiku-4-5 から sonnet-4-5 にアップグレード。応答品質を優先し、
			// チャット用途なら latency / コストとも実用範囲（ユーザー指示・2026-05-08）。
			//
			// 利用可能 inference profile 一覧:
			//   aws bedrock list-inference-profiles --region ap-northeast-1
			ModelID: getEnvOrDefault("BEDROCK_MODEL_ID", "jp.anthropic.claude-sonnet-4-5-20250929-v1:0"),
		},
		DynamoDB: DynamoDBConfig{
			Region:      getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			AiChatTable: getEnvOrDefault("DYNAMODB_AI_CHAT_TABLE", "fre_style_ai_chat_dev"),
		},
		SES: SESConfig{
			Region:      getEnvOrDefault("SES_REGION", getEnvOrDefault("AWS_REGION", "ap-northeast-1")),
			FromAddress: os.Getenv("SES_FROM_ADDRESS"),
		},
	}
	// DATABASE_URL がセットされていればそちら 優先 / DB_HOST フォールバック で
	// 少なくとも どちら か は 設定 されて いる 必要 が ある。
	if cfg.DatabaseURL == "" && cfg.DBHost == "" {
		return nil, fmt.Errorf("DATABASE_URL or DB_HOST is required")
	}
	return cfg, nil
}

// PostgresDSN は GORM に 渡す DSN を 返す。
// DATABASE_URL (例: Supabase pooler の "postgresql://..." 形式) が あれば そのまま 返す。
// 無ければ 旧 RDS 形式 の 個別 設定 から DSN を 組み立てる (= 後方 互換)。
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
