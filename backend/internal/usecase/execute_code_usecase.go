package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CodeRunner は php / go / bash のコード実行とウォームアップを抽象化する port。
// in-process 実装（infra/sandbox.Runner）か HTTP クライアント（infra/coderunner.Client）を
// router が CODE_RUNNER_URL の有無で注入する。テストでは fake を差し替える。
type CodeRunner interface {
	Run(ctx context.Context, in domain.CodeExecutionInput) (*domain.CodeExecutionResult, error)
	Warmup(ctx context.Context, language string) error
}

// ExecuteCodeUseCase は学習者コードをサンドボックスで実行する。実行自体は CodeRunner に委譲し、
// usecase は実行系（in-process / sidecar）に依存しない。
type ExecuteCodeUseCase struct {
	runner CodeRunner
}

// NewExecuteCodeUseCase は CodeRunner を注入して ExecuteCodeUseCase を返す。
func NewExecuteCodeUseCase(runner CodeRunner) *ExecuteCodeUseCase {
	return &ExecuteCodeUseCase{runner: runner}
}

// Execute は入力コードを CodeRunner で実行し結果を返す。
func (uc *ExecuteCodeUseCase) Execute(ctx context.Context, in domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	return uc.runner.Run(ctx, in)
}
