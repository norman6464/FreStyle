package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// fakeFullSubmissionRepo は提出ハンドラの統合テスト用フェイク。
// ListByUserAndExercise は呼び出し時の引数を last* に記録し、
// handler が user_id / exercise_id / kind を正しく渡しているかをテスト側で検証できるようにする。
type fakeFullSubmissionRepo struct {
	created        *domain.ExerciseSubmission
	listed         []domain.ExerciseSubmission
	lastUserID     uint64
	lastExerciseID uint64
	lastKind       string
}

func (r *fakeFullSubmissionRepo) Create(_ context.Context, s *domain.ExerciseSubmission) error {
	s.ID = 42
	r.created = s
	return nil
}
func (r *fakeFullSubmissionRepo) ListByUserAndExercise(_ context.Context, userID, exerciseID uint64, kind string) ([]domain.ExerciseSubmission, error) {
	r.lastUserID = userID
	r.lastExerciseID = exerciseID
	r.lastKind = kind
	return r.listed, nil
}
func (r *fakeFullSubmissionRepo) HasSolved(context.Context, uint64, uint64, string) (bool, error) {
	return false, nil
}
func (r *fakeFullSubmissionRepo) HasAttempted(context.Context, uint64, uint64, string) (bool, error) {
	return false, nil
}

// stubExecutorForHandlerTest は ExecuteCodeUseCase の代替。
type stubExecutorForHandlerTest struct{ stdout string }

func (s *stubExecutorForHandlerTest) Execute(_ context.Context, in usecase.ExecuteCodeInput) (*usecase.ExecuteCodeOutput, error) {
	return &usecase.ExecuteCodeOutput{Stdout: s.stdout}, nil
}

func newSubmissionTestRouter(t *testing.T, exercise *domain.MasterExercise, examples []domain.MasterExerciseExample, executorOut string, listed []domain.ExerciseSubmission) (*gin.Engine, *fakeFullSubmissionRepo) {
	t.Helper()
	exRepo := &fakeMasterExerciseRepo{getResult: exercise}
	exampleRepo := &fakeExampleRepo{byID: map[uint64][]domain.MasterExerciseExample{exercise.ID: examples}}
	subRepo := &fakeFullSubmissionRepo{listed: listed}
	executor := &stubExecutorForHandlerTest{stdout: executorOut}

	h := NewExerciseSubmissionHandler(
		usecase.NewSubmitMasterExerciseUseCase(exRepo, exampleRepo, subRepo, executor),
		usecase.NewListUserMasterSubmissionsUseCase(exRepo, subRepo),
	)
	r := gin.New()
	// テスト用に固定 user_id を context にセットしてから handler を呼ぶ。
	r.Use(func(c *gin.Context) {
		c.Set("currentUserID", uint64(101))
		c.Next()
	})
	r.POST("/exercises/:slug/submit", h.Submit)
	r.GET("/exercises/:slug/submissions", h.List)
	return r, subRepo
}

func TestSubmissionHandler_Submit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	exercise := &domain.MasterExercise{ID: 7, Slug: "php-7"}
	examples := []domain.MasterExerciseExample{
		{ID: 1, ExerciseID: 7, OrderIndex: 1, InputText: "", ExpectedOutput: "Hello"},
	}
	r, subRepo := newSubmissionTestRouter(t, exercise, examples, "Hello", nil)

	body, _ := json.Marshal(map[string]string{"code": "<?php echo 'Hello';"})
	req := httptest.NewRequest(http.MethodPost, "/exercises/php-7/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if subRepo.created == nil {
		t.Fatal("submission should have been persisted")
	}
	if subRepo.created.UserID != 101 {
		t.Errorf("UserID = %d, want 101", subRepo.created.UserID)
	}
	if !subRepo.created.IsCorrect {
		t.Error("IsCorrect should be true")
	}
}

func TestSubmissionHandler_Submit_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	exercise := &domain.MasterExercise{ID: 7, Slug: "php-7"}
	r, _ := newSubmissionTestRouter(t, exercise, []domain.MasterExerciseExample{
		{ExerciseID: 7, OrderIndex: 1, ExpectedOutput: ""},
	}, "", nil)
	// middleware で user_id をセットせずに呼ぶシナリオを再現するため新しい engine を作る。
	r2 := gin.New()
	for _, route := range r.Routes() {
		// route.HandlerFunc が登録済の handler chain なのでこのままだと再利用できない。
		// 代わりに同じハンドラを user_id 無しの Engine に再登録する。
		_ = route
	}
	// 簡素化のため: 既存 r で `currentUserID` セット前に walk できないので
	// 認可テストは別ルータを直接組み立てる。
	exRepo := &fakeMasterExerciseRepo{getResult: exercise}
	exampleRepo := &fakeExampleRepo{}
	subRepo := &fakeFullSubmissionRepo{}
	executor := &stubExecutorForHandlerTest{}
	h := NewExerciseSubmissionHandler(
		usecase.NewSubmitMasterExerciseUseCase(exRepo, exampleRepo, subRepo, executor),
		usecase.NewListUserMasterSubmissionsUseCase(exRepo, subRepo),
	)
	r2.POST("/exercises/:slug/submit", h.Submit)

	body, _ := json.Marshal(map[string]string{"code": "<?php"})
	req := httptest.NewRequest(http.MethodPost, "/exercises/php-7/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r2.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestSubmissionHandler_List_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	exercise := &domain.MasterExercise{ID: 7, Slug: "php-7"}
	listed := []domain.ExerciseSubmission{
		{ID: 1, UserID: 101, ExerciseID: 7, ExerciseKind: domain.ExerciseKindMaster, IsCorrect: true},
		{ID: 2, UserID: 101, ExerciseID: 7, ExerciseKind: domain.ExerciseKindMaster, IsCorrect: false},
	}
	r, subRepo := newSubmissionTestRouter(t, exercise, nil, "", listed)

	req := httptest.NewRequest(http.MethodGet, "/exercises/php-7/submissions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var got []domain.ExerciseSubmission
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
	// handler が repository に user_id / exercise_id / kind を正しく渡したかを検証する。
	if subRepo.lastUserID != 101 {
		t.Errorf("ListByUserAndExercise userID = %d, want 101", subRepo.lastUserID)
	}
	if subRepo.lastExerciseID != 7 {
		t.Errorf("ListByUserAndExercise exerciseID = %d, want 7", subRepo.lastExerciseID)
	}
	if subRepo.lastKind != domain.ExerciseKindMaster {
		t.Errorf("ListByUserAndExercise kind = %s, want %s", subRepo.lastKind, domain.ExerciseKindMaster)
	}
}
