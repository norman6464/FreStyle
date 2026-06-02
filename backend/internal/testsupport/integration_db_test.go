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
