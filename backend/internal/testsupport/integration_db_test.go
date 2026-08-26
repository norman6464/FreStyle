//go:build integration

package testsupport

import "testing"

// TestLooksLikeSupabase_Integration は安全弁 looksLikeSupabase の判定を検証する（DB 不要）。
func TestLooksLikeSupabase_Integration(t *testing.T) {
	blocked := []string{
		"postgres://postgres.ref:pw@aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres",
		"postgresql://x:y@db.abcd.supabase.com:5432/postgres?sslmode=require",
		"host=aws-0-ap-northeast-1.POOLER.SUPABASE.com user=postgres",
	}
	for _, dsn := range blocked {
		if !looksLikeSupabase(dsn) {
			t.Errorf("Supabase と判定されるべき: %q", dsn)
		}
	}

	allowed := []string{
		"postgres://frestyle:frestyle@localhost:5433/frestyle_integration?sslmode=disable",
		"host=127.0.0.1 port=5433 user=frestyle dbname=frestyle_integration",
		"postgres://u:p@postgres-integration-test:5432/frestyle_integration",
	}
	for _, dsn := range allowed {
		if looksLikeSupabase(dsn) {
			t.Errorf("ローカルは許可されるべき: %q", dsn)
		}
	}
}

// TestResolveTestDSN_Integration は接続先の解決と「明示指定されたか」の判定を検証する（DB 不要）。
// explicit が true のとき到達不能は skip ではなく fail になるため、ここが分岐の入り口になる。
func TestResolveTestDSN_Integration(t *testing.T) {
	dsn, explicit := resolveTestDSN("")
	if explicit {
		t.Error("未設定は明示指定ではない")
	}
	if dsn != defaultTestDSN {
		t.Errorf("未設定なら既定 DSN を使う: got %q", dsn)
	}

	const custom = "postgres://frestyle:frestyle@localhost:15432/frestyle_integration?sslmode=disable"
	dsn, explicit = resolveTestDSN(custom)
	if !explicit {
		t.Error("設定済みは明示指定として扱う")
	}
	if dsn != custom {
		t.Errorf("設定値をそのまま使う: got %q", dsn)
	}
}
