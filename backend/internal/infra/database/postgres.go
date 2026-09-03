package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
)

// NewPostgres は PostgreSQL に接続し、アプリ全体で共有する接続プール（*sql.DB）を返す。
// pgbouncer 経由（Supabase transaction pooler）の場合は simple query protocol を強制する
// （pgbouncer はセッションをまたいで prepared statement を保てず
// "prepared statement does not exist" になるため）。
func NewPostgres(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.PostgresDSN()
	db, err := OpenSQLDB(dsn, isPgBouncerDSN(dsn))
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	// 起動時に一度だけ疎通を確かめる（接続不能なら listen を始めずに落とす）。
	if err := db.Ping(); err != nil {
		//nolint:errcheck // 打ち切り経路の後始末。閉じられなくても報告する先がない
		_ = db.Close()
		return nil, fmt.Errorf("failed to connect postgres: %w", err)
	}
	return db, nil
}

// OpenSQLDB は DSN から接続プールを開く。preferSimpleProtocol が真なら simple query protocol を
// 強制する（本番の transaction pooler と同じ経路を再現したい結合テストからも使う）。
//
// simple protocol では pgx がパラメータをクライアント側で SQL リテラルへ埋め込むため、
// Go の型がそのまま SQL の構文を決める（extended protocol は型を OID で伝える）。
// この違いで挙動が変わる欠陥があるので、本番と同じモードを選べるようにしてある。
func OpenSQLDB(dsn string, preferSimpleProtocol bool) (*sql.DB, error) {
	connCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres dsn: %w", err)
	}
	if preferSimpleProtocol {
		connCfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}
	return stdlib.OpenDB(*connCfg), nil
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
