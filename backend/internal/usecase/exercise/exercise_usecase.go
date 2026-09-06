package exercise

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
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

// WarmupCodeUseCase は指定言語の実行環境を事前に温める。学習者がコードエディタ
// （演習詳細）ページに入った時点で呼び、最初の Run を即時化する用途。
type WarmupCodeUseCase struct {
	runner CodeRunner
}

// NewWarmupCodeUseCase は CodeRunner を注入して WarmupCodeUseCase を返す。
func NewWarmupCodeUseCase(runner CodeRunner) *WarmupCodeUseCase {
	return &WarmupCodeUseCase{runner: runner}
}

// Execute は language の実行環境を温める（Go はコンパイルキャッシュ、php/bash は no-op）。
func (uc *WarmupCodeUseCase) Execute(ctx context.Context, language string) error {
	return uc.runner.Warmup(ctx, language)
}

type GetExerciseLanguageSummaryUseCase struct {
	exercises repository.MasterExerciseRepository
}

// NewGetExerciseLanguageSummaryUseCase は GetExerciseLanguageSummaryUseCase を生成する。
func NewGetExerciseLanguageSummaryUseCase(exercises repository.MasterExerciseRepository) *GetExerciseLanguageSummaryUseCase {
	return &GetExerciseLanguageSummaryUseCase{exercises: exercises}
}

// Execute は言語別の集計を返す。userID=0（未ログイン）は solved が 0 になる。
func (u *GetExerciseLanguageSummaryUseCase) Execute(ctx context.Context, userID uint64) ([]repository.ExerciseLanguageSummary, error) {
	return u.exercises.SummaryByLanguage(ctx, userID)
}

// GetMasterExerciseDetailOutput は詳細ページに渡す問題本体 + 入出力例セット。
type GetMasterExerciseDetailOutput struct {
	Exercise *domain.MasterExercise         `json:"exercise"`
	Examples []domain.MasterExerciseExample `json:"examples"`
}

// GetMasterExerciseUseCase は slug 指定で運営マスタ演習 + 入出力例を返す（ID 指定は Execute で互換維持）。
type GetMasterExerciseUseCase struct {
	repo     repository.MasterExerciseRepository
	examples repository.MasterExerciseExampleRepository
}

func NewGetMasterExerciseUseCase(
	repo repository.MasterExerciseRepository,
	examples repository.MasterExerciseExampleRepository,
) *GetMasterExerciseUseCase {
	return &GetMasterExerciseUseCase{repo: repo, examples: examples}
}

// Execute は ID 指定で取得する旧 API 互換。 examples は付かない。
func (uc *GetMasterExerciseUseCase) Execute(ctx context.Context, id uint64) (*domain.MasterExercise, error) {
	return uc.repo.GetByID(ctx, id)
}

// ExecuteBySlug は slug ベースの詳細ページ向けに examples を含めて返す。
// NotFound は handler で 404 に分岐できるようそのまま伝搬し、ex == nil は defensive に弾く。
func (uc *GetMasterExerciseUseCase) ExecuteBySlug(ctx context.Context, slug string) (*GetMasterExerciseDetailOutput, error) {
	ex, err := uc.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, fmt.Errorf("exercise not found: %s", slug)
	}
	examples, err := uc.examples.ListByExerciseID(ctx, ex.ID)
	if err != nil {
		return nil, err
	}
	return &GetMasterExerciseDetailOutput{Exercise: ex, Examples: examples}, nil
}

// ListMasterExercisesUseCase は指定言語の運営マスタ演習問題一覧を返す。
type ListMasterExercisesUseCase struct {
	repo repository.MasterExerciseRepository
}

func NewListMasterExercisesUseCase(repo repository.MasterExerciseRepository) *ListMasterExercisesUseCase {
	return &ListMasterExercisesUseCase{repo: repo}
}

// Execute は language 指定があれば該当言語のみ、空文字なら全言語の問題を返す。
func (uc *ListMasterExercisesUseCase) Execute(ctx context.Context, language string) ([]domain.MasterExercise, error) {
	return uc.repo.ListByLanguage(ctx, language)
}

// MasterExerciseWithStatus は read model（repository 定義）を handler / OpenAPI 向けに再エクスポートした別名。
// 正準型は repository パッケージにあり、persistence はそちらを返す（層境界のため）。
type MasterExerciseWithStatus = repository.MasterExerciseWithStatus

// ListMasterExercisesWithStatusUseCase は問題一覧 + 各問題の current user 状態 + 集計を返す。
// 取得は repository が 1 クエリ（master_exercises ⟕ exercise_submissions 集計）で行い、N+1 / 多段往復を避ける。
type ListMasterExercisesWithStatusUseCase struct {
	exercises repository.MasterExerciseRepository
}

func NewListMasterExercisesWithStatusUseCase(
	exercises repository.MasterExerciseRepository,
) *ListMasterExercisesWithStatusUseCase {
	return &ListMasterExercisesWithStatusUseCase{exercises: exercises}
}

// ListMasterExercisesWithStatusInput は入力。 UserID=0 は未ログイン扱いで status は全部 ""。
// Offset/Limit はスクロール型ページネーション用。Limit=0 は全件取得。
type ListMasterExercisesWithStatusInput struct {
	UserID   uint64
	Language string
	Offset   int
	Limit    int
}

func (uc *ListMasterExercisesWithStatusUseCase) Execute(ctx context.Context, in ListMasterExercisesWithStatusInput) ([]repository.MasterExerciseWithStatus, error) {
	return uc.exercises.ListWithStatusByLanguage(ctx, repository.ListWithStatusInput{
		UserID:   in.UserID,
		Language: in.Language,
		Offset:   in.Offset,
		Limit:    in.Limit,
	})
}

// ListUserMasterSubmissionsInput は履歴一覧 API への入力。
type ListUserMasterSubmissionsInput struct {
	UserID uint64
	Slug   string
}

// ListUserMasterSubmissionsUseCase は current user の指定問題に対する提出履歴を新しい順で返す。
type ListUserMasterSubmissionsUseCase struct {
	exercises   repository.MasterExerciseRepository
	submissions repository.ExerciseSubmissionRepository
}

func NewListUserMasterSubmissionsUseCase(
	exercises repository.MasterExerciseRepository,
	submissions repository.ExerciseSubmissionRepository,
) *ListUserMasterSubmissionsUseCase {
	return &ListUserMasterSubmissionsUseCase{exercises: exercises, submissions: submissions}
}

func (uc *ListUserMasterSubmissionsUseCase) Execute(ctx context.Context, in ListUserMasterSubmissionsInput) ([]domain.ExerciseSubmission, error) {
	ex, err := uc.exercises.GetBySlug(ctx, in.Slug)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, fmt.Errorf("exercise not found: %s", in.Slug)
	}
	return uc.submissions.ListByUserAndExercise(ctx, in.UserID, ex.ID, domain.ExerciseKindMaster)
}

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
	if ex.Mode == domain.ExerciseModePreview {
		return uc.submitPreview(ctx, in, ex)
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
// submitPreview は HTML/CSS 等のライブプレビュー演習の提出を記録する。
// 出来栄えは学習者が見本とプレビューを見比べて判断する(視覚的セルフチェック)ため、
// サーバー側では実行も出力比較もせず、「できた!」の宣言を完了として提出コードごと保存する。
func (uc *SubmitMasterExerciseUseCase) submitPreview(ctx context.Context, in SubmitMasterExerciseInput, ex *domain.MasterExercise) (*SubmitMasterExerciseOutput, error) {
	submission := &domain.ExerciseSubmission{
		UserID:        in.UserID,
		ExerciseKind:  domain.ExerciseKindMaster,
		ExerciseID:    ex.ID,
		SubmittedCode: in.Code,
		Stdout:        "",
		Stderr:        "",
		ExitCode:      0,
		IsCorrect:     true,
		SubmittedAt:   time.Now().UTC(),
	}
	if err := uc.submissions.Create(ctx, submission); err != nil {
		return nil, err
	}

	return &SubmitMasterExerciseOutput{
		SubmissionID: submission.ID,
		IsCorrect:    true,
		Results:      []TestCaseResult{},
	}, nil
}

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

// normalizeOutput は CRLF/CR を LF に統一し、各行末と出力全体末尾の空白・改行を除去する。
// 学習者に見えない行末の空白（例: `fmt.Printf("%d \n")` の余分なスペース）で不合格にしない
// ため、行内（先頭側）の空白は保持しつつ行末の空白/タブだけを落とす。
func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), " \t\n")
}
