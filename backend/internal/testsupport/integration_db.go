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
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// defaultTestDSN は TEST_DATABASE_URL 未設定時の既定接続先（docker-compose.integration.yml と一致）。
const defaultTestDSN = "postgres://frestyle:frestyle@localhost:5433/frestyle_integration?sslmode=disable"

// integrationLockKey は結合テストを直列化する advisory lock のキー。
// マイグレーション用（database.migrateAdvisoryLockKey）とは別の値にする。
const integrationLockKey int64 = 907_353_401

// OpenTestDB は結合テスト用 DB に接続し、全 domain モデルを AutoMigrate して返す。
// TEST_DATABASE_URL が空 かつ 既定 DSN にも繋がらない場合は t.Skip する
// （ローカルで docker を上げずに `-tags=integration` を流しても落ちないように）。
func OpenTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return openTestDB(t, false)
}

// OpenTestDBSimpleProtocol は simple query protocol を強制した接続で OpenTestDB と同じ初期化を行う。
//
// 本番は Supabase transaction pooler 経由で simple protocol になり、pgx がパラメータを
// クライアント側で SQL リテラルへ埋め込む（Go の型がそのまま SQL の構文を決める）。
// extended protocol（ローカル / CI の既定）はパラメータの OID で型が伝わるため、
// []byte と json.RawMessage の取り違えのような欠陥がローカルでは緑のまま本番でだけ落ちる。
// その系統の回帰テストはこちらの接続で書く。
func OpenTestDBSimpleProtocol(t *testing.T) *gorm.DB {
	t.Helper()
	return openTestDB(t, true)
}

func openTestDB(t *testing.T, preferSimpleProtocol bool) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	// 安全弁: 接続先が Supabase / 本番 pooler の場合は接続前に必ず落とす。
	// 結合テストは TruncateAll（TRUNCATE ... CASCADE）でテーブルを破壊するため、
	// 誤って TEST_DATABASE_URL に本番 DATABASE_URL を入れた事故で本番データを消さないようにする。
	if looksLikeSupabase(dsn) {
		t.Fatal("結合テストの接続先が Supabase / 本番 pooler を指しています。" +
			"TEST_DATABASE_URL を解除し、ローカルの postgres-integration-test（make test-integration）を使ってください。")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: preferSimpleProtocol,
	}), &gorm.Config{
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

	serializeIntegration(t, sqlDB)

	if err := database.AutoMigrateAll(db); err != nil {
		t.Fatalf("AutoMigrate 失敗: %v", err)
	}
	// users.role_id の解決（Create の resolveRoleID）が roles マスタを前提にするため、
	// 起動時（database.Migrate）と同じくロールを投入しておく。
	if err := database.SeedRoles(db); err != nil {
		t.Fatalf("SeedRoles 失敗: %v", err)
	}
	// FK / CHECK / 部分 UNIQUE も本番（database.Migrate）と同じに揃える。
	if err := database.ApplyUserNormalizationConstraints(db); err != nil {
		t.Fatalf("ApplyUserNormalizationConstraints 失敗: %v", err)
	}
	// rich_documents の FK / CHECK も本番（database.Migrate）と同じに揃える。
	if err := database.ApplyRichDocumentConstraints(db); err != nil {
		t.Fatalf("ApplyRichDocumentConstraints 失敗: %v", err)
	}
	// session_notes の 1 セッション 1 ノート一意制約も本番（database.Migrate）と同じに揃える。
	if err := database.ApplySessionNoteConstraints(db); err != nil {
		t.Fatalf("ApplySessionNoteConstraints 失敗: %v", err)
	}
	// ナレッジ基盤（workspaces / spaces / pages / blocks / …）は GORM を通さない。
	// 起動時（database.Migrate）と同じ明示 DDL を、同じ接続プールへ流す。
	if err := database.ApplyKnowledgeBaseSchema(t.Context(), sqlDB); err != nil {
		t.Fatalf("ApplyKnowledgeBaseSchema 失敗: %v", err)
	}
	// companies / users の workspace_id 列と FK も本番（database.Migrate）と同じに揃える。
	// バックフィルは呼ばない（テストが自分でデータを用意する。起動相当の再実行は
	// バックフィル自身の結合テストが直接呼んで確かめる）。
	if err := database.ApplyTenantBridgeSchema(t.Context(), sqlDB); err != nil {
		t.Fatalf("ApplyTenantBridgeSchema 失敗: %v", err)
	}
	return db
}

// serializeIntegration は結合テストをテスト関数の単位で直列化する。
//
// 結合テストは 1 台の PostgreSQL を共有し、TruncateAll でテーブルを TRUNCATE CASCADE しながら
// 使う。go test はパッケージを並列に走らせるので、結合テストを持つパッケージが 2 つ以上になると
// 互いの行を消し合い、テストの成否が実行順に左右される（デッドロックにもなる）。
// 接続時に session 単位の advisory lock を取り、テスト終了時に解放することで、
// パッケージをまたいでも同時に走るのは 1 テスト関数だけになる。
//
// ロックは pool 内の 1 本の接続に固定して取る（pool 任せだと解放が別の接続で走り、
// ロックが残ったままになる）。テストが途中で落ちても、接続が閉じれば PostgreSQL 側で解放される。
func serializeIntegration(t *testing.T, sqlDB *sql.DB) {
	t.Helper()
	ctx := t.Context()
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		t.Fatalf("結合テスト直列化用の接続を取得できません: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, integrationLockKey); err != nil {
		_ = conn.Close()
		t.Fatalf("結合テスト直列化用の advisory lock を取得できません: %v", err)
	}
	t.Cleanup(func() {
		//nolint:errcheck // 解放に失敗しても接続を閉じれば PostgreSQL 側で解放される
		_, _ = conn.ExecContext(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, integrationLockKey)
		_ = conn.Close()
	})
}

// looksLikeSupabase は DSN が Supabase / 本番 pooler を指しているかを返す（安全弁用）。
func looksLikeSupabase(dsn string) bool {
	l := strings.ToLower(dsn)
	return strings.Contains(l, "supabase.com") || strings.Contains(l, "pooler.supabase")
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
