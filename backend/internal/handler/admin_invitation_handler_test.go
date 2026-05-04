package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// fakeAdminInvRepo は AdminInvitationRepository の最小スタブ。
// list 系のテストで「どのメソッドが呼ばれたか」を確認するため calledWith を記録する。
type fakeAdminInvRepo struct {
	all     []domain.AdminInvitation
	company []domain.AdminInvitation
	called  string
}

func (r *fakeAdminInvRepo) ListAll(_ context.Context) ([]domain.AdminInvitation, error) {
	r.called = "all"
	return r.all, nil
}
func (r *fakeAdminInvRepo) ListByCompanyID(_ context.Context, companyID uint64) ([]domain.AdminInvitation, error) {
	r.called = "company"
	return r.company, nil
}
func (r *fakeAdminInvRepo) Create(_ context.Context, _ *domain.AdminInvitation) error { return nil }
func (r *fakeAdminInvRepo) UpdateStatus(_ context.Context, _ uint64, _ string) error  { return nil }
func (r *fakeAdminInvRepo) FindPendingByEmail(_ context.Context, _ string) (*domain.AdminInvitation, error) {
	return nil, nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestHandler は List handler をテストするための薄い harness。
// CurrentUser middleware の代わりに context に *domain.User を直接 set する。
func newTestHandler(repo *fakeAdminInvRepo, currentUser *domain.User) (*AdminInvitationHandler, *gin.Engine) {
	h := NewAdminInvitationHandler(
		usecase.NewListAdminInvitationsUseCase(repo),
		nil,
		nil,
	)
	r := gin.New()
	r.GET("/admin/invitations", func(c *gin.Context) {
		if currentUser != nil {
			c.Set(middleware.ContextKeyCurrentUser, currentUser)
		}
		h.List(c)
	})
	return h, r
}

func TestAdminInvitationHandler_List_SuperAdmin_AllScope(t *testing.T) {
	repo := &fakeAdminInvRepo{
		all:     []domain.AdminInvitation{{ID: 1}, {ID: 2}},
		company: []domain.AdminInvitation{{ID: 99}},
	}
	_, r := newTestHandler(repo, &domain.User{ID: 1, Role: domain.RoleSuperAdmin})

	req := httptest.NewRequest(http.MethodGet, "/admin/invitations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if repo.called != "all" {
		t.Fatalf("expected ListAll, got %q", repo.called)
	}
	var got []domain.AdminInvitation
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}
}

func TestAdminInvitationHandler_List_SuperAdmin_WithCompanyIDQuery(t *testing.T) {
	repo := &fakeAdminInvRepo{company: []domain.AdminInvitation{{ID: 99}}}
	_, r := newTestHandler(repo, &domain.User{ID: 1, Role: domain.RoleSuperAdmin})

	req := httptest.NewRequest(http.MethodGet, "/admin/invitations?companyId=42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if repo.called != "company" {
		t.Fatalf("expected ListByCompanyID, got %q", repo.called)
	}
}

func TestAdminInvitationHandler_List_CompanyAdmin_AutoFiltersOwnCompany(t *testing.T) {
	repo := &fakeAdminInvRepo{company: []domain.AdminInvitation{{ID: 7}}}
	cid := uint64(123)
	_, r := newTestHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid})

	req := httptest.NewRequest(http.MethodGet, "/admin/invitations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if repo.called != "company" {
		t.Fatalf("expected ListByCompanyID, got %q", repo.called)
	}
}

func TestAdminInvitationHandler_List_CompanyAdmin_WithoutCompanyIDIsForbidden(t *testing.T) {
	repo := &fakeAdminInvRepo{}
	_, r := newTestHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: nil})

	req := httptest.NewRequest(http.MethodGet, "/admin/invitations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestAdminInvitationHandler_List_Trainee_Forbidden(t *testing.T) {
	repo := &fakeAdminInvRepo{}
	_, r := newTestHandler(repo, &domain.User{ID: 1, Role: domain.RoleTrainee})

	req := httptest.NewRequest(http.MethodGet, "/admin/invitations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "forbidden") {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestAdminInvitationHandler_List_Unauthenticated(t *testing.T) {
	repo := &fakeAdminInvRepo{}
	_, r := newTestHandler(repo, nil) // current user 未設定

	req := httptest.NewRequest(http.MethodGet, "/admin/invitations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}
