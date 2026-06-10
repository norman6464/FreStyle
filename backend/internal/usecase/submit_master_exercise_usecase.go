package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CodeExecutor は ExecuteCodeUseCase を抽象化し、usecase 同士の直接依存を避ける。
type CodeExecutor interface {
	Execute(ctx context.Context, in domain.CodeExecutionInput) (*domain.CodeExecutionResult, error)
}

// SubmitMasterExerciseInput は提出 API への入力。
type SubmitMasterExerciseInput struct {
	UserID uint64
	Slug   string
	Code   string
}

// TestCaseResult はテストケース 1 件の採点結果。
type TestCaseResult struct {
	OrderIndex     int16  `json:"orderIndex"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expectedOutput"`
	ActualOutput   string `json:"actualOutput"`
	Stderr         string `json:"stderr"`
	Passed         bool   `json:"passed"`
}

// SubmitMasterExerciseOutput は提出 API の戻り値。
type SubmitMasterExerciseOutput struct {
	SubmissionID uint64           `json:"submissionId"`
	IsCorrect    bool             `json:"isCorrect"`
	Results      []TestCaseResult `json:"results"`
}

// SubmitMasterExerciseUseCase はユーザコードを slug の master_exercise に対して採点し履歴に保存する。
//
// 採点は mode で分岐する:
//   - qa: 実行せず提出文字列と ExpectedOutput を normalize 比較するだけ。
//   - execute: examples（無ければ exercise 自身の ExpectedOutput を単一ケース化）を全件実行し、
//     stdout を normalize 比較。全件 pass かつ exit_code 0 で isCorrect=true。
//     どこで落ちたか全部見せるため最初の不一致で打ち切らず全件実行する。
//
// 履歴は 1 行にまとめて保存（失敗時は最初の失敗、成功時は最後の実行結果を採用）。
type SubmitMasterExerciseUseCase struct {
	exercises   repository.MasterExerciseRepository
	examples    repository.MasterExerciseExampleRepository
	submissions repository.ExerciseSubmissionRepository
	executor    CodeExecutor
}

func NewSubmitMasterExerciseUseCase(
	exercises repository.MasterExerciseRepository,
	examples repository.MasterExerciseExampleRepository,
	submissions repository.ExerciseSubmissionRepository,
	executor CodeExecutor,
) *SubmitMasterExerciseUseCase {
	return &SubmitMasterExerciseUseCase{
		exercises:   exercises,
		examples:    examples,
		submissions: submissions,
		executor:    executor,
	}
}

func (uc *SubmitMasterExerciseUseCase) Execute(ctx context.Context, in SubmitMasterExerciseInput) (*SubmitMasterExerciseOutput, error) {
	if in.UserID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	if strings.TrimSpace(in.Slug) == "" {
		return nil, fmt.Errorf("slug is required")
	}
	if strings.TrimSpace(in.Code) == "" {
		return nil, fmt.Errorf("code is required")
	}

	ex, err := uc.exercises.GetBySlug(ctx, in.Slug)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, fmt.Errorf("exercise not found: %s", in.Slug)
	}

	if ex.Mode == domain.ExerciseModeQA {
		return uc.submitQA(ctx, in, ex)
	}

	examples, err := uc.examples.ListByExerciseID(ctx, ex.ID)
	if err != nil {
		return nil, err
	}
	// examples が無い演習は exercise 自身の ExpectedOutput を単一の仮想テストケースとして使う。
	if len(examples) == 0 {
		examples = []domain.MasterExerciseExample{{
			ExerciseID:     ex.ID,
			OrderIndex:     1,
			InputText:      "",
			ExpectedOutput: ex.ExpectedOutput,
		}}
	}

	results := make([]TestCaseResult, 0, len(examples))
	allPassed := true
	// 履歴保存用の代表結果（失敗時は最初の失敗、全件 pass 時は最後の結果）。
	var representativeStdout, representativeStderr string
	var representativeExit int

	for _, tc := range examples {
		out, err := uc.executor.Execute(ctx, domain.CodeExecutionInput{
			Code:     in.Code,
			Language: ex.Language,
			Stdin:    tc.InputText,
		})
		if err != nil {
			// 実行できなかった場合は提出を保存せずエラーを返す（タイムアウト等）。
			return nil, fmt.Errorf("execution failed: %w", err)
		}
		actual := normalizeOutput(out.Stdout)
		expected := normalizeOutput(tc.ExpectedOutput)
		passed := out.ExitCode == 0 && actual == expected
		results = append(results, TestCaseResult{
			OrderIndex:     tc.OrderIndex,
			Input:          tc.InputText,
			ExpectedOutput: tc.ExpectedOutput,
			ActualOutput:   out.Stdout,
			Stderr:         out.Stderr,
			Passed:         passed,
		})
		if !passed && allPassed {
			// 最初の失敗を representative に固定する（以降は下の if も通らない）。
			representativeStdout = out.Stdout
			representativeStderr = out.Stderr
			representativeExit = out.ExitCode
			allPassed = false
		}
		if allPassed {
			// 全件 pass の間は最後の出力で上書きし続ける。
			representativeStdout = out.Stdout
			representativeStderr = out.Stderr
			representativeExit = out.ExitCode
		}
	}

	submission := &domain.ExerciseSubmission{
		UserID:        in.UserID,
		ExerciseKind:  domain.ExerciseKindMaster,
		ExerciseID:    ex.ID,
		SubmittedCode: in.Code,
		Stdout:        representativeStdout,
		Stderr:        representativeStderr,
		ExitCode:      representativeExit,
		IsCorrect:     allPassed,
		SubmittedAt:   time.Now().UTC(),
	}
	if err := uc.submissions.Create(ctx, submission); err != nil {
		return nil, err
	}

	return &SubmitMasterExerciseOutput{
		SubmissionID: submission.ID,
		IsCorrect:    allPassed,
		Results:      results,
	}, nil
}

// submitQA は QA モードの採点。コード実行せず提出文字列と ExpectedOutput を normalize 比較する。
func (uc *SubmitMasterExerciseUseCase) submitQA(ctx context.Context, in SubmitMasterExerciseInput, ex *domain.MasterExercise) (*SubmitMasterExerciseOutput, error) {
	expected := normalizeOutput(ex.ExpectedOutput)
	actual := normalizeOutput(in.Code)
	isCorrect := actual == expected

	submission := &domain.ExerciseSubmission{
		UserID:        in.UserID,
		ExerciseKind:  domain.ExerciseKindMaster,
		ExerciseID:    ex.ID,
		SubmittedCode: in.Code,
		Stdout:        in.Code,
		Stderr:        "",
		ExitCode:      0,
		IsCorrect:     isCorrect,
		SubmittedAt:   time.Now().UTC(),
	}
	if err := uc.submissions.Create(ctx, submission); err != nil {
		return nil, err
	}

	return &SubmitMasterExerciseOutput{
		SubmissionID: submission.ID,
		IsCorrect:    isCorrect,
		Results: []TestCaseResult{{
			OrderIndex:     1,
			Input:          "",
			ExpectedOutput: ex.ExpectedOutput,
			ActualOutput:   in.Code,
			Stderr:         "",
			Passed:         isCorrect,
		}},
	}, nil
}

// normalizeOutput は CRLF/CR を LF に統一し末尾の改行・空白を除去する（行内の空白は厳密一致）。
func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.TrimRight(s, " \t\n")
}
