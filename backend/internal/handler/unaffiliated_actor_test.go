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

// 未所属 actor（users.workspace_id = NULL。運営管理者がこれに当たる）が各経路を通ったときの
// 挙動をここに固定する。「未所属」の表現を変えても素通り / 許可 / 空一覧 / 403 の別が
// ずれないことを保証するための回帰テスト（経路ごとに結論が違うのは現状の仕様）。

// otherWorkspaceID は未所属 actor から見た「自分のものではないワークスペース」。
const otherWorkspaceID = "0198a000-0000-7000-8000-0000000000f1"

// unaffiliatedSuperAdmin はどのワークスペースにも属さない super_admin。
func unaffiliatedSuperAdmin() *domain.User {
	return &domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: nil}
}

// courseInOtherWorkspace / materialInOtherWorkspace は otherWorkspaceID に属する対象を組み立てる。
func courseInOtherWorkspace(c domain.Course) *domain.Course {
	wid := otherWorkspaceID
	c.WorkspaceID = &wid
	return &c
}

func materialInOtherWorkspace(m domain.TeachingMaterial) *domain.TeachingMaterial {
	wid := otherWorkspaceID
	m.WorkspaceID = &wid
	return &m
}

// recordingChapterViewRepo は UpsertView が呼ばれたかを記録する fake。
type recordingChapterViewRepo struct {
	upserted bool
}

func (f *recordingChapterViewRepo) UpsertView(context.Context, uint64, uint64, uint64) error {
	f.upserted = true
	return nil
}

func (f *recordingChapterViewRepo) ListRecentByUser(context.Context, uint64, int) ([]domain.UserChapterView, error) {
	return nil, nil
}

func (f *recordingChapterViewRepo) GetLastViewedByUserAndCourse(context.Context, uint64, uint64) (*domain.UserChapterView, error) {
	return nil, nil
}

func Test_未所属actor_コース経路(t *testing.T) {
	t.Run("一覧は空配列（所属ワークスペースが無いので 1 件も返さない）", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, unaffiliatedSuperAdmin())
		newCourseHandler(&fakeCourseRepo{rows: []domain.Course{*courseInOtherWorkspace(domain.Course{ID: 1, Title: "C"})}}).List(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if got := strings.TrimSpace(w.Body.String()); got != "[]" {
			t.Fatalf("want empty array, got %s", got)
		}
	})

	t.Run("詳細は super_admin として他ワークスペースのコースも読める", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), unaffiliatedSuperAdmin())
		newCourseHandler(&fakeCourseRepo{one: courseInOtherWorkspace(domain.Course{ID: 5, Title: "C"})}).Get(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})

	t.Run("作成は 403（所属ワークスペースが無いので作れない）", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPost, `{"title":"New"}`, nil, unaffiliatedSuperAdmin())
		newCourseHandler(&fakeCourseRepo{}).Create(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", w.Code)
		}
	})

	t.Run("更新は super_admin として通る", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPut, `{"title":"X"}`, idParam("5"), unaffiliatedSuperAdmin())
		newCourseHandler(&fakeCourseRepo{one: courseInOtherWorkspace(domain.Course{ID: 5})}).Update(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})

	t.Run("削除は super_admin として通る", func(t *testing.T) {
		_, c := ctxJSON(http.MethodDelete, "", idParam("5"), unaffiliatedSuperAdmin())
		newCourseHandler(&fakeCourseRepo{one: courseInOtherWorkspace(domain.Course{ID: 5})}).Delete(c)
		if c.Writer.Status() != http.StatusNoContent {
			t.Fatalf("want 204, got %d", c.Writer.Status())
		}
	})

	t.Run("最終閲覧章は super_admin として他ワークスペースのコースでも読める", func(t *testing.T) {
		course := courseInOtherWorkspace(domain.Course{ID: 5, IsPublished: true})
		view := &domain.UserChapterView{UserID: 1, TeachingMaterialID: 42, CourseID: 5}
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), unaffiliatedSuperAdmin())
		newCourseHandlerWithViews(
			&fakeCourseRepo{one: course},
			&fakeChapterViewRepoH{lastViewed: view},
		).LastViewed(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
}

func Test_未所属actor_教材経路(t *testing.T) {
	publishedCourse := courseInOtherWorkspace(domain.Course{ID: 5, IsPublished: true})

	t.Run("全件一覧は空配列", func(t *testing.T) {
		h := NewTeachingMaterialHandler(
			usecase.NewTeachingMaterialUseCase(fakeMaterialRepo{}, &fakeCourseRepo{one: publishedCourse}),
		)
		w, c := ctxJSON(http.MethodGet, "", nil, unaffiliatedSuperAdmin())
		h.List(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if got := strings.TrimSpace(w.Body.String()); got != "[]" {
			t.Fatalf("want empty array, got %s", got)
		}
	})

	t.Run("コース別一覧は super_admin として他ワークスペースのコースでも通る", func(t *testing.T) {
		h := NewTeachingMaterialHandler(
			usecase.NewTeachingMaterialUseCase(fakeMaterialRepo{}, &fakeCourseRepo{one: publishedCourse}),
		)
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), unaffiliatedSuperAdmin())
		h.ListByCourse(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})

	t.Run("作成は 403（所属ワークスペースが無いので作れない）", func(t *testing.T) {
		h := NewTeachingMaterialHandler(
			usecase.NewTeachingMaterialUseCase(fakeMaterialRepo{}, &fakeCourseRepo{one: publishedCourse}),
		)
		w, c := ctxJSON(http.MethodPost, `{"courseId":5,"title":"X"}`, nil, unaffiliatedSuperAdmin())
		h.Create(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", w.Code)
		}
	})

	t.Run("更新は super_admin として通る", func(t *testing.T) {
		materials := &fakeMaterialRepoH{m: materialInOtherWorkspace(domain.TeachingMaterial{ID: 1, CourseID: 5})}
		h := NewTeachingMaterialHandler(
			usecase.NewTeachingMaterialUseCase(materials, &fakeCourseRepo{one: publishedCourse}),
		)
		w, c := ctxJSON(http.MethodPut, `{"title":"X"}`, idParam("1"), unaffiliatedSuperAdmin())
		h.Update(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})

	t.Run("削除は super_admin として通る", func(t *testing.T) {
		materials := &fakeMaterialRepoH{m: materialInOtherWorkspace(domain.TeachingMaterial{ID: 1, CourseID: 5})}
		h := NewTeachingMaterialHandler(
			usecase.NewTeachingMaterialUseCase(materials, &fakeCourseRepo{one: publishedCourse}),
		)
		_, c := ctxJSON(http.MethodDelete, "", idParam("1"), unaffiliatedSuperAdmin())
		h.Delete(c)
		if c.Writer.Status() != http.StatusNoContent {
			t.Fatalf("want 204, got %d", c.Writer.Status())
		}
	})
}

func Test_未所属actor_章閲覧記録は記録される(t *testing.T) {
	// canRead は super_admin を無条件で許可するため、未所属でも upsert まで到達する。
	views := &recordingChapterViewRepo{}
	materials := &fakeMaterialRepoH{m: materialInOtherWorkspace(domain.TeachingMaterial{ID: 5, CourseID: 9, IsPublished: true})}
	courses := &fakeCourseRepoH{c: courseInOtherWorkspace(domain.Course{ID: 9, IsPublished: true})}
	h := NewChapterViewHandler(usecase.NewRecordChapterViewUseCase(views, materials, courses))

	_, c := ctxJSON(http.MethodPost, "", idParam("5"), unaffiliatedSuperAdmin())
	h.RecordView(c)

	if c.Writer.Status() != http.StatusNoContent {
		t.Fatalf("want 204, got %d", c.Writer.Status())
	}
	if !views.upserted {
		t.Fatal("未所属 super_admin でも閲覧記録は upsert されるべき")
	}
}

func Test_未所属actor_レッスン完了は許可される(t *testing.T) {
	// canRead が super_admin を無条件で許可するため、未所属でも完了にできる。
	progress := &fakeProgressRepoH{}
	materials := &fakeMaterialRepoH{m: materialInOtherWorkspace(domain.TeachingMaterial{ID: 5, CourseID: 9, IsPublished: true})}
	courses := &fakeCourseRepoH{c: courseInOtherWorkspace(domain.Course{ID: 9, IsPublished: true})}
	h := NewLessonProgressHandler(
		usecase.NewMarkLessonCompletedUseCase(progress, materials, courses, &nopActivityRepo{}),
		usecase.NewMarkLessonIncompleteUseCase(progress),
		usecase.NewListLessonProgressUseCase(progress),
	)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUserID, uint64(1))
		c.Set(middleware.ContextKeyCurrentUser, unaffiliatedSuperAdmin())
		c.Next()
	})
	r.POST("/lesson-progress", h.Complete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/lesson-progress", strings.NewReader(`{"teachingMaterialId":5}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
}

func Test_未所属actor_招待取り消しは運営管理者として通る(t *testing.T) {
	invWorkspace := otherWorkspaceID
	repo := &fakeAdminInvRepo{all: []domain.AdminInvitation{{ID: 7, WorkspaceID: &invWorkspace}}}
	h := NewAdminInvitationHandler(nil, nil, nil, usecase.NewCancelAdminInvitationUseCase(repo))

	r := gin.New()
	r.DELETE("/admin/invitations/:id", func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUser, unaffiliatedSuperAdmin())
		h.Cancel(c)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/admin/invitations/7", nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d, body=%s", w.Code, w.Body.String())
	}
}

func Test_未所属actor_招待取り消しはワークスペース管理者だと404(t *testing.T) {
	// company_admin は自ワークスペースの招待しか取り消せない。未所属ならどの招待とも
	// 一致しないため、存在を漏らさない 404 になる。
	invWorkspace := otherWorkspaceID
	repo := &fakeAdminInvRepo{all: []domain.AdminInvitation{{ID: 7, WorkspaceID: &invWorkspace}}}
	h := NewAdminInvitationHandler(nil, nil, nil, usecase.NewCancelAdminInvitationUseCase(repo))

	r := gin.New()
	r.DELETE("/admin/invitations/:id", func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 2, Role: domain.RoleCompanyAdmin, WorkspaceID: nil})
		h.Cancel(c)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/admin/invitations/7", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d, body=%s", w.Code, w.Body.String())
	}
}
