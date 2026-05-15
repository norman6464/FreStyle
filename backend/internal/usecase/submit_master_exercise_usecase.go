package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/legacyrepository"
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
//   - mode='qa' の場合: コードを実行せず、 提出文字列と ExpectedOutput を trim 比較するだけ。
//     docker / kubernetes など サンドボックス実行が困難な題材向け。
//   - mode='execute' (default) の場合:
//   - master_exercise_examples が登録されていれば 各 example の InputText を stdin に
//     流して ExecuteCodeUseCase で実行し、 stdout を normalize して expected_output と比較
//   - examples が無ければ exercise 自身の ExpectedOutput を 単一テストケースとして使う
//     (linux / git / go 等の seed.py 経由で投入された演習向け)
//     stdout を normalize（末尾改行 / CRLF を吸収）して expected_output と完全一致するか比較。
//     全件 pass で isCorrect=true。1 件でも失敗・実行エラー（exit_code != 0）なら false。
//     実行コスト削減のため、最初の不一致で打ち切らず全件実行（ユーザに「どこで落ちたか」を全部見せる方針）。
//
// 履歴保存はテストケース個別ではなく、まとめて 1 行（最初に失敗した stdout / stderr / exit_code を採用）。
type SubmitMasterExerciseUseCase struct {
	exercises   legacyrepository.MasterExerciseRepository
	examples    legacyrepository.MasterExerciseExampleRepository
	submissions legacyrepository.ExerciseSubmissionRepository
	executor    CodeExecutor
}

func NewSubmitMasterExerciseUseCase(
	exercises legacyrepository.MasterExerciseRepository,
	examples legacyrepository.MasterExerciseExampleRepository,
	submissions legacyrepository.ExerciseSubmissionRepository,
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

	// QA モード: コード実行せず、 提出文字列と ExpectedOutput を直接比較する。
	if ex.Mode == domain.ExerciseModeQA {
		return uc.submitQA(in, ex)
	}

	examples, err := uc.examples.ListByExerciseID(ex.ID)
	if err != nil {
		return nil, err
	}
	// examples が登録されていない演習 (seed.py 経由で投入された linux / git / go 等)
	// は exercise 自身の ExpectedOutput を 単一の仮想テストケースとして使う。
	// stdin は空、 OrderIndex=1 として扱う。
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
	// 履歴保存用に「最初に失敗した実行結果」を保持。 全件 pass のときは最後の結果を使う。
	var representativeStdout, representativeStderr string
	var representativeExit int

	for _, tc := range examples {
		out, err := uc.executor.Execute(ctx, ExecuteCodeInput{
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

// submitQA は QA モードの採点。 提出文字列と ExpectedOutput を normalize 比較するだけで
// コード実行は行わない。 docker / kubernetes など サンドボックス実行が困難な題材向け。
func (uc *SubmitMasterExerciseUseCase) submitQA(in SubmitMasterExerciseInput, ex *domain.MasterExercise) (*SubmitMasterExerciseOutput, error) {
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
	if err := uc.submissions.Create(submission); err != nil {
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
