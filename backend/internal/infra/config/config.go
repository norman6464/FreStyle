package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv     string
	ServerPort string
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
		AppEnv:     getEnvOrDefault("APP_ENV", "local"),
		ServerPort: getEnvOrDefault("PORT", "8080"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getEnvOrDefault("DB_NAME", "fre_style"),
		DBSSLMode:  getEnvOrDefault("DB_SSLMODE", "require"),
		AppBaseURL: getEnvOrDefault("APP_BASE_URL", ""),
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
			// ap-northeast-1 で ACTIVE な Anthropic モデルを default にする。
			// 旧 default の "anthropic.claude-3-5-haiku-20241022-v1:0" は当 region に存在せず
			// Bedrock Converse API が ValidationException: invalid model identifier を返していた
			// （2026-05-06 本番でユーザーから AI チャット返信なしの報告で発覚）。
			// 利用可能 model 確認:
			//   aws bedrock list-foundation-models --region ap-northeast-1 \
			//     --query 'modelSummaries[?providerName==`Anthropic`].modelId'
			ModelID: getEnvOrDefault("BEDROCK_MODEL_ID", "anthropic.claude-haiku-4-5-20251001-v1:0"),
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
	if cfg.DBHost == "" {
		return nil, fmt.Errorf("DB_HOST is required")
	}
	return cfg, nil
}

func (c *Config) PostgresDSN() string {
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
