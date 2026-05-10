package usecase_test

import (
	"context"
	"os/exec"
	"testing"

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
