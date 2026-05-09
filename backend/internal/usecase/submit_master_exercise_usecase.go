package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// CodeExecutor は ExecuteCodeUseCase を抽象化したインターフェイス。
// usecase 層が usecase 同士に直接依存するのを避け、テストでは fake に差し替える。
type CodeExecutor interface {
	Execute(ctx context.Context, in ExecuteCodeInput) (*ExecuteCodeOutput, error)
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

// SubmitMasterExerciseUseCase はユーザコードを slug の master_exercise に対して
// 採点・提出履歴に保存する。
//
// 採点ロジック:
//   - examples を全件取得（OrderIndex 昇順）
//   - 各 example の InputText を stdin に流して ExecuteCodeUseCase で実行
//   - stdout を normalize（末尾改行 / CRLF を吸収）して expected_output と完全一致するか比較
//   - 全件 pass で isCorrect=true。1 件でも失敗・実行エラー（exit_code != 0）なら false
//   - 実行コスト削減のため、最初の不一致で打ち切らず全件実行（ユーザに「どこで落ちたか」を全部見せる方針）
//
// 履歴保存はテストケース個別ではなく、まとめて 1 行（最初に失敗した stdout / stderr / exit_code を採用）。
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

	ex, err := uc.exercises.GetBySlug(in.Slug)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, fmt.Errorf("exercise not found: %s", in.Slug)
	}
	examples, err := uc.examples.ListByExerciseID(ex.ID)
	if err != nil {
		return nil, err
	}
	if len(examples) == 0 {
		return nil, fmt.Errorf("exercise %s has no test cases", in.Slug)
	}

	results := make([]TestCaseResult, 0, len(examples))
	allPassed := true
	// 履歴保存用に「最初に失敗した実行結果」を保持。 全件 pass のときは最後の結果を使う。
	var representativeStdout, representativeStderr string
	var representativeExit int

	for _, ex := range examples {
		out, err := uc.executor.Execute(ctx, ExecuteCodeInput{
			Code:     in.Code,
			Language: "php",
			Stdin:    ex.InputText,
		})
		if err != nil {
			// 実行できなかった場合は提出を保存せずエラーを返す（タイムアウト等）。
			return nil, fmt.Errorf("execution failed: %w", err)
		}
		actual := normalizeOutput(out.Stdout)
		expected := normalizeOutput(ex.ExpectedOutput)
		passed := out.ExitCode == 0 && actual == expected
		results = append(results, TestCaseResult{
			OrderIndex:     ex.OrderIndex,
			Input:          ex.InputText,
			ExpectedOutput: ex.ExpectedOutput,
			ActualOutput:   out.Stdout,
			Stderr:         out.Stderr,
			Passed:         passed,
		})
		if !passed && allPassed {
			// 最初の失敗を representative として記録し allPassed を false に倒す。
			// 以降のループでは下の if も通らないため、 この値が representative として固定される。
			representativeStdout = out.Stdout
			representativeStderr = out.Stderr
			representativeExit = out.ExitCode
			allPassed = false
		}
		if allPassed {
			// 全件 pass し続けている経路。 ループのたびに最後の出力で上書きすることで、
			// 最終的に representative が「最後に通ったテストケースの出力」になる。
			// 1 件でも失敗したら上の if 文で allPassed=false に倒れ、 こちらは以降スキップされる。
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
	if err := uc.submissions.Create(submission); err != nil {
		return nil, err
	}

	return &SubmitMasterExerciseOutput{
		SubmissionID: submission.ID,
		IsCorrect:    allPassed,
		Results:      results,
	}, nil
}

// normalizeOutput は stdout 比較時の表記揺れを吸収する。
//   - CRLF / CR を LF に統一
//   - 末尾の改行・空白をすべて除去（"42\n" と "42" を同一視）
//
// PHP の `echo` は末尾改行を付けない / 付けるが問題により混在するため、
// 「末尾だけ揺れを許容」する方針で十分。 行内の空白は厳密一致。
func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.TrimRight(s, " \t\n")
}
