package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// fakeMasterExerciseRepo は MasterExerciseRepository の最小スタブ。
// language 引数や id 引数を記録して assert する。
type fakeMasterExerciseRepo struct {
	listLanguage string
	listResult   []domain.MasterExercise
	getResult    *domain.MasterExercise
	getErr       error
}

func (r *fakeMasterExerciseRepo) ListByLanguage(language string) ([]domain.MasterExercise, error) {
	r.listLanguage = language
	return r.listResult, nil
}

func (r *fakeMasterExerciseRepo) GetByID(_ uint64) (*domain.MasterExercise, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.getResult, nil
}

func newMasterExerciseTestHandler(repo *fakeMasterExerciseRepo) *gin.Engine {
	h := NewMasterExerciseHandler(
		usecase.NewListMasterExercisesUseCase(repo),
		usecase.NewGetMasterExerciseUseCase(repo),
	)
	r := gin.New()
	r.GET("/exercises", h.List)
	r.GET("/exercises/:id", h.Get)
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
	r := newMasterExerciseTestHandler(repo)

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
	r := newMasterExerciseTestHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)
	if repo.listLanguage != "" {
		t.Errorf("listLanguage = %q, want empty", repo.listLanguage)
	}
}

// /exercises/:id → 取得成功時に 200 + JSON。
func TestMasterExerciseHandler_Get_Success(t *testing.T) {
	repo := &fakeMasterExerciseRepo{
		getResult: &domain.MasterExercise{ID: 7, Slug: "php-7", Language: "php", Title: "OOP"},
	}
	r := newMasterExerciseTestHandler(repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/exercises/7", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var got domain.MasterExercise
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != 7 || got.Slug != "php-7" {
		t.Errorf("got %+v, want id=7 slug=php-7", got)
	}
}

// 不正な :id → 400 invalid id。
func TestMasterExerciseHandler_Get_BadID(t *testing.T) {
	repo := &fakeMasterExerciseRepo{}
	r := newMasterExerciseTestHandler(repo)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/exercises/notanumber", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
