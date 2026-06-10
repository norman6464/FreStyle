// Package sandbox は学習者コード（php / go / bash）を os/exec でサンドボックス実行する
// in-process 実装を提供する。backend 本体に同居させる場合（CODE_RUNNER_URL 未設定）と、
// 別バイナリ cmd/coderunner（HTTP サーバ）の中身の双方から共有して使う。
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
	// execTimeout は PHP / bash の上限。
	execTimeout = 8 * time.Second
	// goExecTimeout は Go 専用。go run はコンパイルも走るため余裕を持たせる（cold compile ~6-8s）。
	goExecTimeout  = 15 * time.Second
	maxOutputBytes = 64 * 1024 // 64 KB
)

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

// Runner は php / go / bash を os/exec でサンドボックス実行する in-process 実装。
// 共通制約: timeout / 64 KB code-size / 64 KB output-size。
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
	default:
		return nil, fmt.Errorf("未対応の言語: %s", input.Language)
	}
}

// Warmup は実行環境を事前に温める。Go は go run のコンパイルキャッシュを温めるため
// trivial なプログラムを 1 回コンパイル/実行する。php / bash は起動が軽量なので no-op。
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

	return runCommand(cmd, input.Stdin)
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
	for _, p := range []string{"AWS_", "COGNITO_", "DB_", "DATABASE", "SES_", "DYNAMODB_", "NOTE_IMAGES", "BEDROCK_", "CODE_RUNNER"} {
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
