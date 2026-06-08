package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// fakeMasterExerciseRepo は MasterExerciseRepository の最小スタブ。
// language / id / slug 引数を記録して assert する。
type fakeMasterExerciseRepo struct {
	listLanguage string
	listResult   []domain.MasterExercise
	getResult    *domain.MasterExercise
	getErr       error
	gotSlug      string
}

func (r *fakeMasterExerciseRepo) ListByLanguage(_ context.Context, language string) ([]domain.MasterExercise, error) {
	r.listLanguage = language
	return r.listResult, nil
}

func (r *fakeMasterExerciseRepo) ListWithStatusByLanguage(_ context.Context, _ uint64, language string) ([]repository.MasterExerciseWithStatus, error) {
	r.listLanguage = language
	out := make([]repository.MasterExerciseWithStatus, 0, len(r.listResult))
	for _, e := range r.listResult {
		out = append(out, repository.MasterExerciseWithStatus{MasterExercise: e})
	}
	return out, nil
}

func (r *fakeMasterExerciseRepo) GetByID(_ context.Context, _ uint64) (*domain.MasterExercise, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.getResult, nil
}

func (r *fakeMasterExerciseRepo) GetBySlug(_ context.Context, slug string) (*domain.MasterExercise, error) {
	r.gotSlug = slug
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.getResult, nil
}

// fakeExampleRepo は MasterExerciseExampleRepository の最小スタブ。
type fakeExampleRepo struct {
	byID map[uint64][]domain.MasterExerciseExample
}

func (r *fakeExampleRepo) ListByExerciseID(_ context.Context, exerciseID uint64) ([]domain.MasterExerciseExample, error) {
	return r.byID[exerciseID], nil
}

func (r *fakeExampleRepo) ListByExerciseIDs(context.Context, []uint64) (map[uint64][]domain.MasterExerciseExample, error) {
	return r.byID, nil
}

func newMasterExerciseTestHandler(repo *fakeMasterExerciseRepo, examples *fakeExampleRepo) *gin.Engine {
	if examples == nil {
		examples = &fakeExampleRepo{}
	}
	h := NewMasterExerciseHandler(
		usecase.NewListMasterExercisesUseCase(repo),
		usecase.NewListMasterExercisesWithStatusUseCase(repo),
		usecase.NewGetMasterExerciseUseCase(repo, examples),
	)
	r := gin.New()
	r.GET("/exercises", h.List)
	r.GET("/exercises/:slug", h.GetBySlug)
	return r
}

// /exercises?language=php → repo に "php" が伝わり、結果が JSON で返る。
func TestMasterExerciseHandler_List_FiltersByLanguage(t *testing.T) {
	repo := &fakeMasterExerciseRepo{
		listResult: []domain.MasterExercise{
			{ID: 1, Slug: "php-1", Language: "php", Title: "Hello", IsPublished: true},
			{ID: 2, Slug: "php-2", Language: "php", Title: "Var", IsPublished: true},
		},
	}
	r := newMasterExerciseTestHandler(repo, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/exercises?language=php", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.listLanguage != "php" {
		t.Errorf("listLanguage = %q, want %q", repo.listLanguage, "php")
	}
	var got []domain.MasterExercise
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

// language 未指定なら全言語（repo 側に空文字が伝わる）。
func TestMasterExerciseHandler_List_AllLanguages(t *testing.T) {
	repo := &fakeMasterExerciseRepo{}
	r := newMasterExerciseTestHandler(repo, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/exercises", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.listLanguage != "" {
		t.Errorf("listLanguage = %q, want empty", repo.listLanguage)
	}
}

// /exercises/:slug → exercise + examples を含むレスポンスを返す。
func TestMasterExerciseHandler_GetBySlug_Success(t *testing.T) {
	repo := &fakeMasterExerciseRepo{
		getResult: &domain.MasterExercise{ID: 7, Slug: "php-7", Language: "php", Title: "OOP"},
	}
	examples := &fakeExampleRepo{
		byID: map[uint64][]domain.MasterExerciseExample{
			7: {
				{ID: 1, ExerciseID: 7, OrderIndex: 1, InputText: "", ExpectedOutput: "Hello"},
				{ID: 2, ExerciseID: 7, OrderIndex: 2, InputText: "x", ExpectedOutput: "x"},
			},
		},
	}
	r := newMasterExerciseTestHandler(repo, examples)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/exercises/php-7", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.gotSlug != "php-7" {
		t.Errorf("gotSlug = %q, want php-7", repo.gotSlug)
	}
	var got struct {
		Exercise *domain.MasterExercise         `json:"exercise"`
		Examples []domain.MasterExerciseExample `json:"examples"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Exercise == nil || got.Exercise.Slug != "php-7" {
		t.Errorf("exercise = %+v, want slug=php-7", got.Exercise)
	}
	if len(got.Examples) != 2 {
		t.Errorf("examples len = %d, want 2", len(got.Examples))
	}
}

// 存在しない slug → 404 (`gorm.ErrRecordNotFound` のケース)。
func TestMasterExerciseHandler_GetBySlug_NotFound(t *testing.T) {
	repo := &fakeMasterExerciseRepo{getErr: gorm.ErrRecordNotFound}
	r := newMasterExerciseTestHandler(repo, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/exercises/missing", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// `gorm.ErrRecordNotFound` 以外の DB エラーは 500 として返す（404 と区別する）。
func TestMasterExerciseHandler_GetBySlug_InternalError(t *testing.T) {
	repo := &fakeMasterExerciseRepo{getErr: errors.New("connection refused")}
	r := newMasterExerciseTestHandler(repo, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/exercises/whatever", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
