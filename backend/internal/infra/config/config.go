package config

import (
	"fmt"
	"os"
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
	// Region は USER_PASSWORD_AUTH の InitiateAuth を呼ぶ cognitoidp クライアント用。
	Region string
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
			// Cognito が別リージョンの構成も表現できるよう COGNITO_REGION を優先し、未設定時のみ AWS_REGION。
			Region: getEnvOrDefault("COGNITO_REGION", getEnvOrDefault("AWS_REGION", "ap-northeast-1")),
		},
		S3: S3Config{
			Region:           getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			NoteImagesBucket: os.Getenv("NOTE_IMAGES_BUCKET"),
			// ecs.yml が渡す環境変数名は NOTE_IMAGES_CDN_URL。
			NoteImagesCDNBase: getEnvOrDefault("NOTE_IMAGES_CDN_URL", ""),
		},
		Bedrock: BedrockConfig{
			Region: getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			// Claude 4 系は on-demand では呼べず Inference Profile 経由必須
			// （foundation-model ARN 直指定だと ConverseStream が ValidationException を返す）。
			// IAM 側でも inference-profile ARN を許可しておくこと。
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
	// DATABASE_URL か DB_HOST の少なくとも一方は必須。
	if cfg.DatabaseURL == "" && cfg.DBHost == "" {
		return nil, fmt.Errorf("DATABASE_URL or DB_HOST is required")
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
