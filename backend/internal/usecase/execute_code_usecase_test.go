package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCodeRunner は CodeRunner を満たすテスト用 fake。
type fakeCodeRunner struct {
	gotInput  domain.CodeExecutionInput
	gotWarmup string
	result    *domain.CodeExecutionResult
	runErr    error
	warmupErr error
}

func (f *fakeCodeRunner) Run(_ context.Context, in domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	f.gotInput = in
	if f.runErr != nil {
		return nil, f.runErr
	}
	return f.result, nil
}

func (f *fakeCodeRunner) Warmup(_ context.Context, language string) error {
	f.gotWarmup = language
	return f.warmupErr
}

func TestExecuteCodeUseCase_DelegatesToRunner(t *testing.T) {
	fake := &fakeCodeRunner{result: &domain.CodeExecutionResult{Stdout: "ok", ExitCode: 0}}
	uc := usecase.NewExecuteCodeUseCase(fake)

	out, err := uc.Execute(context.Background(), domain.CodeExecutionInput{
		Code:     `<?php echo "ok";`,
		Language: "php",
		Stdin:    "x",
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", out.Stdout)
	// 入力がそのまま runner に渡る。
	assert.Equal(t, "php", fake.gotInput.Language)
	assert.Equal(t, "x", fake.gotInput.Stdin)
}

func TestExecuteCodeUseCase_PropagatesRunnerError(t *testing.T) {
	fake := &fakeCodeRunner{runErr: errors.New("boom")}
	uc := usecase.NewExecuteCodeUseCase(fake)

	_, err := uc.Execute(context.Background(), domain.CodeExecutionInput{Language: "go"})
	assert.Error(t, err)
}

func TestWarmupCodeUseCase_DelegatesLanguage(t *testing.T) {
	fake := &fakeCodeRunner{}
	uc := usecase.NewWarmupCodeUseCase(fake)

	require.NoError(t, uc.Execute(context.Background(), "go"))
	assert.Equal(t, "go", fake.gotWarmup)
}

func TestWarmupCodeUseCase_PropagatesError(t *testing.T) {
	fake := &fakeCodeRunner{warmupErr: errors.New("warm failed")}
	uc := usecase.NewWarmupCodeUseCase(fake)

	assert.Error(t, uc.Execute(context.Background(), "go"))
}
