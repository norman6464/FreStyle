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
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeCompanyAppRepo は repository.CompanyApplicationRepository の最小 fake。
type fakeCompanyAppRepo struct {
	listRows  []domain.CompanyApplication
	listErr   error
	updateErr error
}

func (f *fakeCompanyAppRepo) Create(context.Context, *domain.CompanyApplication) error { return nil }

func (f *fakeCompanyAppRepo) ListAll(context.Context) ([]domain.CompanyApplication, error) {
	return f.listRows, f.listErr
}

func (f *fakeCompanyAppRepo) UpdateStatus(context.Context, uint64, string) error { return f.updateErr }

func newCompanyAppHandler(repo repository.CompanyApplicationRepository) *CompanyApplicationHandler {
	// Create usecase の userRepo / notifRepo は nil。本テストは「入力不正で repo 到達前に
	// 返る」分岐のみ Create を叩くため nil で安全（成功パスは結合テスト側）。
	return NewCompanyApplicationHandler(
		usecase.NewCreateCompanyApplicationUseCase(repo, nil, nil),
		usecase.NewListCompanyApplicationsUseCase(repo),
		usecase.NewUpdateCompanyApplicationStatusUseCase(repo),
	)
}

// ctxJSON は JSON body / params / current user を持つ test context を作る。
func ctxJSON(method, body string, params gin.Params, user *domain.User) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	if user != nil {
		c.Set(middleware.ContextKeyCurrentUser, user)
	}
	return w, c
}

func superAdmin() *domain.User   { return &domain.User{ID: 1, Role: domain.RoleSuperAdmin} }
func companyAdmin() *domain.User { return &domain.User{ID: 2, Role: domain.RoleCompanyAdmin} }

// --- List ---

func Test_会社申請ハンドラ_一覧_未認証(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	w, c := ctxJSON(http.MethodGet, "", nil, nil)
	h.List(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func Test_会社申請ハンドラ_一覧_禁止(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	w, c := ctxJSON(http.MethodGet, "", nil, companyAdmin())
	h.List(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func Test_会社申請ハンドラ_一覧_正常系(t *testing.T) {
	repo := &fakeCompanyAppRepo{listRows: []domain.CompanyApplication{{ID: 1, CompanyName: "A"}}}
	h := newCompanyAppHandler(repo)
	w, c := ctxJSON(http.MethodGet, "", nil, superAdmin())
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "\"A\"") {
		t.Errorf("body should contain the row, got %s", w.Body.String())
	}
}

func Test_会社申請ハンドラ_一覧_リポジトリエラーは500(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{listErr: context.DeadlineExceeded})
	w, c := ctxJSON(http.MethodGet, "", nil, superAdmin())
	h.List(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

// --- UpdateStatus ---

func idParam(v string) gin.Params { return gin.Params{{Key: "id", Value: v}} }

func Test_会社申請ハンドラ_ステータス更新_禁止(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	w, c := ctxJSON(http.MethodPatch, `{"status":"approved"}`, idParam("1"), companyAdmin())
	h.UpdateStatus(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func Test_会社申請ハンドラ_ステータス更新_不正なID(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	w, c := ctxJSON(http.MethodPatch, `{"status":"approved"}`, idParam("abc"), superAdmin())
	h.UpdateStatus(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_会社申請ハンドラ_ステータス更新_不正なJSON(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	w, c := ctxJSON(http.MethodPatch, `{`, idParam("1"), superAdmin())
	h.UpdateStatus(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_会社申請ハンドラ_ステータス更新_不正なステータスは400(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	w, c := ctxJSON(http.MethodPatch, `{"status":"weird"}`, idParam("1"), superAdmin())
	h.UpdateStatus(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_会社申請ハンドラ_ステータス更新_正常系(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	_, c := ctxJSON(http.MethodPatch, `{"status":"approved"}`, idParam("1"), superAdmin())
	h.UpdateStatus(c)
	// c.Status(204) は body 無しのため recorder にフラッシュされない。gin の Writer が
	// 保持する intended status を見る。
	if c.Writer.Status() != http.StatusNoContent {
		t.Fatalf("want 204, got %d", c.Writer.Status())
	}
}

func Test_会社申請ハンドラ_ステータス更新_リポジトリエラーは500(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{updateErr: context.DeadlineExceeded})
	w, c := ctxJSON(http.MethodPatch, `{"status":"approved"}`, idParam("1"), superAdmin())
	h.UpdateStatus(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

// --- Create（公開 / 入力検証のみ）---

func Test_会社申請ハンドラ_作成_不正なJSON(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	w, c := ctxJSON(http.MethodPost, `{`, nil, nil)
	h.Create(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_会社申請ハンドラ_作成_不正なメールは400(t *testing.T) {
	h := newCompanyAppHandler(&fakeCompanyAppRepo{})
	body := `{"companyName":"X","applicantName":"Y","email":"not-an-email"}`
	w, c := ctxJSON(http.MethodPost, body, nil, nil)
	h.Create(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
