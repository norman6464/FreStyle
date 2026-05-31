package database

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgres は GORM で PostgreSQL に接続する。
// pgbouncer 経由（Supabase transaction pooler）の場合は prepared statement を無効化し
// simple query protocol を強制する（"prepared statement does not exist" 対策）。
func NewPostgres(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	dsn := cfg.PostgresDSN()
	pgCfg := postgres.Config{
		DSN: dsn,
	}
	if isPgBouncerDSN(dsn) {
		gormCfg.PrepareStmt = false
		pgCfg.PreferSimpleProtocol = true
	}

	db, err := gorm.Open(postgres.New(pgCfg), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetMaxOpenConns(10)

	return db, nil
}

// isPgBouncerDSN は DSN が pgbouncer 経由かを判定する。
// URL 形式は query の pgbouncer=true / host に pooler.supabase.com を厳密に見て
// パスワードや path に紛れた文字列で false positive にならないようにする。
// key=value 形式は host= の値だけを切り出して判定する。
func isPgBouncerDSN(dsn string) bool {
	trimmed := strings.TrimSpace(dsn)
	lower := strings.ToLower(trimmed)

	if strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return false
		}
		if u.Query().Get("pgbouncer") == "true" {
			return true
		}
		if strings.Contains(strings.ToLower(u.Host), "pooler.supabase.com") {
			return true
		}
		return false
	}

	// key=value 形式は host= キーだけを厳密に取る。
	for _, kv := range strings.Fields(trimmed) {
		eq := strings.IndexByte(kv, '=')
		if eq <= 0 {
			continue
		}
		k := strings.ToLower(kv[:eq])
		v := strings.ToLower(kv[eq+1:])
		if k == "host" && strings.Contains(v, "pooler.supabase.com") {
			return true
		}
	}
	return false
}
