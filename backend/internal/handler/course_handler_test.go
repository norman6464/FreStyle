package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeCourseRepo は repository.CourseRepository の最小 fake。
type fakeCourseRepo struct {
	rows     []domain.Course
	one      *domain.Course
	getErr   error
	listErr  error
	writeErr error
}

func (f *fakeCourseRepo) ListByCompany(context.Context, uint64, bool) ([]domain.Course, error) {
	return f.rows, f.listErr
}

func (f *fakeCourseRepo) GetByID(context.Context, uint64) (*domain.Course, error) {
	return f.one, f.getErr
}

func (f *fakeCourseRepo) Create(_ context.Context, c *domain.Course) error {
	if f.writeErr == nil {
		c.ID = 100
	}
	return f.writeErr
}
func (f *fakeCourseRepo) Update(context.Context, *domain.Course) error { return f.writeErr }
func (f *fakeCourseRepo) Delete(context.Context, uint64) error         { return f.writeErr }

// fakeMaterialRepo は repository.TeachingMaterialRepository の no-op fake（Delete cascade 用）。
type fakeMaterialRepo struct{}

func (fakeMaterialRepo) ListByCompany(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (fakeMaterialRepo) ListByCourse(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (fakeMaterialRepo) GetByID(context.Context, uint64) (*domain.TeachingMaterial, error) {
	return nil, nil
}

func (fakeMaterialRepo) CountByCourseForCompany(context.Context, uint64, bool) (map[uint64]int, error) {
	return map[uint64]int{}, nil
}
func (fakeMaterialRepo) Create(context.Context, *domain.TeachingMaterial) error { return nil }
func (fakeMaterialRepo) Update(context.Context, *domain.TeachingMaterial) error { return nil }
func (fakeMaterialRepo) Delete(context.Context, uint64) error                   { return nil }
func (fakeMaterialRepo) DeleteByCourse(context.Context, uint64) error           { return nil }

func newCourseHandler(cr repository.CourseRepository) *CourseHandler {
	return NewCourseHandler(
		usecase.NewCourseUseCase(cr, fakeMaterialRepo{}),
		usecase.NewListCoursesWithMaterialCountUseCase(cr, fakeMaterialRepo{}),
	)
}

// superAdminCo は company_id 付きの super_admin（course handler の actorContext 用）。
func superAdminCo() *domain.User {
	cid := uint64(1)
	return &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: &cid}
}

func Test_コースハンドラ_一覧(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, nil)
		newCourseHandler(&fakeCourseRepo{}).List(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, superAdminCo())
		newCourseHandler(&fakeCourseRepo{rows: []domain.Course{{ID: 1, Title: "C"}}}).List(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("リポジトリエラー → 500", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, superAdminCo())
		newCourseHandler(&fakeCourseRepo{listErr: context.DeadlineExceeded}).List(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("want 500, got %d", w.Code)
		}
	})
}

func Test_コースハンドラ_取得(t *testing.T) {
	t.Run("不正な id → 400", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("abc"), superAdminCo())
		newCourseHandler(&fakeCourseRepo{}).Get(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), superAdminCo())
		newCourseHandler(&fakeCourseRepo{one: &domain.Course{ID: 5, CompanyID: 1, Title: "C"}}).Get(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
}

func Test_コースハンドラ_作成(t *testing.T) {
	t.Run("不正な JSON → 400", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPost, `{`, nil, superAdminCo())
		newCourseHandler(&fakeCourseRepo{}).Create(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系 → 201", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPost, `{"title":"New"}`, nil, superAdminCo())
		newCourseHandler(&fakeCourseRepo{}).Create(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d", w.Code)
		}
	})
}

func Test_コースハンドラ_更新(t *testing.T) {
	t.Run("不正な id → 400", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPut, `{"title":"X"}`, idParam("abc"), superAdminCo())
		newCourseHandler(&fakeCourseRepo{}).Update(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系 → 200", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPut, `{"title":"X"}`, idParam("5"), superAdminCo())
		newCourseHandler(&fakeCourseRepo{one: &domain.Course{ID: 5, CompanyID: 1}}).Update(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
}

func Test_コースハンドラ_削除(t *testing.T) {
	t.Run("不正な id → 400", func(t *testing.T) {
		w, c := ctxJSON(http.MethodDelete, "", idParam("abc"), superAdminCo())
		newCourseHandler(&fakeCourseRepo{}).Delete(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系 → 204", func(t *testing.T) {
		_, c := ctxJSON(http.MethodDelete, "", idParam("5"), superAdminCo())
		newCourseHandler(&fakeCourseRepo{one: &domain.Course{ID: 5, CompanyID: 1}}).Delete(c)
		if c.Writer.Status() != http.StatusNoContent {
			t.Fatalf("want 204, got %d", c.Writer.Status())
		}
	})
}

func Test_コースハンドラ_作成_カテゴリ(t *testing.T) {
	t.Run("不正なカテゴリ → 400", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPost, `{"title":"X","category":"not-a-category"}`, nil, superAdminCo())
		newCourseHandler(&fakeCourseRepo{}).Create(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("定義済みカテゴリ → 201", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPost, `{"title":"PostgreSQL","category":"database"}`, nil, superAdminCo())
		newCourseHandler(&fakeCourseRepo{}).Create(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d", w.Code)
		}
	})
}
