package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMasterExerciseRepo は SubmitMasterExerciseUseCase テスト用のスタブ。
type fakeMasterExerciseRepo struct {
	get *domain.MasterExercise
	err error
}

func (r *fakeMasterExerciseRepo) ListByLanguage(context.Context, string) ([]domain.MasterExercise, error) {
	return nil, nil
}

func (r *fakeMasterExerciseRepo) GetByID(context.Context, uint64) (*domain.MasterExercise, error) {
	return r.get, r.err
}

func (r *fakeMasterExerciseRepo) GetBySlug(context.Context, string) (*domain.MasterExercise, error) {
	return r.get, r.err
}

func (r *fakeMasterExerciseRepo) ListWithStatusByLanguage(context.Context, uint64, string) ([]repository.MasterExerciseWithStatus, error) {
	return nil, nil
}

type fakeExampleRepo struct {
	rows []domain.MasterExerciseExample
}

func (r *fakeExampleRepo) ListByExerciseID(context.Context, uint64) ([]domain.MasterExerciseExample, error) {
	return r.rows, nil
}

func (r *fakeExampleRepo) ListByExerciseIDs(context.Context, []uint64) (map[uint64][]domain.MasterExerciseExample, error) {
	return nil, nil
}

type fakeSubmissionRepo struct {
	created   *domain.ExerciseSubmission
	createErr error
}

func (r *fakeSubmissionRepo) Create(_ context.Context, s *domain.ExerciseSubmission) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = s
	s.ID = 999
	return nil
}

func (r *fakeSubmissionRepo) ListByUserAndExercise(context.Context, uint64, uint64, string) ([]domain.ExerciseSubmission, error) {
	return nil, nil
}

func (r *fakeSubmissionRepo) HasSolved(context.Context, uint64, uint64, string) (bool, error) {
	return false, nil
}

func (r *fakeSubmissionRepo) HasAttempted(context.Context, uint64, uint64, string) (bool, error) {
	return false, nil
}

// fakeExecutor は ExecuteCodeUseCase の代わり。 入力 stdin で出力を切り替える。
type fakeExecutor struct {
	calls []usecase.ExecuteCodeInput
	// stdinToOut: stdin に対応する stdout を返す。 なければ空。
	stdinToOut map[string]string
	// runErr が non-nil なら Execute がエラーを返す。
	runErr error
	// 失敗ケース: 指定 stdin のとき exitCode != 0 を返す。
	failExit map[string]int
}

func (f *fakeExecutor) Execute(_ context.Context, in usecase.ExecuteCodeInput) (*usecase.ExecuteCodeOutput, error) {
	if f.runErr != nil {
		return nil, f.runErr
	}
	f.calls = append(f.calls, in)
	out := &usecase.ExecuteCodeOutput{Stdout: f.stdinToOut[in.Stdin]}
	if code, ok := f.failExit[in.Stdin]; ok {
		out.ExitCode = code
		out.Stderr = "boom"
	}
	return out, nil
}

func TestSubmitMasterExercise_AllPass(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{ID: 7, Slug: "php-7", Language: "php"},
	}
	examples := &fakeExampleRepo{rows: []domain.MasterExerciseExample{
		{ID: 1, ExerciseID: 7, OrderIndex: 1, InputText: "", ExpectedOutput: "Hello"},
		{ID: 2, ExerciseID: 7, OrderIndex: 2, InputText: "x", ExpectedOutput: "x"},
	}}
	submissions := &fakeSubmissionRepo{}
	executor := &fakeExecutor{stdinToOut: map[string]string{
		"":  "Hello\n",
		"x": "x",
	}}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, examples, submissions, executor)
	out, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "php-7", Code: "<?php echo 'Hello';",
	})
	require.NoError(t, err)
	assert.True(t, out.IsCorrect)
	assert.Len(t, out.Results, 2)
	assert.True(t, out.Results[0].Passed)
	assert.True(t, out.Results[1].Passed)
	require.NotNil(t, submissions.created)
	assert.Equal(t, uint64(1), submissions.created.UserID)
	assert.Equal(t, uint64(7), submissions.created.ExerciseID)
	assert.Equal(t, domain.ExerciseKindMaster, submissions.created.ExerciseKind)
	assert.True(t, submissions.created.IsCorrect)
	// 全 example に対して executor が呼ばれた。
	assert.Len(t, executor.calls, 2)
}

func TestSubmitMasterExercise_OneFails(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{ID: 7, Slug: "php-7", Language: "php"},
	}
	examples := &fakeExampleRepo{rows: []domain.MasterExerciseExample{
		{ID: 1, ExerciseID: 7, OrderIndex: 1, InputText: "1", ExpectedOutput: "1"},
		{ID: 2, ExerciseID: 7, OrderIndex: 2, InputText: "2", ExpectedOutput: "2"},
	}}
	submissions := &fakeSubmissionRepo{}
	// 1 件目を失敗させ、 失敗後も全ケースが実行されることまで検証する
	// （短絡実装で 2 件目が呼ばれない実装に退化したらこのテストで落ちる）。
	executor := &fakeExecutor{stdinToOut: map[string]string{
		"1": "wrong",
		"2": "2",
	}}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, examples, submissions, executor)
	out, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "php-7", Code: "<?php",
	})
	require.NoError(t, err)
	assert.False(t, out.IsCorrect)
	assert.False(t, out.Results[0].Passed)
	assert.True(t, out.Results[1].Passed)
	require.NotNil(t, submissions.created)
	assert.False(t, submissions.created.IsCorrect)
	// 最初の失敗ケースの stdout が representative として記録されている。
	assert.Equal(t, "wrong", submissions.created.Stdout)
	// 失敗後も全 example が実行される（短絡しない）。
	assert.Len(t, executor.calls, 2)
}

func TestSubmitMasterExercise_NormalizeOutput(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{ID: 1, Slug: "s", Language: "php"},
	}
	examples := &fakeExampleRepo{rows: []domain.MasterExerciseExample{
		{ID: 1, ExerciseID: 1, OrderIndex: 1, InputText: "", ExpectedOutput: "42"},
	}}
	submissions := &fakeSubmissionRepo{}
	// 末尾に \r\n を含めても normalize で吸収される。
	executor := &fakeExecutor{stdinToOut: map[string]string{"": "42\r\n"}}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, examples, submissions, executor)
	out, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "s", Code: "<?php",
	})
	require.NoError(t, err)
	assert.True(t, out.IsCorrect)
}

func TestSubmitMasterExercise_NonZeroExit_Fails(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{ID: 1, Slug: "s", Language: "php"},
	}
	examples := &fakeExampleRepo{rows: []domain.MasterExerciseExample{
		{ID: 1, ExerciseID: 1, OrderIndex: 1, InputText: "", ExpectedOutput: "ok"},
	}}
	submissions := &fakeSubmissionRepo{}
	executor := &fakeExecutor{
		stdinToOut: map[string]string{"": "ok"},
		failExit:   map[string]int{"": 1},
	}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, examples, submissions, executor)
	out, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "s", Code: "<?php",
	})
	require.NoError(t, err)
	// stdout は一致するが exit_code != 0 で失敗扱い。
	assert.False(t, out.IsCorrect)
	assert.False(t, out.Results[0].Passed)
}

// examples が登録されていない演習 (seed.py 経由で投入された linux/git/go 等) は
// exercise 自身の ExpectedOutput を 単一テストケースとして fallback する。
func TestSubmitMasterExercise_NoExamples_FallbackToExerciseExpected(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{
			ID: 1, Slug: "linux-1", Language: "bash",
			ExpectedOutput: "Hello, Linux!",
		},
	}
	examples := &fakeExampleRepo{} // empty
	submissions := &fakeSubmissionRepo{}
	executor := &fakeExecutor{stdinToOut: map[string]string{"": "Hello, Linux!\n"}}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, examples, submissions, executor)
	out, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "linux-1", Code: `echo "Hello, Linux!"`,
	})
	require.NoError(t, err)
	assert.True(t, out.IsCorrect)
	assert.Len(t, out.Results, 1)
	// fallback は exercise.Language を passed する (旧実装のように "php" 固定ではない)。
	assert.Len(t, executor.calls, 1)
	assert.Equal(t, "bash", executor.calls[0].Language)
}

// SubmitMasterExerciseUseCase が exercise.Language を ExecuteCodeInput に正しく渡しているか。
// 旧実装は "php" 固定だったため、 Go / bash / Linux 演習が正しく実行できない不具合があった。
func TestSubmitMasterExercise_PassesExerciseLanguage(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{ID: 9, Slug: "go-1", Language: "go"},
	}
	examples := &fakeExampleRepo{rows: []domain.MasterExerciseExample{
		{ID: 1, ExerciseID: 9, OrderIndex: 1, InputText: "", ExpectedOutput: "Hello, Go!"},
	}}
	submissions := &fakeSubmissionRepo{}
	executor := &fakeExecutor{stdinToOut: map[string]string{"": "Hello, Go!\n"}}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, examples, submissions, executor)
	_, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "go-1", Code: "package main",
	})
	require.NoError(t, err)
	require.Len(t, executor.calls, 1)
	assert.Equal(t, "go", executor.calls[0].Language)
}

func TestSubmitMasterExercise_QA_Correct(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{
			ID: 100, Slug: "docker-1", Language: "shell",
			Mode:           domain.ExerciseModeQA,
			ExpectedOutput: "docker run hello-world",
		},
	}
	submissions := &fakeSubmissionRepo{}
	executor := &fakeExecutor{}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, &fakeExampleRepo{}, submissions, executor)
	out, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "docker-1", Code: "docker run hello-world",
	})
	require.NoError(t, err)
	assert.True(t, out.IsCorrect)
	require.Len(t, out.Results, 1)
	assert.True(t, out.Results[0].Passed)
	// QA モードは executor を呼ばない。
	assert.Empty(t, executor.calls)
	require.NotNil(t, submissions.created)
	assert.True(t, submissions.created.IsCorrect)
}

func TestSubmitMasterExercise_QA_Wrong(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{
			ID: 100, Slug: "docker-1", Language: "shell",
			Mode:           domain.ExerciseModeQA,
			ExpectedOutput: "docker run hello-world",
		},
	}
	submissions := &fakeSubmissionRepo{}
	executor := &fakeExecutor{}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, &fakeExampleRepo{}, submissions, executor)
	out, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "docker-1", Code: "docker run nginx",
	})
	require.NoError(t, err)
	assert.False(t, out.IsCorrect)
	assert.False(t, out.Results[0].Passed)
	assert.Empty(t, executor.calls)
}

// QA モードでも 末尾改行 / 空白の差異は normalize で吸収する。
func TestSubmitMasterExercise_QA_NormalizesTrailingWhitespace(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{
			ID: 100, Slug: "docker-1", Language: "shell",
			Mode:           domain.ExerciseModeQA,
			ExpectedOutput: "git init",
		},
	}
	submissions := &fakeSubmissionRepo{}
	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, &fakeExampleRepo{}, submissions, &fakeExecutor{})
	out, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "docker-1", Code: "git init   \n",
	})
	require.NoError(t, err)
	assert.True(t, out.IsCorrect)
}

func TestSubmitMasterExercise_ExerciseNotFound(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{err: errors.New("record not found")}
	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, &fakeExampleRepo{}, &fakeSubmissionRepo{}, &fakeExecutor{})
	_, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "missing", Code: "<?php",
	})
	require.Error(t, err)
}

func TestSubmitMasterExercise_RequiresInput(t *testing.T) {
	uc := usecase.NewSubmitMasterExerciseUseCase(&fakeMasterExerciseRepo{}, &fakeExampleRepo{}, &fakeSubmissionRepo{}, &fakeExecutor{})
	_, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 0, Slug: "s", Code: "x",
	})
	require.Error(t, err)
	_, err = uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "", Code: "x",
	})
	require.Error(t, err)
	_, err = uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "s", Code: "  ",
	})
	require.Error(t, err)
}
