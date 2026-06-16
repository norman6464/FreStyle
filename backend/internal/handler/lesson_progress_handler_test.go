package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// fakeProgressRepoH / fakeMaterialRepoH は handler テスト用の最小 fake。
type fakeProgressRepoH struct{ rows []domain.UserLessonProgress }

func (f *fakeProgressRepoH) MarkCompleted(context.Context, uint64, uint64, uint64) error { return nil }

func (f *fakeProgressRepoH) MarkIncomplete(context.Context, uint64, uint64) error { return nil }

func (f *fakeProgressRepoH) ListByUser(context.Context, uint64) ([]domain.UserLessonProgress, error) {
	return f.rows, nil
}

type fakeMaterialRepoH struct{ m *domain.TeachingMaterial }

func (f *fakeMaterialRepoH) GetByID(context.Context, uint64) (*domain.TeachingMaterial, error) {
	return f.m, nil
}

func (f *fakeMaterialRepoH) ListByCompany(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (f *fakeMaterialRepoH) ListByCourse(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}
func (f *fakeMaterialRepoH) Create(context.Context, *domain.TeachingMaterial) error { return nil }
func (f *fakeMaterialRepoH) Update(context.Context, *domain.TeachingMaterial) error { return nil }
func (f *fakeMaterialRepoH) Delete(context.Context, uint64) error                   { return nil }
func (f *fakeMaterialRepoH) DeleteByCourse(context.Context, uint64) error           { return nil }

// newLessonProgressEngine は実際の gin engine にルートを張ったテスト用ルータを返す。
// 直接 handler を呼ぶと c.Status(204) が body 無しでフラッシュされず recorder が 200 のままになる
// gin の挙動を避けるため、本番同様 ServeHTTP 経由で検証する。
// withUser=false のときは current user middleware を挟まず未認証を再現する。
func newLessonProgressEngine(material *domain.TeachingMaterial, withUser bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	progress := &fakeProgressRepoH{rows: []domain.UserLessonProgress{{TeachingMaterialID: 1, CourseID: 9}}}
	materials := &fakeMaterialRepoH{m: material}
	h := NewLessonProgressHandler(
		usecase.NewMarkLessonCompletedUseCase(progress, materials),
		usecase.NewMarkLessonIncompleteUseCase(progress),
		usecase.NewListLessonProgressUseCase(progress),
	)
	r := gin.New()
	if withUser {
		r.Use(func(c *gin.Context) {
			c.Set(middleware.ContextKeyCurrentUserID, uint64(7))
			c.Next()
		})
	}
	r.GET("/lesson-progress", h.List)
	r.POST("/lesson-progress", h.Complete)
	r.DELETE("/lesson-progress/:teachingMaterialId", h.Incomplete)
	return r
}

func doLessonProgressReq(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	r.ServeHTTP(w, req)
	return w
}

func Test_進捗ハンドラ_一覧_正常系(t *testing.T) {
	r := newLessonProgressEngine(nil, true)
	w := doLessonProgressReq(r, http.MethodGet, "/lesson-progress", "")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_完了_正常系(t *testing.T) {
	r := newLessonProgressEngine(&domain.TeachingMaterial{ID: 5, CourseID: 9}, true)
	w := doLessonProgressReq(r, http.MethodPost, "/lesson-progress", `{"teachingMaterialId":5}`)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_完了_存在しない教材は404(t *testing.T) {
	r := newLessonProgressEngine(nil, true) // material=nil → not found
	w := doLessonProgressReq(r, http.MethodPost, "/lesson-progress", `{"teachingMaterialId":5}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_完了取消_正常系(t *testing.T) {
	r := newLessonProgressEngine(nil, true)
	w := doLessonProgressReq(r, http.MethodDelete, "/lesson-progress/5", "")
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_未認証は401(t *testing.T) {
	r := newLessonProgressEngine(nil, false) // current user middleware なし
	w := doLessonProgressReq(r, http.MethodGet, "/lesson-progress", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}
