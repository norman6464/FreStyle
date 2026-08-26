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
	const custom = "postgres://frestyle:frestyle@localhost:15432/frestyle_integration?sslmode=disable"

	cases := []struct {
		name     string
		env      string
		wantDSN  string
		explicit bool
	}{
		{"未設定は既定 DSN を使い明示指定ではない", "", defaultTestDSN, false},
		{"設定値はそのまま使い明示指定として扱う", custom, custom, true},
		// 空白だけの値を「未設定」に丸めてはいけない。丸めると到達不能が skip に落ち、
		// 接続枯渇や設定ミスがパッケージの exit 0 に紛れて誰も気付けなくなる。
		// 明示指定として扱い、接続に失敗させて落とすのが正しい。
		{"空白だけでも明示指定として扱う", "   ", "   ", true},
		{"タブや改行だけでも明示指定として扱う", "\t\n", "\t\n", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dsn, explicit := resolveTestDSN(c.env)
			if dsn != c.wantDSN {
				t.Errorf("dsn: got %q, want %q", dsn, c.wantDSN)
			}
			if explicit != c.explicit {
				t.Errorf("explicit: got %v, want %v", explicit, c.explicit)
			}
		})
	}
}

// TestCloseGormPool_Integration は接続プールの解放が nil でも落ちないことを検証する（DB 不要）。
// 実際に開いたプールを閉じる経路は、結合テストごとの t.Cleanup が毎回通る。
func TestCloseGormPool_Integration(t *testing.T) {
	closeGormPool(nil) // panic しなければよい
}
