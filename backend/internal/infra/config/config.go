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
