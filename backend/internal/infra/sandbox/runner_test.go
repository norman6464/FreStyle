package sandbox_test

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_PHP_HelloWorld(t *testing.T) {
	if _, err := exec.LookPath("php"); err != nil {
		t.Skip("php not found in PATH, skipping integration test")
	}

	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `<?php echo "Hello, World!\n";`,
		Language: "php",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

func TestRunner_PHP_SyntaxError(t *testing.T) {
	if _, err := exec.LookPath("php"); err != nil {
		t.Skip("php not found in PATH, skipping integration test")
	}

	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `<?php echo "unclosed`,
		Language: "php",
	})
	require.NoError(t, err)
	assert.NotEqual(t, 0, out.ExitCode)
}

func TestRunner_UnsupportedLanguage(t *testing.T) {
	r := sandbox.NewRunner()
	_, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `print("hi")`,
		Language: "python",
	})
	assert.Error(t, err)
}

func TestRunner_CodeTooLarge(t *testing.T) {
	r := sandbox.NewRunner()
	bigCode := make([]byte, 65*1024)
	_, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     string(bigCode),
		Language: "php",
	})
	assert.Error(t, err)
}

// `<?php` 開始タグが無いコードは、 PHP CLI のデフォルトで「ソースをそのまま stdout に
// 出力して exit 0」になるため、 検証層で弾いて分かりやすい stderr メッセージに置き換える。
func TestRunner_PHP_RejectsCodeWithoutOpenTag(t *testing.T) {
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code: `import java.util.*;
class Main {
    public static void main(String[] args) {
        System.out.println("Hello World");
    }
}`,
		Language: "php",
	})
	require.NoError(t, err)
	assert.Empty(t, out.Stdout)
	assert.NotEqual(t, 0, out.ExitCode)
	assert.Contains(t, out.Stderr, "<?php")
}

// `<?=` (short echo tag) も PHP として扱われる必要があるので、 こちらは通す。
func TestRunner_PHP_AllowsShortEchoTag(t *testing.T) {
	if _, err := exec.LookPath("php"); err != nil {
		t.Skip("php not found in PATH")
	}
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `<?= "Hello"; ?>`,
		Language: "php",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, out.ExitCode)
}

// --- Go ---

func TestRunner_Go_HelloWorld(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found in PATH, skipping integration test")
	}

	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code: `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`,
		Language: "go",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

func TestRunner_Go_RejectsMissingPackageMain(t *testing.T) {
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `fmt.Println("hi")`,
		Language: "go",
	})
	require.NoError(t, err)
	assert.Empty(t, out.Stdout)
	assert.NotEqual(t, 0, out.ExitCode)
	assert.Contains(t, out.Stderr, "package main")
}

func TestRunner_Go_CompileError(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found in PATH")
	}
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code: `package main

func main() {
	undefinedFn()
}
`,
		Language: "go",
	})
	require.NoError(t, err)
	assert.NotEqual(t, 0, out.ExitCode)
	// コンパイルエラーは stderr に出る
	assert.NotEmpty(t, out.Stderr)
}

func TestRunner_Go_ReadsStdin(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found in PATH")
	}
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code: `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		fmt.Println("got:", scanner.Text())
	}
}
`,
		Language: "go",
		Stdin:    "hello\n",
	})
	require.NoError(t, err)
	assert.Equal(t, "got: hello\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

// Warmup(go) はコンパイルキャッシュを温める。go があれば成功する。
func TestRunner_Warmup_Go(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found in PATH")
	}
	r := sandbox.NewRunner()
	require.NoError(t, r.Warmup(context.Background(), "go"))
}

// Warmup(php/bash/未対応) は no-op で常に成功する。
func TestRunner_Warmup_NonGoIsNoop(t *testing.T) {
	r := sandbox.NewRunner()
	require.NoError(t, r.Warmup(context.Background(), "php"))
	require.NoError(t, r.Warmup(context.Background(), "bash"))
	require.NoError(t, r.Warmup(context.Background(), "unknown"))
}

// --- Bash ---

func TestRunner_Bash_HelloWorld(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `echo "Hello, World!"`,
		Language: "bash",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

// HOME / PWD は temp dir に固定されるため、 副作用は外に漏れない。
func TestRunner_Bash_HomeIsTempDir(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `echo "$HOME"`,
		Language: "bash",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, out.ExitCode)
	// macOS は /var/folders/... 、 Linux は /tmp/... と OS により分岐するため、
	// "bash-exec-" prefix が含まれることだけ確認する。
	assert.Contains(t, out.Stdout, "bash-exec-")
}

// AWS / DB の credential が子プロセスに継承されないことを確認する（環境変数を絞っている）。
func TestRunner_Bash_DropsParentEnv(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	t.Setenv("FRESTYLE_SECRET_TEST", "must-not-leak")
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `echo "value=${FRESTYLE_SECRET_TEST:-missing}"`,
		Language: "bash",
	})
	require.NoError(t, err)
	assert.Equal(t, "value=missing\n", out.Stdout)
}

func TestRunner_Bash_ReadsStdin(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `read line && echo "got: $line"`,
		Language: "bash",
		Stdin:    "ping\n",
	})
	require.NoError(t, err)
	assert.Equal(t, "got: ping\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

func TestRunner_Bash_ExitCodePropagated(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	r := sandbox.NewRunner()
	out, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `echo "before" && exit 7`,
		Language: "bash",
	})
	require.NoError(t, err)
	assert.Equal(t, "before\n", out.Stdout)
	assert.Equal(t, 7, out.ExitCode)
}

// timeout 時に bash 配下の子孫プロセスもまとめて kill されることを確認する regression test。
// `Setpgid + cmd.Cancel = group SIGKILL + WaitDelay 1s` で、timeout(1s)+WaitDelay(1s) で
// 必ず数秒以内に return する。
func TestRunner_Bash_TimeoutKillsBackgroundChildren(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	r := sandbox.NewRunner()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	out, err := r.Run(ctx, domain.CodeExecutionInput{
		Code: `sleep 30 &
sleep 30
`,
		Language: "bash",
	})
	elapsed := time.Since(start)
	require.NoError(t, err)
	assert.NotZero(t, out.ExitCode, "process should be killed by timeout")
	assert.Less(t, elapsed, 5*time.Second, "should NOT wait for orphan sleep child to exit")
}
