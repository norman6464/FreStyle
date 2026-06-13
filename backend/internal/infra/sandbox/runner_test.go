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

func Test_ランナー_PHP_HelloWorld(t *testing.T) {
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

func Test_ランナー_PHP_構文エラー(t *testing.T) {
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

func Test_ランナー_非対応言語(t *testing.T) {
	r := sandbox.NewRunner()
	_, err := r.Run(context.Background(), domain.CodeExecutionInput{
		Code:     `print("hi")`,
		Language: "python",
	})
	assert.Error(t, err)
}

func Test_ランナー_コードが大きすぎる(t *testing.T) {
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
func Test_ランナー_PHP_開始タグなしを拒否(t *testing.T) {
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
func Test_ランナー_PHP_短縮echoタグを許可(t *testing.T) {
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

func Test_ランナー_Go_HelloWorld(t *testing.T) {
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

func Test_ランナー_Go_package_main欠落を拒否(t *testing.T) {
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

func Test_ランナー_Go_コンパイルエラー(t *testing.T) {
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
	// 一時ディレクトリの内部パスを露出せず、`./main.go` に整形される。
	assert.NotContains(t, out.Stderr, "go-exec-")
	assert.Contains(t, out.Stderr, "./main.go")
}

func Test_ランナー_Go_標準入力を読む(t *testing.T) {
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
func Test_ランナー_ウォームアップ_Go(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not found in PATH")
	}
	r := sandbox.NewRunner()
	require.NoError(t, r.Warmup(context.Background(), "go"))
}

// Warmup(php/bash/未対応) は no-op で常に成功する。
func Test_ランナー_ウォームアップ_Go以外は何もしない(t *testing.T) {
	r := sandbox.NewRunner()
	require.NoError(t, r.Warmup(context.Background(), "php"))
	require.NoError(t, r.Warmup(context.Background(), "bash"))
	require.NoError(t, r.Warmup(context.Background(), "unknown"))
}

// --- Bash ---

func Test_ランナー_Bash_HelloWorld(t *testing.T) {
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
func Test_ランナー_Bash_HOMEは一時ディレクトリ(t *testing.T) {
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
func Test_ランナー_Bash_親envを落とす(t *testing.T) {
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

func Test_ランナー_Bash_標準入力を読む(t *testing.T) {
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

func Test_ランナー_Bash_終了コードを伝播(t *testing.T) {
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
func Test_ランナー_Bash_タイムアウトで子プロセスを停止(t *testing.T) {
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
