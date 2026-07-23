// Package sandbox は学習者コード（php / go / bash / sql / javascript / typescript / java）を
// os/exec でサンドボックス実行する in-process 実装を提供する。backend 本体に同居させる場合
// （CODE_RUNNER_URL 未設定）と、別バイナリ cmd/coderunner（HTTP サーバ）の中身の双方から共有して使う。
package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

const (
	maxCodeBytes = 64 * 1024 // 64 KB
	// execTimeout は PHP / bash / javascript / typescript の上限。
	execTimeout = 8 * time.Second
	// goExecTimeout は Go 専用。go run はコンパイルも走るため余裕を持たせる（cold compile ~6-8s）。
	goExecTimeout = 15 * time.Second
	// javaExecTimeout は Java 専用。JVM 起動 + source launcher のメモリ内コンパイルが
	// Fargate では数秒かかるため Go と同水準に取る。
	javaExecTimeout = 15 * time.Second
	maxOutputBytes  = 64 * 1024 // 64 KB

	// sqlExecTimeout は SQL 1 提出あたりの上限（CREATE/実行/DROP を含む全体）。
	sqlExecTimeout = 10 * time.Second
	// sqlStatementTimeoutMS は個々の文の上限（暴走クエリ・pg_sleep を打ち切る）。
	sqlStatementTimeoutMS = "5000"
	// sqlLockTimeoutMS / sqlIdleTxTimeoutMS でロック待ち・放置トランザクションも縛る。
	sqlLockTimeoutMS   = "2000"
	sqlIdleTxTimeoutMS = "5000"
)

// sqlDeniedMeta は psql の backslash メタコマンドのうち、psql クライアント側で
// ホスト shell / ファイル / プロセスへ逃げられるものを禁止する denylist。
// 真の権限境界は非 superuser ロール `student` だが、多層防御として入力段でも弾く。
// 部分一致（大文字小文字無視）で過検出気味に倒す（学習用 SQL に実害は無い）。
var sqlDeniedMeta = []string{
	`\!`,       // ホスト shell 実行
	`\copy`,    // クライアント側 COPY（runner のファイル読み書き）
	`\c`,       // \c / \connect で別ロール（CREATEDB の dbadmin 等）へ再接続させない
	`\connect`, // 同上（明示）
	`\g`,       // \g| でパイプ先プログラム実行されうる
	`\o`,       // 出力をファイルへ
	`\w`,       // クエリバッファをファイルへ
	`\i`,       // \i / \ir でファイル include
	`\e`,       // \e / \ef / \ev でエディタプロセス起動
	`\lo_`,     // ラージオブジェクト import/export（ファイル I/O）
	`\setenv`,  // 環境変数操作
}

// 実行を禁止する PHP 関数。ファイル操作・OS 操作・ネットワーク通信などを封じる。
var disableFunctions = strings.Join([]string{
	"exec", "system", "shell_exec", "passthru", "popen", "proc_open",
	"pcntl_exec", "proc_get_status", "proc_terminate", "proc_close",
	"file_get_contents", "file_put_contents", "file", "fopen", "fwrite",
	"fread", "fclose", "fgets", "fputs", "file_exists", "unlink",
	"rename", "copy", "mkdir", "rmdir", "opendir", "readdir", "closedir",
	"glob", "scandir", "tempnam", "tmpfile",
	"socket_create", "fsockopen", "pfsockopen",
	"curl_init", "curl_exec", "curl_multi_exec",
	"dl", "phpinfo", "posix_kill", "posix_setuid",
	"getenv", "putenv", "apache_getenv",
	"syslog", "openlog", "closelog",
}, ",")

// Runner は php / go / bash / sql / javascript / typescript / java を os/exec でサンドボックス実行する
// in-process 実装。共通制約: timeout / 64 KB code-size / 64 KB output-size。
type Runner struct{}

// NewRunner は in-process サンドボックス Runner を返す。
func NewRunner() *Runner {
	return &Runner{}
}

// Run は言語別にサンドボックス実行する。
func (r *Runner) Run(ctx context.Context, input domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	if len(input.Code) > maxCodeBytes {
		return nil, fmt.Errorf("コードが大きすぎます (最大 64 KB)")
	}
	switch input.Language {
	case "php":
		return r.executePHP(ctx, input)
	case "go":
		return r.executeGo(ctx, input)
	case "bash":
		return r.executeBash(ctx, input)
	case "sql":
		return r.executeSQL(ctx, input)
	case "javascript", "typescript":
		return r.executeNode(ctx, input)
	case "java":
		return r.executeJava(ctx, input)
	default:
		return nil, fmt.Errorf("未対応の言語: %s", input.Language)
	}
}

// Warmup は実行環境を事前に温める。Go は go run のコンパイルキャッシュを温めるため
// trivial なプログラムを 1 回コンパイル/実行する。php / bash / javascript / typescript / java は
// 永続キャッシュを持たないため no-op（Java の起動コストは javaExecTimeout 側で吸収する）。
func (r *Runner) Warmup(ctx context.Context, language string) error {
	if language != "go" {
		return nil
	}
	_, err := r.executeGo(ctx, domain.CodeExecutionInput{
		Code:     "package main\n\nfunc main() {}\n",
		Language: "go",
	})
	return err
}

func (r *Runner) executePHP(ctx context.Context, input domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	// 開始タグが無いと PHP はソースをそのまま stdout に流すため、別言語の誤貼付けを明示的に弾く。
	if !strings.Contains(input.Code, "<?php") && !strings.Contains(input.Code, "<?=") {
		return &domain.CodeExecutionResult{
			Stdout:   "",
			Stderr:   "PHP コードには `<?php` 開始タグが必要です。 PHP 以外のコードは現在実行できません。",
			ExitCode: 1,
		}, nil
	}

	tmpDir := os.TempDir()
	filename := filepath.Join(tmpDir, "code_"+uuid.NewString()+".php")
	if err := os.WriteFile(filename, []byte(input.Code), 0o600); err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}
	defer os.Remove(filename)

	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"php",
		"-d", "max_execution_time=5",
		"-d", "memory_limit=32M",
		"-d", "disable_functions="+disableFunctions,
		"-d", "open_basedir="+tmpDir,
		"-d", "display_errors=stderr",
		"-d", "log_errors=0",
		// variables_order から E を外し $_ENV を空にする（disable_functions=getenv の抜け穴を塞ぐ）。
		"-d", "variables_order=GPCS",
		filename,
	)
	// AWS 認証情報・注入シークレットを子プロセスに渡さない。
	cmd.Env = sandboxEnv()

	return runCommand(cmd, input.Stdin)
}

func (r *Runner) executeGo(ctx context.Context, input domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	// 別言語コードが意味不明なコンパイルエラーになるのを避けるため package main を必須にする。
	if !strings.Contains(input.Code, "package main") {
		return &domain.CodeExecutionResult{
			Stdout:   "",
			Stderr:   "Go コードには `package main` と `func main()` が必要です。",
			ExitCode: 1,
		}, nil
	}

	// リクエストごとに独立した HOME を作る。build cache は共有で compile 短縮。
	tmpDir, err := os.MkdirTemp("", "go-exec-")
	if err != nil {
		return nil, fmt.Errorf("一時ディレクトリの作成に失敗: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(filename, []byte(input.Code), 0o600); err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, goExecTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", filename)
	// GO111MODULE=off で単一ファイル実行（go.mod 不要）。AWS 認証情報・注入シークレットは
	// sandboxEnv が除外する。
	cmd.Env = sandboxEnv(
		"GO111MODULE=off",
		"HOME="+tmpDir,
	)

	out, err := runCommand(cmd, input.Stdin)
	if out != nil {
		// go のコンパイルエラーは一時ディレクトリの絶対パス（/tmp/go-exec-XXX/main.go:7:9:）を
		// 出力する。内部パスを学習者に見せず、`./main.go:7:9:` の形に整える。
		out.Stderr = strings.ReplaceAll(out.Stderr, tmpDir+string(os.PathSeparator), "./")
	}
	return out, err
}

func (r *Runner) executeBash(ctx context.Context, input domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	// リクエストごとに独立した temp dir を HOME / PWD に固定し、副作用を temp に閉じる。
	tmpDir, err := os.MkdirTemp("", "bash-exec-")
	if err != nil {
		return nil, fmt.Errorf("一時ディレクトリの作成に失敗: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "script.sh")
	if err := os.WriteFile(filename, []byte(input.Code), 0o600); err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", filename)
	cmd.Dir = tmpDir
	// AWS credential や DB password 等を子プロセスに継承しないため、 環境変数は最小限に絞る。
	cmd.Env = []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=" + tmpDir,
		"PWD=" + tmpDir,
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
	}

	// bash はバックグラウンド子プロセスを起動しうるので、独立プロセスグループに置いて一括 kill する。
	configureProcessGroup(cmd)

	return runCommand(cmd, input.Stdin)
}

// executeNode は javascript / typescript を Node.js で実行する。TypeScript は Node 組み込みの
// 型除去（type stripping）でそのまま実行し、別途のコンパイル工程を持たない（起動が速い）。
// enum 等の型除去だけでは動かない構文も --experimental-transform-types で変換して通す。
func (r *Runner) executeNode(ctx context.Context, input domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	// リクエストごとに独立した temp dir を HOME / PWD に固定し、副作用を temp に閉じる。
	tmpDir, err := os.MkdirTemp("", "node-exec-")
	if err != nil {
		return nil, fmt.Errorf("一時ディレクトリの作成に失敗: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 拡張子で Node の TypeScript 処理（型除去）が切り替わるため、言語に合わせる。
	name := "main.js"
	args := []string{}
	if input.Language == "typescript" {
		name = "main.ts"
		args = append(
			args,
			"--experimental-transform-types",
			// 実験機能の警告が stderr に混ざると学習者の実行結果を汚すため抑止する。
			"--disable-warning=ExperimentalWarning",
		)
	}
	// 512MB タスクに同居する他プロセスを圧迫しないようヒープ上限を絞る（学習用途には十分）。
	args = append(args, "--max-old-space-size=128", name)

	filename := filepath.Join(tmpDir, name)
	if err := os.WriteFile(filename, []byte(input.Code), 0o600); err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "node", args...)
	cmd.Dir = tmpDir
	// AWS credential や DB password 等を子プロセスに継承しないため、環境変数は最小限に絞る
	// （NODE_OPTIONS 等の注入も防ぐ）。
	cmd.Env = []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=" + tmpDir,
		"PWD=" + tmpDir,
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
	}

	// child_process 等で子孫プロセスを起動しうるので、独立プロセスグループに置いて一括 kill する。
	configureProcessGroup(cmd)

	out, runErr := runCommand(cmd, input.Stdin)
	if out != nil {
		// スタックトレース等に出る一時ディレクトリの絶対パスを学習者に見せない。
		// Node は main モジュールを realpath 解決してから出力する（macOS は /var → /private/var の
		// symlink）ため、解決後のパスを先に置換しないと部分一致で `/private./main.js` に崩れる。
		paths := []string{tmpDir}
		if real, err := filepath.EvalSymlinks(tmpDir); err == nil && real != tmpDir {
			paths = []string{real, tmpDir}
		}
		for _, p := range paths {
			out.Stderr = strings.ReplaceAll(out.Stderr, p+string(os.PathSeparator), "./")
			out.Stderr = strings.ReplaceAll(out.Stderr, p, ".")
		}
	}
	return out, runErr
}

// executeJava は Java を JDK の single-file source launcher（`java Main.java`）で実行する。
// javac の明示コンパイル工程を持たず、メモリ内コンパイルで直接実行される（Java 11+）。
func (r *Runner) executeJava(ctx context.Context, input domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	// 別言語コードが意味不明なコンパイルエラーになるのを避けるため、Java の骨格を必須にする
	// （executeGo の package main チェックと同型のフェイルファスト）。
	if !strings.Contains(input.Code, "class") || !strings.Contains(input.Code, "static void main") {
		return &domain.CodeExecutionResult{
			Stdout:   "",
			Stderr:   "Java コードには `class` と `public static void main(String[] args)` が必要です。",
			ExitCode: 1,
		}, nil
	}

	tmpDir, err := os.MkdirTemp("", "java-exec-")
	if err != nil {
		return nil, fmt.Errorf("一時ディレクトリの作成に失敗: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// source launcher はファイル名とクラス名の一致を要求しない（javac と異なる）ため、
	// 学習者がどんなクラス名を書いても Main.java で受けられる。
	filename := filepath.Join(tmpDir, "Main.java")
	if err := os.WriteFile(filename, []byte(input.Code), 0o600); err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, javaExecTimeout)
	defer cancel()

	// 512MB タスクに同居する他プロセスを圧迫しないようヒープ上限を絞る（学習用途には十分）。
	// SerialGC は極小ヒープ + 単発実行で最速・最小フットプリント。
	cmd := exec.CommandContext(ctx, "java", "-XX:+UseSerialGC", "-Xmx128m", "Main.java")
	cmd.Dir = tmpDir
	// AWS credential や DB password 等を子プロセスに継承しない（JAVA_TOOL_OPTIONS 等の注入も防ぐ）。
	cmd.Env = []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=" + tmpDir,
		"PWD=" + tmpDir,
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
	}

	// ProcessBuilder 等で子孫プロセスを起動しうるので、独立プロセスグループに置いて一括 kill する。
	configureProcessGroup(cmd)

	out, runErr := runCommand(cmd, input.Stdin)
	if out != nil {
		// コンパイルエラーやスタックトレースに出る一時ディレクトリの絶対パスを学習者に見せない。
		out.Stderr = strings.ReplaceAll(out.Stderr, tmpDir+string(os.PathSeparator), "./")
		out.Stderr = strings.ReplaceAll(out.Stderr, tmpDir, ".")
	}
	return out, runErr
}

// pgConn は psql 接続に必要な最小情報。host が "/" 始まりなら unix socket ディレクトリ。
type pgConn struct {
	host, port, user, password string
}

// env は指定 database へ接続する psql 用の最小環境変数を返す（os.Environ は継承しない）。
func (c pgConn) env(database string, extra ...string) []string {
	e := []string{
		"PGHOST=" + c.host,
		"PGPORT=" + c.port,
		"PGUSER=" + c.user,
		"PGDATABASE=" + database,
		"PGCONNECT_TIMEOUT=5",
		// psql 自身と libpq の動作に必要な最小 PATH のみ。秘匿 env は一切渡さない。
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	}
	if c.password != "" {
		e = append(e, "PGPASSWORD="+c.password)
	}
	return append(e, extra...)
}

// sqlPgConfig は同居 PostgreSQL への接続情報を env から組む（既定はコンテナの socket）。
//   - admin: CREATE/DROP DATABASE を行う CREATEDB ロール（superuser ではない）
//   - student: 学習者 SQL を実行する非 superuser ロール
//
// superuser は socket からの接続自体を pg_hba で reject するため、学習者が `\c` 等で
// superuser へ昇格する経路を塞ぐ。admin が CREATEDB の非 superuser でも DB の作成/破棄は可能。
func sqlPgConfig() (admin, student pgConn) {
	host := getenvDefault("CODE_PG_HOST", "/var/run/postgresql")
	port := getenvDefault("CODE_PG_PORT", "5432")
	admin = pgConn{host, port, getenvDefault("CODE_PG_ADMIN", "dbadmin"), os.Getenv("CODE_PG_ADMIN_PASSWORD")}
	student = pgConn{host, port, getenvDefault("CODE_PG_STUDENT", "student"), os.Getenv("CODE_PG_STUDENT_PASSWORD")}
	return admin, student
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// validateSQLInput は危険な psql メタコマンドを含む入力を弾く（多層防御）。空文字なら OK。
func validateSQLInput(code string) string {
	lower := strings.ToLower(code)
	for _, m := range sqlDeniedMeta {
		if strings.Contains(lower, m) {
			return "このメタコマンドは sandbox では使用できません: " + strings.TrimSpace(m)
		}
	}
	return ""
}

// executeSQL は学習者 SQL を、提出ごとの使い捨て DB に対し非 superuser で実行する。
// 同居 PostgreSQL（unix socket 専用 / ネットワーク無し）に対してのみ動き、本番 DB には到達しない。
func (r *Runner) executeSQL(ctx context.Context, input domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	if msg := validateSQLInput(input.Code); msg != "" {
		return &domain.CodeExecutionResult{Stderr: msg, ExitCode: 1}, nil
	}

	admin, student := sqlPgConfig()
	// db 名は識別子として安全な hex のみ（uuid のハイフンを除去）。
	dbName := "s_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	ctx, cancel := context.WithTimeout(ctx, sqlExecTimeout)
	defer cancel()

	// 1) 使い捨て DB を作成（CREATEDB ロール。作成者が owner になる）。
	if err := runPsqlAdmin(ctx, admin, "postgres", `CREATE DATABASE "`+dbName+`"`); err != nil {
		return nil, fmt.Errorf("sql sandbox の準備に失敗: %w", err)
	}
	// 後始末は必ず実行する（ctx が timeout していても別 ctx で DROP する）。
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dcancel()
		_ = runPsqlAdmin(dctx, admin, "postgres", `DROP DATABASE IF EXISTS "`+dbName+`" WITH (FORCE)`)
	}()

	// 2) PG15+ は public スキーマの CREATE 権限が PUBLIC から剥がれているため student に付与。
	//    admin は DB owner（= pg_database_owner）なので public スキーマに GRANT できる。
	if err := runPsqlAdmin(ctx, admin, dbName, `GRANT ALL ON SCHEMA public TO "`+student.user+`"`); err != nil {
		return nil, fmt.Errorf("sql sandbox の権限設定に失敗: %w", err)
	}

	// 3) 学習者 SQL を student で実行。出力は採点が安定するよう unaligned + '|' 区切り + footer 無し。
	tmpDir, err := os.MkdirTemp("", "sql-exec-")
	if err != nil {
		return nil, fmt.Errorf("一時ディレクトリの作成に失敗: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	filename := filepath.Join(tmpDir, "query.sql")
	if err := os.WriteFile(filename, []byte(input.Code), 0o600); err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}

	cmd := exec.CommandContext(
		ctx, "psql",
		"-v", "ON_ERROR_STOP=1",
		"--no-psqlrc", "-q",
		"-A", "-F", "|", "-P", "footer=off",
		"-f", filename,
	)
	cmd.Env = student.env(
		dbName,
		"PGOPTIONS=-c statement_timeout="+sqlStatementTimeoutMS+
			" -c lock_timeout="+sqlLockTimeoutMS+
			" -c idle_in_transaction_session_timeout="+sqlIdleTxTimeoutMS,
	)

	out, runErr := runCommand(cmd, input.Stdin)
	if out != nil {
		// 一時パスを学習者に見せない。
		out.Stderr = strings.ReplaceAll(out.Stderr, filename, "query.sql")
	}
	return out, runErr
}

// runPsqlAdmin は superuser で 1 つの SQL を実行する管理コマンド（CREATE/DROP/GRANT 用）。
func runPsqlAdmin(ctx context.Context, c pgConn, database, sql string) error {
	cmd := exec.CommandContext(
		ctx, "psql",
		"-v", "ON_ERROR_STOP=1", "--no-psqlrc", "-q",
		"-c", sql,
	)
	cmd.Env = c.env(database)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// configureProcessGroup は cmd を独立 process group に置き、cancel 時にグループ全体へ SIGKILL を送る。
// bash など、子孫プロセスを起動しうる言語の leak を防ぐ用途で使う。
func configureProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		// -pid でグループ全体に signal（POSIX）。失敗しても ExitError 上書き防止に nil を返す。
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return nil
	}
	// 残存子孫がいても Wait が張り付かないよう WaitDelay を付ける。
	cmd.WaitDelay = 1 * time.Second
}

func runCommand(cmd *exec.Cmd, stdin string) (*domain.CodeExecutionResult, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	err := cmd.Run()

	out := &domain.CodeExecutionResult{
		Stdout: truncate(stdout.String(), maxOutputBytes),
		Stderr: truncate(stderr.String(), maxOutputBytes),
	}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			out.ExitCode = exitErr.ExitCode()
		} else {
			out.ExitCode = 1
		}
	}
	return out, nil
}

// sandboxEnv は os.Environ() から認証情報・秘密の env を除いて返す（末尾に extra を足す）。
// 特に AWS_CONTAINER_CREDENTIALS_* を落とすことで、ユーザコードが ECS Task Role の
// 一時クレデンシャル取得エンドポイントへ到達する経路を塞ぐ。DATABASE_URL や
// COGNITO_CLIENT_SECRET 等の注入シークレットも読めないようにする。
func sandboxEnv(extra ...string) []string {
	base := os.Environ()
	out := make([]string, 0, len(base)+len(extra))
	for _, kv := range base {
		eq := strings.IndexByte(kv, '=')
		if eq <= 0 {
			continue
		}
		if isSensitiveEnvKey(kv[:eq]) {
			continue
		}
		out = append(out, kv)
	}
	return append(out, extra...)
}

// isSensitiveEnvKey はユーザコードに渡してはいけない env 名かを判定する。
func isSensitiveEnvKey(key string) bool {
	k := strings.ToUpper(key)
	for _, p := range []string{"AWS_", "COGNITO_", "DB_", "DATABASE", "SES_", "DYNAMODB_", "NOTE_IMAGES", "BEDROCK_", "CODE_RUNNER", "CODE_PG", "PG"} {
		if strings.HasPrefix(k, p) {
			return true
		}
	}
	for _, s := range []string{"SECRET", "PASSWORD", "TOKEN", "CREDENTIAL", "_KEY"} {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n...(出力が長すぎるため省略されました)"
}
