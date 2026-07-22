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
	"gorm.io/gorm"
)

// fakeProgressRepoH / fakeMaterialRepoH / fakeCourseRepoH は handler テスト用の最小 fake。
type fakeProgressRepoH struct{ rows []domain.UserLessonProgress }

func (f *fakeProgressRepoH) MarkCompleted(context.Context, uint64, uint64, uint64) (bool, error) {
	return true, nil
}

func (f *fakeProgressRepoH) MarkIncomplete(context.Context, uint64, uint64) error { return nil }

func (f *fakeProgressRepoH) CountCompletedByUserGroupedByCourse(context.Context, uint64) (map[uint64]int, error) {
	counts := map[uint64]int{}
	for _, r := range f.rows {
		counts[r.CourseID]++
	}
	return counts, nil
}

func (f *fakeProgressRepoH) ListByUser(context.Context, uint64) ([]domain.UserLessonProgress, error) {
	return f.rows, nil
}

type fakeMaterialRepoH struct {
	m      *domain.TeachingMaterial
	getErr error
}

func (f *fakeMaterialRepoH) GetByID(context.Context, uint64) (*domain.TeachingMaterial, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.m, nil
}

func (f *fakeMaterialRepoH) ListByCompany(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (f *fakeMaterialRepoH) ListByCourse(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (f *fakeMaterialRepoH) CountByCourseForCompany(context.Context, uint64, bool) (map[uint64]int, error) {
	return nil, nil
}
func (f *fakeMaterialRepoH) Create(context.Context, *domain.TeachingMaterial) error { return nil }
func (f *fakeMaterialRepoH) Update(context.Context, *domain.TeachingMaterial) error { return nil }
func (f *fakeMaterialRepoH) Delete(context.Context, uint64) error                   { return nil }
func (f *fakeMaterialRepoH) DeleteByCourse(context.Context, uint64) error           { return nil }

type fakeCourseRepoH struct{ c *domain.Course }

func (f *fakeCourseRepoH) GetByID(context.Context, uint64) (*domain.Course, error) {
	if f.c == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.c, nil
}

func (f *fakeCourseRepoH) ListByCompany(context.Context, uint64, bool) ([]domain.Course, error) {
	return nil, nil
}
func (f *fakeCourseRepoH) Create(context.Context, *domain.Course) error { return nil }
func (f *fakeCourseRepoH) Update(context.Context, *domain.Course) error { return nil }
func (f *fakeCourseRepoH) Delete(context.Context, uint64) error         { return nil }

// engineOpts はテストルータ構築のオプション。
type engineOpts struct {
	material  *fakeMaterialRepoH
	course    *domain.Course
	withUser  bool
	companyID uint64
}

// newLessonProgressEngine は実際の gin engine にルートを張ったテスト用ルータを返す。
// 直接 handler を呼ぶと c.Status(204) が body 無しでフラッシュされず recorder が 200 のままになる
// gin の挙動を避けるため、 本番同様 ServeHTTP 経由で検証する。
// withUser=false のときは current user middleware を挟まず未認証を再現する。
func newLessonProgressEngine(o engineOpts) *gin.Engine {
	gin.SetMode(gin.TestMode)
	progress := &fakeProgressRepoH{rows: []domain.UserLessonProgress{{TeachingMaterialID: 1, CourseID: 9}}}
	materials := o.material
	if materials == nil {
		materials = &fakeMaterialRepoH{}
	}
	courses := &fakeCourseRepoH{c: o.course}
	h := NewLessonProgressHandler(
		usecase.NewMarkLessonCompletedUseCase(progress, materials, courses, &nopActivityRepo{}),
		usecase.NewMarkLessonIncompleteUseCase(progress),
		usecase.NewListLessonProgressUseCase(progress),
	)
	r := gin.New()
	if o.withUser {
		companyID := o.companyID
		r.Use(func(c *gin.Context) {
			c.Set(middleware.ContextKeyCurrentUserID, uint64(7))
			c.Set(middleware.ContextKeyCurrentUser, &domain.User{
				ID: 7, CompanyID: &companyID, Role: domain.RoleTrainee,
			})
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

// publishedMaterial は自社・公開教材 + 公開コースの正常系セットを返す。
func publishedMaterial(companyID, courseID uint64) (*fakeMaterialRepoH, *domain.Course) {
	mat := &fakeMaterialRepoH{m: &domain.TeachingMaterial{
		ID: 5, CompanyID: companyID, CourseID: courseID, IsPublished: true,
	}}
	crs := &domain.Course{ID: courseID, CompanyID: companyID, IsPublished: true}
	return mat, crs
}

func Test_進捗ハンドラ_一覧_正常系(t *testing.T) {
	r := newLessonProgressEngine(engineOpts{withUser: true, companyID: 10})
	w := doLessonProgressReq(r, http.MethodGet, "/lesson-progress", "")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_完了_正常系(t *testing.T) {
	mat, crs := publishedMaterial(10, 9)
	r := newLessonProgressEngine(engineOpts{material: mat, course: crs, withUser: true, companyID: 10})
	w := doLessonProgressReq(r, http.MethodPost, "/lesson-progress", `{"teachingMaterialId":5}`)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_完了_他社教材は403(t *testing.T) {
	mat, crs := publishedMaterial(10, 9) // company 10 の教材
	r := newLessonProgressEngine(engineOpts{material: mat, course: crs, withUser: true, companyID: 20})
	w := doLessonProgressReq(r, http.MethodPost, "/lesson-progress", `{"teachingMaterialId":5}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_完了_存在しない教材は404(t *testing.T) {
	mat := &fakeMaterialRepoH{getErr: gorm.ErrRecordNotFound}
	r := newLessonProgressEngine(engineOpts{material: mat, withUser: true, companyID: 10})
	w := doLessonProgressReq(r, http.MethodPost, "/lesson-progress", `{"teachingMaterialId":5}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_完了取消_正常系(t *testing.T) {
	r := newLessonProgressEngine(engineOpts{withUser: true, companyID: 10})
	w := doLessonProgressReq(r, http.MethodDelete, "/lesson-progress/5", "")
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}

func Test_進捗ハンドラ_未認証は401(t *testing.T) {
	r := newLessonProgressEngine(engineOpts{withUser: false}) // current user middleware なし
	w := doLessonProgressReq(r, http.MethodGet, "/lesson-progress", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}
