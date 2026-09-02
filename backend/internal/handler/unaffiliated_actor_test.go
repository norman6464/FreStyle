package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// 未所属 actor（users.workspace_id = NULL）が各経路を通ったときの挙動をここに固定する。
//
// # 教材の経路は「何もできない」で揃った
//
// 以前は super_admin なら所属していなくても他ワークスペースの教材を読み書きできた。
// 教材の可否を対象ごとの付与で決めるようにしたので、その抜け道は無くなっている
// （付与は主体を経由して届き、主体はワークスペースへの所属そのものだから）。
// **ここが緩むと、所属していない人が他社の教材を書き換えられる。**

// otherWorkspaceID は未所属 actor から見た「自分のものではないワークスペース」。
const otherWorkspaceID = "0198a000-0000-7000-8000-0000000000f1"

// unaffiliatedSuperAdmin はどのワークスペースにも属さない super_admin。
func unaffiliatedSuperAdmin() *domain.User {
	return &domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: nil}
}

// courseInOtherWorkspace は otherWorkspaceID に属するコースを組み立てる。
func courseInOtherWorkspace(c domain.Course) *domain.Course {
	wid := otherWorkspaceID
	c.WorkspaceID = &wid
	return &c
}

func Test_未所属actor_教材の経路は何も通らない(t *testing.T) {
	// 未所属なら付与も届かないので、読むことも書くこともできない。
	// 経路ごとに結論が割れないこと自体が、この検査の見どころ。
	published := courseInOtherWorkspace(domain.Course{ID: 5, IsPublished: true})
	newTM := func() *TeachingMaterialHandler {
		return NewTeachingMaterialHandler(
			usecase.NewTeachingMaterialUseCase(fakeMaterialRepo{}, &fakeCourseRepo{one: published},
				usecase.NewCheckMaterialPermissionUseCase(&fakeMaterialPerm{})),
		)
	}

	t.Run("コース別一覧は 404（実在を教えない）", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), unaffiliatedSuperAdmin())
		newTM().ListByCourse(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	// 書き込みも返るコードを固定する。「成功ではない」だけだと 500 でも通ってしまい、
	// 拒否のセンチネルが 403 / 404 に写らなくなった退行を拾えない。
	t.Run("作成は 403（作る操作には隠す対象がまだ無い）", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPost, `{"courseId":5,"title":"t"}`, nil, unaffiliatedSuperAdmin())
		newTM().Create(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", w.Code)
		}
	})

	t.Run("更新は 404（実在を教えない）", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPut, `{"title":"t"}`, idParam("9"), unaffiliatedSuperAdmin())
		newTM().Update(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("削除も 404", func(t *testing.T) {
		w, c := ctxJSON(http.MethodDelete, "", idParam("9"), unaffiliatedSuperAdmin())
		newTM().Delete(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
}

func Test_未所属actor_コースの経路は何も通らない(t *testing.T) {
	other := courseInOtherWorkspace(domain.Course{ID: 5, IsPublished: true})
	h := func() *CourseHandler {
		return newCourseHandlerWith(&fakeCourseRepo{one: other}, &fakeChapterViewRepoH{}, &fakeMaterialPerm{})
	}

	t.Run("一覧は空配列", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, unaffiliatedSuperAdmin())
		h().List(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if got := strings.TrimSpace(w.Body.String()); got != "[]" {
			t.Fatalf("want empty array, got %s", got)
		}
	})

	t.Run("詳細は 404（実在を教えない）", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), unaffiliatedSuperAdmin())
		h().Get(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("作成は 403（作る操作には隠す対象がまだ無い）", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPost, `{"title":"t","category":"`+domain.ValidCourseCategories[0]+`"}`, nil, unaffiliatedSuperAdmin())
		h().Create(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", w.Code)
		}
	})

	t.Run("更新は 404（実在を教えない）", func(t *testing.T) {
		w, c := ctxJSON(http.MethodPut, `{"title":"t","category":"`+domain.ValidCourseCategories[0]+`"}`, idParam("5"), unaffiliatedSuperAdmin())
		h().Update(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("削除も 404", func(t *testing.T) {
		w, c := ctxJSON(http.MethodDelete, "", idParam("5"), unaffiliatedSuperAdmin())
		h().Delete(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("最終閲覧章も 404", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("5"), unaffiliatedSuperAdmin())
		h().LastViewed(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
}

func Test_未所属actor_招待取り消しは運営管理者として通る(t *testing.T) {
	invWorkspace := otherWorkspaceID
	repo := &fakeAdminInvRepo{all: []domain.AdminInvitation{{ID: 7, WorkspaceID: &invWorkspace}}}
	h := NewAdminInvitationHandler(nil, nil, usecase.NewCancelAdminInvitationUseCase(repo))

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
	h := NewAdminInvitationHandler(nil, nil, usecase.NewCancelAdminInvitationUseCase(repo))

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
