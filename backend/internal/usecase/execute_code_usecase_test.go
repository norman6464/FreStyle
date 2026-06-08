package usecase_test

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCodeUseCase_PHP_HelloWorld(t *testing.T) {
	_, err := exec.LookPath("php")
	if err != nil {
		t.Skip("php not found in PATH, skipping integration test")
	}

	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `<?php echo "Hello, World!\n";`,
		Language: "php",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

func TestExecuteCodeUseCase_PHP_SyntaxError(t *testing.T) {
	_, err := exec.LookPath("php")
	if err != nil {
		t.Skip("php not found in PATH, skipping integration test")
	}

	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `<?php echo "unclosed`,
		Language: "php",
	})
	require.NoError(t, err)
	assert.NotEqual(t, 0, out.ExitCode)
}

func TestExecuteCodeUseCase_UnsupportedLanguage(t *testing.T) {
	uc := usecase.NewExecuteCodeUseCase()
	_, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `print("hi")`,
		Language: "python",
	})
	assert.Error(t, err)
}

func TestExecuteCodeUseCase_CodeTooLarge(t *testing.T) {
	uc := usecase.NewExecuteCodeUseCase()
	bigCode := make([]byte, 65*1024)
	_, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     string(bigCode),
		Language: "php",
	})
	assert.Error(t, err)
}

// `<?php` 開始タグが無いコード（例: Java を貼り付けた）は、 PHP CLI のデフォルトで
// 「ソースをそのまま stdout に出力して exit 0」になるため、 検証層で弾いて分かりやすい
// stderr メッセージに置き換える。
func TestExecuteCodeUseCase_PHP_RejectsCodeWithoutOpenTag(t *testing.T) {
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
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
func TestExecuteCodeUseCase_PHP_AllowsShortEchoTag(t *testing.T) {
	_, err := exec.LookPath("php")
	if err != nil {
		t.Skip("php not found in PATH")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `<?= "Hello"; ?>`,
		Language: "php",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, out.ExitCode)
}

// --- Go ---

func TestExecuteCodeUseCase_Go_HelloWorld(t *testing.T) {
	_, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not found in PATH, skipping integration test")
	}

	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
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

func TestExecuteCodeUseCase_Go_RejectsMissingPackageMain(t *testing.T) {
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `fmt.Println("hi")`,
		Language: "go",
	})
	require.NoError(t, err)
	assert.Empty(t, out.Stdout)
	assert.NotEqual(t, 0, out.ExitCode)
	assert.Contains(t, out.Stderr, "package main")
}

func TestExecuteCodeUseCase_Go_CompileError(t *testing.T) {
	_, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not found in PATH")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
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

func TestExecuteCodeUseCase_Go_ReadsStdin(t *testing.T) {
	_, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not found in PATH")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
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

// --- Bash ---

func TestExecuteCodeUseCase_Bash_HelloWorld(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `echo "Hello, World!"`,
		Language: "bash",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

// HOME / PWD は temp dir に固定されるため、 副作用は外に漏れない。
// echo "$HOME" の結果が `/tmp/...` 始まりであることを確認する。
func TestExecuteCodeUseCase_Bash_HomeIsTempDir(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
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
func TestExecuteCodeUseCase_Bash_DropsParentEnv(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	t.Setenv("FRESTYLE_SECRET_TEST", "must-not-leak")
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `echo "value=${FRESTYLE_SECRET_TEST:-missing}"`,
		Language: "bash",
	})
	require.NoError(t, err)
	assert.Equal(t, "value=missing\n", out.Stdout)
}

func TestExecuteCodeUseCase_Bash_ReadsStdin(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `read line && echo "got: $line"`,
		Language: "bash",
		Stdin:    "ping\n",
	})
	require.NoError(t, err)
	assert.Equal(t, "got: ping\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

func TestExecuteCodeUseCase_Bash_ExitCodePropagated(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `echo "before" && exit 7`,
		Language: "bash",
	})
	require.NoError(t, err)
	assert.Equal(t, "before\n", out.Stdout)
	assert.Equal(t, 7, out.ExitCode)
}

// timeout 時に bash 配下の子孫プロセスもまとめて kill されることを確認する regression test。
//
// 既定の `exec.CommandContext` は親 bash のみに SIGKILL を投げるため、 ユーザコードが
// バックグラウンドで投げた子孫 (例: sleep 30 &) が stdout パイプを保持し続け、
// `cmd.Wait()` が 30 秒間ブロックする問題があった。
//
// `runCommand` で `Setpgid + cmd.Cancel = group SIGKILL + WaitDelay 1s` を入れたので、
// timeout (1 秒) + WaitDelay (1 秒) で必ず数秒以内に return するはず。
func TestExecuteCodeUseCase_Bash_TimeoutKillsBackgroundChildren(t *testing.T) {
	if _, err := exec.LookPath("/bin/bash"); err != nil {
		t.Skip("/bin/bash not found")
	}
	uc := usecase.NewExecuteCodeUseCase()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	out, err := uc.Execute(ctx, usecase.ExecuteCodeInput{
		Code: `sleep 30 &
sleep 30
`,
		Language: "bash",
	})
	elapsed := time.Since(start)
	require.NoError(t, err)
	assert.NotZero(t, out.ExitCode, "process should be killed by timeout")
	// 修正前: 30 秒 ブロック / 修正後: timeout 1s + WaitDelay 1s で 数秒以内に return。
	assert.Less(t, elapsed, 5*time.Second, "should NOT wait for orphan sleep child to exit")
}

// --- Java ---

func TestExecuteCodeUseCase_Java_HelloWorld(t *testing.T) {
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java not found in PATH, skipping integration test")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code: `public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}
`,
		Language: "java",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

func TestExecuteCodeUseCase_Java_RejectsMissingMainClass(t *testing.T) {
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code:     `System.out.println("hi");`,
		Language: "java",
	})
	require.NoError(t, err)
	assert.Empty(t, out.Stdout)
	assert.NotEqual(t, 0, out.ExitCode)
	assert.Contains(t, out.Stderr, "class Main")
}

func TestExecuteCodeUseCase_Java_CompileError(t *testing.T) {
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java not found in PATH")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code: `public class Main {
    public static void main(String[] args) {
        undefinedMethod();
    }
}
`,
		Language: "java",
	})
	require.NoError(t, err)
	assert.NotEqual(t, 0, out.ExitCode)
	// コンパイルエラーは stderr に出る
	assert.NotEmpty(t, out.Stderr)
}

func TestExecuteCodeUseCase_Java_ReadsStdin(t *testing.T) {
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java not found in PATH")
	}
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code: `import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        String line = sc.nextLine();
        System.out.println("got: " + line);
    }
}
`,
		Language: "java",
		Stdin:    "hello\n",
	})
	require.NoError(t, err)
	assert.Equal(t, "got: hello\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}

// AWS / DB の credential が子プロセスに継承されないことを確認する（環境変数を絞っている）。
func TestExecuteCodeUseCase_Java_DropsParentEnv(t *testing.T) {
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java not found in PATH")
	}
	t.Setenv("AWS_SECRET_ACCESS_KEY", "must-not-leak")
	uc := usecase.NewExecuteCodeUseCase()
	out, err := uc.Execute(context.Background(), usecase.ExecuteCodeInput{
		Code: `public class Main {
    public static void main(String[] args) {
        String v = System.getenv("AWS_SECRET_ACCESS_KEY");
        System.out.println("value=" + (v == null ? "missing" : v));
    }
}
`,
		Language: "java",
	})
	require.NoError(t, err)
	assert.Equal(t, "value=missing\n", out.Stdout)
	assert.Equal(t, 0, out.ExitCode)
}
