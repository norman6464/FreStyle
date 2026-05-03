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
