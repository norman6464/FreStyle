//go:build integration

// Package testsupport は結合テスト（-tags=integration）用のヘルパを提供する。
// 本物の PostgreSQL（docker-compose.integration.yml）に接続し、スキーマ初期化と
// テスト間のクリーンアップを行う。単体テストのビルドには含まれない（build tag で隔離）。
//
// 命名規約: 結合テストの関数名には "Integration" を含めること。
// CI / make test-integration は `go test -tags=integration -run Integration ./...` で
// 結合テストだけを選別して回す（env 依存の単体テストを巻き込まないため）。
package testsupport

import (
	"os"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// defaultTestDSN は TEST_DATABASE_URL 未設定時の既定接続先（docker-compose.integration.yml と一致）。
const defaultTestDSN = "postgres://frestyle:frestyle@localhost:5433/frestyle_integration?sslmode=disable"

// OpenTestDB は結合テスト用 DB に接続し、全 domain モデルを AutoMigrate して返す。
// TEST_DATABASE_URL が空 かつ 既定 DSN にも繋がらない場合は t.Skip する
// （ローカルで docker を上げずに `-tags=integration` を流しても落ちないように）。
func OpenTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("結合テスト用 PostgreSQL に接続できません（docker compose -f docker-compose.integration.yml up -d 済か確認）: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql.DB 取得失敗: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("結合テスト用 PostgreSQL に ping 失敗: %v", err)
	}

	if err := database.AutoMigrateAll(db); err != nil {
		t.Fatalf("AutoMigrate 失敗: %v", err)
	}
	return db
}

// TruncateAll はテーブルを TRUNCATE して連番をリセットする。テスト間の独立性確保用。
// 列挙したテーブルは結合テストが触る範囲に限定する（必要に応じて足す）。
func TruncateAll(t *testing.T, db *gorm.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error; err != nil {
			t.Fatalf("TRUNCATE %s 失敗: %v", table, err)
		}
	}
}
