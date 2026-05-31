package usecase

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
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

// ExecuteCodeInput はコード実行の入力。
type ExecuteCodeInput struct {
	Code     string
	Language string // "php" / "go" / "bash"
	// Stdin は実行時に標準入力として流す内容（テストケース採点で使う）。
	Stdin string
}

// ExecuteCodeOutput はコード実行の結果。
type ExecuteCodeOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

// ExecuteCodeUseCase は PHP / Go / bash のコードをサンドボックスで実行する。
// 共通制約: timeout / 64 KB code-size / 64 KB output-size。
type ExecuteCodeUseCase struct{}

func NewExecuteCodeUseCase() *ExecuteCodeUseCase {
	return &ExecuteCodeUseCase{}
}

func (uc *ExecuteCodeUseCase) Execute(ctx context.Context, input ExecuteCodeInput) (*ExecuteCodeOutput, error) {
	if len(input.Code) > maxCodeBytes {
		return nil, fmt.Errorf("コードが大きすぎます (最大 64 KB)")
	}
	switch input.Language {
	case "php":
		return uc.executePHP(ctx, input)
	case "go":
		return uc.executeGo(ctx, input)
	case "bash":
		return uc.executeBash(ctx, input)
	default:
		return nil, fmt.Errorf("未対応の言語: %s", input.Language)
	}
}

func (uc *ExecuteCodeUseCase) executePHP(ctx context.Context, input ExecuteCodeInput) (*ExecuteCodeOutput, error) {
	// 開始タグが無いと PHP はソースをそのまま stdout に流すため、別言語の誤貼付けを明示的に弾く。
	if !strings.Contains(input.Code, "<?php") && !strings.Contains(input.Code, "<?=") {
		return &ExecuteCodeOutput{
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

	cmd := exec.CommandContext(ctx,
		"php",
		"-d", "max_execution_time=5",
		"-d", "memory_limit=32M",
		"-d", "disable_functions="+disableFunctions,
		"-d", "open_basedir="+tmpDir,
		"-d", "display_errors=stderr",
		"-d", "log_errors=0",
		filename,
	)

	return runCommand(cmd, input.Stdin)
}

func (uc *ExecuteCodeUseCase) executeGo(ctx context.Context, input ExecuteCodeInput) (*ExecuteCodeOutput, error) {
	// 別言語コードが意味不明なコンパイルエラーになるのを避けるため package main を必須にする。
	if !strings.Contains(input.Code, "package main") {
		return &ExecuteCodeOutput{
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
	// GO111MODULE=off で go.mod 不要の単一ファイル実行。
	cmd.Env = append(os.Environ(),
		"GO111MODULE=off",
		"HOME="+tmpDir,
	)

	return runCommand(cmd, input.Stdin)
}

func (uc *ExecuteCodeUseCase) executeBash(ctx context.Context, input ExecuteCodeInput) (*ExecuteCodeOutput, error) {
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
// bash の子孫プロセス leak を防ぐ用途で、bash 実行時のみ使う。
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

func runCommand(cmd *exec.Cmd, stdin string) (*ExecuteCodeOutput, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	err := cmd.Run()

	out := &ExecuteCodeOutput{
		Stdout: truncate(stdout.String(), maxOutputBytes),
		Stderr: truncate(stderr.String(), maxOutputBytes),
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			out.ExitCode = exitErr.ExitCode()
		} else {
			out.ExitCode = 1
		}
	}
	return out, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n...(出力が長すぎるため省略されました)"
}
