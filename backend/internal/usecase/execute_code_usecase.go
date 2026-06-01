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
	// execTimeout は PHP / bash の デフォルト 上限。 起動 + 実行が ~ms オーダー。
	execTimeout = 8 * time.Second
	// goExecTimeout は Go 専用。 `go run` は コンパイル + リンク + 実行 が走るため、
	// ECS Fargate の ephemeral storage I/O 遅さも勘案して 余裕を持たせる。
	// production の cold compile は ~6-8s の実測がある。
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

// ExecuteCodeUseCase はユーザーコードをサンドボックスで実行する。
//
// PHP / Go / bash の 3 言語サポート:
//   - PHP: php CLI + disable_functions で危険関数を封じる
//   - Go: go run で単一 main.go ファイルを実行。 build cache は共有して compile 時間短縮
//   - bash: /bin/bash でシェルスクリプトを実行。 PATH を限定し HOME を temp dir に固定
//
// 共通制約: 8 秒 timeout / 64 KB code-size / 64 KB output-size。
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
	// PHP CLI は `<?php` 開始タグが無いコードを「ただの HTML テキスト」として扱い、
	// ソースコードをそのまま stdout に流して exit code 0 を返す。
	// 別言語 (Java など) のコードを誤って貼り付けると「実行成功 + 出力 = 元コード」になり
	// ユーザを混乱させるので、 ここで明示的に弾く。
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
		// variables_order から E を外し $_ENV を空にする（disable_functions=getenv の抜け穴を塞ぐ）。
		"-d", "variables_order=GPCS",
		filename,
	)
	// AWS 認証情報・注入シークレットを子プロセスに渡さない。
	cmd.Env = sandboxEnv()

	return runCommand(cmd, input.Stdin)
}

func (uc *ExecuteCodeUseCase) executeGo(ctx context.Context, input ExecuteCodeInput) (*ExecuteCodeOutput, error) {
	// Go コードは `package main` + `func main()` を必須にする。
	// 別言語のコードを「Go モード」 で投げて意味不明なコンパイルエラーになるのを避ける。
	if !strings.Contains(input.Code, "package main") {
		return &ExecuteCodeOutput{
			Stdout:   "",
			Stderr:   "Go コードには `package main` と `func main()` が必要です。",
			ExitCode: 1,
		}, nil
	}

	// 各リクエストごとに独立した GOPATH 用ディレクトリを作る。
	// build cache は共有 (/tmp/go-build-cache) で compile 時間短縮。
	tmpDir, err := os.MkdirTemp("", "go-exec-")
	if err != nil {
		return nil, fmt.Errorf("一時ディレクトリの作成に失敗: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(filename, []byte(input.Code), 0o600); err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}

	// Go は コンパイル + 実行 が走るため timeout を 専用の 長めの値で取る。
	ctx, cancel := context.WithTimeout(ctx, goExecTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", filename)
	// GO111MODULE=off で単一ファイル実行（go.mod 不要）。GOCACHE 等の go 動作に必要な env は
	// sandboxEnv が残すが、AWS 認証情報・注入シークレットは除外する。
	cmd.Env = sandboxEnv(
		"GO111MODULE=off",
		"HOME="+tmpDir,
	)

	return runCommand(cmd, input.Stdin)
}

func (uc *ExecuteCodeUseCase) executeBash(ctx context.Context, input ExecuteCodeInput) (*ExecuteCodeOutput, error) {
	// 各リクエストごとに独立した一時ディレクトリを作る。
	// 演習で `cat > /tmp/sample.txt` のような書き込みが起きても他リクエストに影響させないため
	// HOME / PWD を temp dir に固定し、 副作用を temp に閉じる。
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

	// bash はユーザが `sleep 30 &` のような バックグラウンド子プロセスを 起動しうる。
	// timeout 時にそれら子孫プロセスを 確実に kill するため、 独立プロセスグループに置く。
	configureProcessGroup(cmd)

	return runCommand(cmd, input.Stdin)
}

// configureProcessGroup は cmd を独立した process group に置き、 ctx cancel 時に
// グループ全体へ SIGKILL を送るよう設定する。 bash で バックグラウンド子孫が leak しないよう
// に bash 実行のみで使う。 PHP / Go の `go run` は普通 子孫を起動しないため不要。
func configureProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		// -pid を渡すと プロセスグループ全体に signal が飛ぶ (POSIX)。
		// kill が失敗しても (既に終了済 / ESRCH) ExitError 上書きを防ぐため nil を返す。
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return nil
	}
	// stdout / stderr のパイプが 残ったままの子孫がいても Wait が 無限に張り付かないよう、
	// 小さい WaitDelay を付ける。 timeout 後 1 秒で 強制 close + kill。
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
	for _, p := range []string{"AWS_", "COGNITO_", "DB_", "DATABASE", "SES_", "DYNAMODB_", "NOTE_IMAGES", "BEDROCK_"} {
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
