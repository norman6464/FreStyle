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

	Cognito  CognitoConfig
	S3       S3Config
	Bedrock  BedrockConfig
	DynamoDB DynamoDBConfig
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

// CognitoConfig は Cognito Hosted UI / OAuth2 token endpoint との通信に必要な設定。
type CognitoConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	TokenURI     string
	JwkSetURI    string
	UserPoolID   string // AdminCreateUser API 用
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
		Cognito: CognitoConfig{
			ClientID:     os.Getenv("COGNITO_CLIENT_ID"),
			ClientSecret: os.Getenv("COGNITO_CLIENT_SECRET"),
			RedirectURI:  os.Getenv("COGNITO_REDIRECT_URI"),
			TokenURI:     os.Getenv("COGNITO_TOKEN_URI"),
			JwkSetURI:    os.Getenv("COGNITO_JWK_SET_URI"),
			UserPoolID:   os.Getenv("COGNITO_USER_POOL_ID"),
		},
		S3: S3Config{
			Region:           getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			NoteImagesBucket: os.Getenv("NOTE_IMAGES_BUCKET"),
			// ecs.yml が渡す環境変数名は NOTE_IMAGES_CDN_URL。
			NoteImagesCDNBase: getEnvOrDefault("NOTE_IMAGES_CDN_URL", ""),
		},
		Bedrock: BedrockConfig{
			Region:  getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			ModelID: getEnvOrDefault("BEDROCK_MODEL_ID", "anthropic.claude-3-5-haiku-20241022-v1:0"),
		},
		DynamoDB: DynamoDBConfig{
			Region:      getEnvOrDefault("AWS_REGION", "ap-northeast-1"),
			AiChatTable: getEnvOrDefault("DYNAMODB_AI_CHAT_TABLE", "fre_style_ai_chat_dev"),
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
