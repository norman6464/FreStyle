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
	"gorm.io/gorm"
)

// fakeCompanyCounter は CompanyMemberCounter の最小 fake。
type fakeCompanyCounter struct {
	rows []repository.CompanyMemberCount
}

func (f *fakeCompanyCounter) CountMembersByCompany(context.Context) ([]repository.CompanyMemberCount, error) {
	return f.rows, nil
}

// fakeCompanyRepo は repository.CompanyRepository の最小 fake。
type fakeCompanyRepo struct {
	rows         []domain.Company
	err          error
	activeCalled bool
	gotActiveID  uint64
	gotActive    bool
}

func (f *fakeCompanyRepo) ListAll(context.Context) ([]domain.Company, error) { return f.rows, f.err }

func (f *fakeCompanyRepo) FindByID(context.Context, uint64) (*domain.Company, error) {
	return nil, f.err
}

func (f *fakeCompanyRepo) UpdateAiChatEnabled(context.Context, uint64, bool) error { return nil }
func (f *fakeCompanyRepo) UpdateActive(_ context.Context, id uint64, active bool) error {
	f.activeCalled = true
	f.gotActiveID = id
	f.gotActive = active
	return f.err
}

func newAdminCompanyHandler(repo *fakeCompanyRepo) *AdminCompanyHandler {
	return NewAdminCompanyHandler(
		usecase.NewListCompaniesUseCase(repo),
		usecase.NewListCompanyStatsUseCase(repo, &fakeCompanyCounter{}),
		usecase.NewSetCompanyActiveUseCase(repo),
	)
}

func Test_会社管理ハンドラ_横断ビュー_super_admin_正常系(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 1, Role: domain.RoleSuperAdmin})
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/companies/stats", nil)
	newAdminCompanyHandler(&fakeCompanyRepo{rows: []domain.Company{{ID: 1, Name: "Co"}}}).Stats(c)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func Test_会社管理ハンドラ_横断ビュー_super_admin以外は禁止(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 2, Role: domain.RoleCompanyAdmin})
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/companies/stats", nil)
	newAdminCompanyHandler(&fakeCompanyRepo{}).Stats(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func Test_会社管理ハンドラ_一覧_正常系(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/companies", nil)
	newAdminCompanyHandler(&fakeCompanyRepo{rows: []domain.Company{{ID: 1, Name: "Co"}}}).List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func Test_会社管理ハンドラ_一覧_エラー(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/companies", nil)
	newAdminCompanyHandler(&fakeCompanyRepo{err: context.DeadlineExceeded}).List(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func patchCompanyActive(t *testing.T, actor *domain.User, id, body string, repo *fakeCompanyRepo) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/admin/companies/"+id+"/active", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: id}}
	if actor != nil {
		c.Set(middleware.ContextKeyCurrentUser, actor)
	}
	newAdminCompanyHandler(repo).SetActive(c)
	return w
}

func Test_会社管理ハンドラ_有効化_運営管理者_正常系(t *testing.T) {
	repo := &fakeCompanyRepo{}
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}

	w := patchCompanyActive(t, actor, "5", `{"active":false}`, repo)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (%s)", w.Code, w.Body.String())
	}
	if !repo.activeCalled || repo.gotActiveID != 5 || repo.gotActive != false {
		t.Fatalf("repo not called correctly: %+v", repo)
	}
}

func Test_会社管理ハンドラ_有効化_運営管理者以外_禁止(t *testing.T) {
	for _, role := range []string{domain.RoleCompanyAdmin, domain.RoleTrainee} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeCompanyRepo{}
			w := patchCompanyActive(t, &domain.User{ID: 2, Role: role}, "5", `{"active":false}`, repo)
			if w.Code != http.StatusForbidden {
				t.Fatalf("want 403, got %d", w.Code)
			}
			if repo.activeCalled {
				t.Fatal("非 super_admin で UpdateActive を呼んではならない")
			}
		})
	}
}

func Test_会社管理ハンドラ_有効化_不正なボディ_400(t *testing.T) {
	repo := &fakeCompanyRepo{}
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}
	w := patchCompanyActive(t, actor, "5", `{}`, repo)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_会社管理ハンドラ_有効化_見つからない(t *testing.T) {
	// 存在しない会社 ID（0 件更新）は repository が ErrRecordNotFound を返し、handler が 404 にマップ。
	repo := &fakeCompanyRepo{err: gorm.ErrRecordNotFound}
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}
	w := patchCompanyActive(t, actor, "999", `{"active":false}`, repo)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}
