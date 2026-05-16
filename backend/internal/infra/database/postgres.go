package database

import (
	"fmt"
	"strings"

	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgres は GORM で PostgreSQL に 接続 する。
//
// DATABASE_URL が "pgbouncer=true" を 含む (= Supabase transaction pooler 接続) 場合 は
// GORM の prepared statement キャッシュ を 無効 化 + driver 側 で simple query protocol を
// 強制 する。 transaction-mode の pgbouncer は 接続 単位 の prepared statement を サポート
// し ない ため、 デフォルト の まま だ と "prepared statement does not exist" エラー が
// ランダム に 発生 する。
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

// isPgBouncerDSN は DSN が Supabase transaction pooler 等 の pgbouncer 経由 か を 判定 する。
// 判定 基準: URL 形式 で "pgbouncer=true" クエリ パラメータ が 付いて いる か、
// または ホスト 名 に "pooler.supabase.com" が 含まれる か。
func isPgBouncerDSN(dsn string) bool {
	d := strings.ToLower(dsn)
	return strings.Contains(d, "pgbouncer=true") ||
		strings.Contains(d, "pooler.supabase.com")
}
