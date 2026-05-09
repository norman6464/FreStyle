package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMasterExerciseRepo は SubmitMasterExerciseUseCase テスト用のスタブ。
type fakeMasterExerciseRepo struct {
	get *domain.MasterExercise
	err error
}

func (r *fakeMasterExerciseRepo) ListByLanguage(string) ([]domain.MasterExercise, error) {
	return nil, nil
}
func (r *fakeMasterExerciseRepo) GetByID(uint64) (*domain.MasterExercise, error) { return r.get, r.err }
func (r *fakeMasterExerciseRepo) GetBySlug(string) (*domain.MasterExercise, error) {
	return r.get, r.err
}

type fakeExampleRepo struct {
	rows []domain.MasterExerciseExample
}

func (r *fakeExampleRepo) ListByExerciseID(uint64) ([]domain.MasterExerciseExample, error) {
	return r.rows, nil
}
func (r *fakeExampleRepo) ListByExerciseIDs([]uint64) (map[uint64][]domain.MasterExerciseExample, error) {
	return nil, nil
}

type fakeSubmissionRepo struct {
	created   *domain.ExerciseSubmission
	createErr error
}

func (r *fakeSubmissionRepo) Create(s *domain.ExerciseSubmission) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = s
	s.ID = 999
	return nil
}
func (r *fakeSubmissionRepo) ListByUserAndExercise(uint64, uint64, string) ([]domain.ExerciseSubmission, error) {
	return nil, nil
}
func (r *fakeSubmissionRepo) HasSolved(uint64, uint64, string) (bool, error)    { return false, nil }
func (r *fakeSubmissionRepo) HasAttempted(uint64, uint64, string) (bool, error) { return false, nil }
func (r *fakeSubmissionRepo) BatchUserStatuses(uint64, []uint64, string) (map[uint64]string, error) {
	return nil, nil
}
func (r *fakeSubmissionRepo) ExerciseStats(uint64, string) (repository.ExerciseSubmissionStats, error) {
	return repository.ExerciseSubmissionStats{}, nil
}
func (r *fakeSubmissionRepo) ExerciseStatsBatch([]uint64, string) (map[uint64]repository.ExerciseSubmissionStats, error) {
	return nil, nil
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

func TestSubmitMasterExercise_NoExamples_Errors(t *testing.T) {
	exRepo := &fakeMasterExerciseRepo{
		get: &domain.MasterExercise{ID: 1, Slug: "s"},
	}
	examples := &fakeExampleRepo{}
	submissions := &fakeSubmissionRepo{}
	executor := &fakeExecutor{}

	uc := usecase.NewSubmitMasterExerciseUseCase(exRepo, examples, submissions, executor)
	_, err := uc.Execute(context.Background(), usecase.SubmitMasterExerciseInput{
		UserID: 1, Slug: "s", Code: "<?php",
	})
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no test cases"))
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
