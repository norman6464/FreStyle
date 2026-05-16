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
//
// 判定 基準 (順 番 に):
//  1. URL 形式 (postgres:// / postgresql://) なら net/url で parse し、 query string の
//     pgbouncer=true / ホスト 名 が pooler.supabase.com を 含む かを 厳密 に 見る。
//     これ により パスワード / path に 偶然 "pgbouncer=true" の 文字 列 が 含まれて も
//     false positive に なら ない。
//  2. key=value 形式 (host=... user=...) なら host= の 値 だけ を 切り 出して 判定。
func isPgBouncerDSN(dsn string) bool {
	trimmed := strings.TrimSpace(dsn)
	lower := strings.ToLower(trimmed)

	// 1. URL 形式
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

	// 2. key=value 形式 (= 旧 RDS DSN フォールバック)
	// "host=foo.pooler.supabase.com" を 検知 する ため、 host= キー だけ を 厳密 に 取る。
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
