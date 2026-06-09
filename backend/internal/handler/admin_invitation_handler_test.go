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

func (r *fakeAdminInvRepo) FindPendingByToken(_ context.Context, _ string) (*domain.AdminInvitation, error) {
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

// ===== Create 認可テスト =====
//
// SoD ルール:
//   - SuperAdmin → company_admin の招待のみ可、trainee は禁止
//   - CompanyAdmin → trainee の招待のみ可、自社固定
//   - Trainee → 全部禁止

// fakeAdminInvRepoWithCreate は Create を計測する版。
type fakeAdminInvRepoWithCreate struct {
	fakeAdminInvRepo
	createCalls int
	lastCreate  *domain.AdminInvitation
}

func (r *fakeAdminInvRepoWithCreate) Create(_ context.Context, inv *domain.AdminInvitation) error {
	r.createCalls++
	inv.ID = 100
	r.lastCreate = inv
	return nil
}

// newTestCreateHandler は Create handler 用 harness。usecase は本物 +
// 上の fake repo を inject。sender 等は nil でフォールバックモード。
func newTestCreateHandler(repo *fakeAdminInvRepoWithCreate, currentUser *domain.User) *gin.Engine {
	h := NewAdminInvitationHandler(
		usecase.NewListAdminInvitationsUseCase(repo),
		usecase.NewCreateAdminInvitationUseCase(repo, nil, nil, nil),
		usecase.NewCancelAdminInvitationUseCase(repo),
	)
	r := gin.New()
	r.POST("/admin/invitations", func(c *gin.Context) {
		if currentUser != nil {
			c.Set(middleware.ContextKeyCurrentUser, currentUser)
		}
		h.Create(c)
	})
	return r
}

func postJSON(t *testing.T, r *gin.Engine, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/admin/invitations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAdminInvitationHandler_Create_SuperAdmin_CompanyAdmin_OK(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleSuperAdmin})

	w := postJSON(t, r, `{"companyId":1,"email":"a@b","role":"company_admin"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if repo.createCalls != 1 {
		t.Errorf("expected 1 create, got %d", repo.createCalls)
	}
}

func TestAdminInvitationHandler_Create_SuperAdmin_Trainee_Forbidden(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleSuperAdmin})

	w := postJSON(t, r, `{"companyId":1,"email":"a@b","role":"trainee"}`)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "super_admin_can_only_invite_company_admin") {
		t.Errorf("expected error code, got %s", w.Body.String())
	}
	if repo.createCalls != 0 {
		t.Errorf("create must not be called, got %d", repo.createCalls)
	}
}

func TestAdminInvitationHandler_Create_CompanyAdmin_Trainee_OK_AndCompanyForcedToOwn(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	cid := uint64(42)
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid})

	// 別の company_id を指定しても、handler 側で自社に上書きされること
	w := postJSON(t, r, `{"companyId":999,"email":"t@b","role":"trainee"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if repo.lastCreate == nil || repo.lastCreate.CompanyID != cid {
		t.Errorf("expected companyID forced to %d, got %+v", cid, repo.lastCreate)
	}
}

func TestAdminInvitationHandler_Create_CompanyAdmin_CompanyAdmin_Forbidden(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	cid := uint64(42)
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid})

	w := postJSON(t, r, `{"companyId":42,"email":"a@b","role":"company_admin"}`)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "company_admin_can_only_invite_trainee") {
		t.Errorf("expected error code, got %s", w.Body.String())
	}
}

func TestAdminInvitationHandler_Create_Trainee_Forbidden(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleTrainee})

	w := postJSON(t, r, `{"companyId":1,"email":"a@b","role":"trainee"}`)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAdminInvitationHandler_Create_Unauthenticated(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, nil)

	w := postJSON(t, r, `{"companyId":1,"email":"a@b","role":"trainee"}`)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}
