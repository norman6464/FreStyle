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

func (f *fakeCourseRepo) ListByWorkspaceID(context.Context, string, bool) ([]domain.Course, error) {
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

func (f *fakeCourseRepo) CreateWithOwnerGrant(_ context.Context, c *domain.Course, _ string) error {
	if f.writeErr == nil {
		c.ID = 100
	}
	return f.writeErr
}
func (f *fakeCourseRepo) Update(context.Context, *domain.Course) error { return f.writeErr }
func (f *fakeCourseRepo) Delete(context.Context, uint64) error         { return f.writeErr }

// fakeMaterialRepo は repository.TeachingMaterialRepository の no-op fake（Delete cascade 用）。
type fakeMaterialRepo struct{}

func (fakeMaterialRepo) ListByCourse(context.Context, uint64, bool) ([]domain.TeachingMaterial, error) {
	return nil, nil
}

func (fakeMaterialRepo) GetByID(context.Context, uint64) (*domain.TeachingMaterial, error) {
	return nil, nil
}

func (fakeMaterialRepo) CountByCourseForWorkspace(context.Context, string, bool) (map[uint64]int, error) {
	return map[uint64]int{}, nil
}
func (fakeMaterialRepo) Create(context.Context, *domain.TeachingMaterial) error { return nil }
func (fakeMaterialRepo) Update(context.Context, *domain.TeachingMaterial) error { return nil }
func (fakeMaterialRepo) UpdateDocWithRevision(context.Context, uint64, string, int) (*domain.TeachingMaterial, error) {
	return nil, nil
}
func (fakeMaterialRepo) Delete(context.Context, uint64) error         { return nil }
func (fakeMaterialRepo) DeleteByCourse(context.Context, uint64) error { return nil }

// fakeChapterViewRepoH は repository.UserChapterViewRepository の最小 fake。
type fakeChapterViewRepoH struct {
	lastViewed *domain.UserChapterView
}

func (f *fakeChapterViewRepoH) UpsertView(context.Context, uint64, uint64, uint64) error { return nil }

func (f *fakeChapterViewRepoH) ListRecentByUser(context.Context, uint64, int) ([]domain.UserChapterView, error) {
	return nil, nil
}

func (f *fakeChapterViewRepoH) GetLastViewedByUserAndCourse(context.Context, uint64, uint64) (*domain.UserChapterView, error) {
	return f.lastViewed, nil
}

func newCourseHandler(cr repository.CourseRepository) *CourseHandler {
	return newCourseHandlerWithViews(cr, &fakeChapterViewRepoH{})
}

func newCourseHandlerWithViews(cr repository.CourseRepository, cv repository.UserChapterViewRepository) *CourseHandler {
	// 既定は「編集できる人」。認可そのものを見るテストは newCourseHandlerWith で
	// 見え方を差し替える。
	return newCourseHandlerWith(cr, cv, materialPermOf(domain.GrantRoleEditor))
}

func newCourseHandlerWith(
	cr repository.CourseRepository, cv repository.UserChapterViewRepository, perm *fakeMaterialPerm,
) *CourseHandler {
	permUC := usecase.NewCheckMaterialPermissionUseCase(perm)
	principals := newKbFakePerms(newKbFakePages(), kbCanEdit)
	principals.addMember(courseWorkspaceID, 1)
	return NewCourseHandler(
		usecase.NewCourseUseCase(cr, fakeMaterialRepo{}, permUC, principals),
		usecase.NewListCoursesWithProgressUseCase(fakeMaterialRepo{}, &fakeProgressRepoH{}, perm),
		usecase.NewGetLastViewedChapterUseCase(cv, permUC),
	)
}

// courseWorkspaceID は actor と対象コースを同じワークスペースに置くためのテスト用 ID。
const courseWorkspaceID = "0198a000-0000-7000-8000-0000000000c1"

// superAdminCo は workspace_id 付きの super_admin（course handler の actorWorkspaceFromContext 用）。
func superAdminCo() *domain.User {
	wid := courseWorkspaceID
	return &domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: &wid}
}

// courseInWorkspace は actor と同じワークスペースに属するコースを組み立てる。
func courseInWorkspace(c domain.Course) *domain.Course {
	wid := courseWorkspaceID
	c.WorkspaceID = &wid
	return &c
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
		// 一覧は権限の事実をまとめて引くので、失敗もそちらから来る。
		w, c := ctxJSON(http.MethodGet, "", nil, superAdminCo())
		perm := materialPermOf(domain.GrantRoleEditor)
		perm.listErr = context.DeadlineExceeded
		newCourseHandlerWith(&fakeCourseRepo{}, &fakeChapterViewRepoH{}, perm).List(c)
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
		newCourseHandler(&fakeCourseRepo{one: courseInWorkspace(domain.Course{ID: 5, Title: "C"})}).Get(c)
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
		newCourseHandler(&fakeCourseRepo{one: courseInWorkspace(domain.Course{ID: 5})}).Update(c)
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
		newCourseHandler(&fakeCourseRepo{one: courseInWorkspace(domain.Course{ID: 5})}).Delete(c)
		if c.Writer.Status() != http.StatusNoContent {
			t.Fatalf("want 204, got %d", c.Writer.Status())
		}
	})
}

func Test_コースハンドラ_最終閲覧章(t *testing.T) {
	course := courseInWorkspace(domain.Course{ID: 5, IsPublished: true})

	t.Run("不正な id → 400", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("abc"), superAdminCo())
		newCourseHandler(&fakeCourseRepo{one: course}).LastViewed(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("未認証 → 401", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), nil)
		newCourseHandler(&fakeCourseRepo{one: course}).LastViewed(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("履歴あり → 200", func(t *testing.T) {
		view := &domain.UserChapterView{UserID: 1, TeachingMaterialID: 42, CourseID: 5}
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), superAdminCo())
		newCourseHandlerWithViews(&fakeCourseRepo{one: course}, &fakeChapterViewRepoH{lastViewed: view}).LastViewed(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("履歴なし → 204", func(t *testing.T) {
		_, c := ctxJSON(http.MethodGet, "", idParam("5"), superAdminCo())
		newCourseHandlerWithViews(&fakeCourseRepo{one: course}, &fakeChapterViewRepoH{}).LastViewed(c)
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
