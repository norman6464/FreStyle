package sandbox_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startTestPG は使い捨ての PostgreSQL を /tmp に initdb + pg_ctl で起動し、
// executeSQL が参照する CODE_PG_* env をセットする。t.Cleanup で停止 + 後始末する。
// initdb / pg_ctl / psql が無い環境（Go-only CI コンテナ等）では Skip する。
// runner コンテナの「socket 専用・非 superuser student」構成をローカルで再現する。
func startTestPG(t *testing.T) {
	t.Helper()
	for _, bin := range []string{"initdb", "pg_ctl", "psql"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not found in PATH, skipping SQL sandbox test", bin)
		}
	}

	dataDir := t.TempDir()
	// unix socket パスは ~104 文字上限があるため socket dir は短い /tmp 配下に作る。
	sockDir, err := os.MkdirTemp("/tmp", "pgs")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(sockDir) })

	run := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
		}
	}

	// trust auth + locale C で素早く初期化（使い捨てなので耐久性設定は不要）。
	run("initdb", "-D", dataDir, "-U", "postgres", "-A", "trust", "--encoding=UTF8", "--locale=C")

	logFile := filepath.Join(dataDir, "pg.log")
	// -h '' で TCP を開かず socket 専用。-k で socket をテスト用 dir に置く。
	run("pg_ctl", "-D", dataDir, "-l", logFile, "-o", "-k "+sockDir+" -h ''", "-w", "start")
	t.Cleanup(func() {
		_ = exec.Command("pg_ctl", "-D", dataDir, "-m", "immediate", "-w", "stop").Run()
	})

	// admin: DB の作成/破棄を行う CREATEDB ロール（superuser ではない）。
	// student: 学習者 SQL を実行する非 superuser ロール。コンテナの本番構成と同じ。
	run("psql", "-h", sockDir, "-U", "postgres", "-d", "postgres", "-v", "ON_ERROR_STOP=1",
		"-c", "CREATE ROLE dbadmin LOGIN CREATEDB NOSUPERUSER NOCREATEROLE")
	run("psql", "-h", sockDir, "-U", "postgres", "-d", "postgres", "-v", "ON_ERROR_STOP=1",
		"-c", "CREATE ROLE student LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE")

	t.Setenv("CODE_PG_HOST", sockDir)
	t.Setenv("CODE_PG_PORT", "5432")
	t.Setenv("CODE_PG_ADMIN", "dbadmin")
	t.Setenv("CODE_PG_STUDENT", "student")
}

func Test_ランナー_SQL_単純SELECT(t *testing.T) {
	startTestPG(t)
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `SELECT 1 AS n;`,
		Language: "sql",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, out.ExitCode)
	// -A -F'|' footer=off: ヘッダ行 + 値行。
	assert.Equal(t, "n\n1\n", out.Stdout)
}

func Test_ランナー_SQL_複数文と集計(t *testing.T) {
	startTestPG(t)
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code: `CREATE TABLE emp(id int, name text, salary int);
INSERT INTO emp VALUES (1,'sato',400),(2,'suzuki',550),(3,'kato',600);
SELECT name FROM emp WHERE salary >= 500 ORDER BY id;`,
		Language: "sql",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, out.ExitCode)
	assert.Equal(t, "name\nsuzuki\nkato\n", out.Stdout)
}

func Test_ランナー_SQL_文タイムアウトで打ち切る(t *testing.T) {
	startTestPG(t)
	r := sandbox.NewRunner()
	start := time.Now()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `SELECT pg_sleep(30);`,
		Language: "sql",
	})
	elapsed := time.Since(start)
	require.NoError(t, err)
	assert.NotEqual(t, 0, out.ExitCode)
	assert.Contains(t, out.Stderr, "timeout")
	assert.Less(t, elapsed, 9*time.Second, "statement_timeout(5s) で早期に打ち切られる")
}

func Test_ランナー_SQL_提出間はDBが隔離される(t *testing.T) {
	startTestPG(t)
	r := sandbox.NewRunner()

	// 1 回目: テーブルを作る（この使い捨て DB は実行後 DROP される）。
	out1, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `CREATE TABLE leaked(x int); INSERT INTO leaked VALUES (42);`,
		Language: "sql",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, out1.ExitCode)

	// 2 回目: 別 DB なので leaked は見えない。
	out2, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `SELECT * FROM leaked;`,
		Language: "sql",
	})
	require.NoError(t, err)
	assert.NotEqual(t, 0, out2.ExitCode)
	assert.Contains(t, out2.Stderr, "leaked")
}

// 非 superuser の student は COPY ... TO PROGRAM（OS コマンド実行）を使えない。
func Test_ランナー_SQL_COPYPROGRAMは権限拒否(t *testing.T) {
	startTestPG(t)
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `COPY (SELECT 1) TO PROGRAM 'echo pwned';`,
		Language: "sql",
	})
	require.NoError(t, err)
	assert.NotEqual(t, 0, out.ExitCode)
	// superuser 限定操作なので permission/ superuser 系のエラーになる。
	assert.NotEmpty(t, out.Stderr)
}

// denylist は PG 起動前の入力検証なので PG 無しでも動く。
func Test_ランナー_SQL_危険メタコマンドを拒否(t *testing.T) {
	r := sandbox.NewRunner()
	// \! はホスト shell、\c / \connect は別ロールへの再接続（昇格）を狙うため必ず弾く。
	for _, code := range []string{"\\! echo pwned", "\\c - dbadmin", "\\connect postgres"} {
		out, err := r.Run(context.Background(), domain.CodeExecutionInput{Code: code, Language: "sql"})
		require.NoError(t, err)
		assert.NotEqual(t, 0, out.ExitCode, "code=%q は拒否されるべき", code)
		assert.Contains(t, out.Stderr, "メタコマンド", "code=%q", code)
	}
}
