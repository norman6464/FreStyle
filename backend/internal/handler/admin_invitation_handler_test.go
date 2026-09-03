package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	all         []domain.AdminInvitation
	workspace   []domain.AdminInvitation
	called      string
	workspaceID string
}

func (r *fakeAdminInvRepo) ListAll(_ context.Context) ([]domain.AdminInvitation, error) {
	r.called = "all"
	return r.all, nil
}

func (r *fakeAdminInvRepo) ListByWorkspaceID(_ context.Context, workspaceID string) ([]domain.AdminInvitation, error) {
	r.called = "workspace"
	r.workspaceID = workspaceID
	return r.workspace, nil
}
func (r *fakeAdminInvRepo) Create(_ context.Context, _ *domain.AdminInvitation) error { return nil }
func (r *fakeAdminInvRepo) UpdateStatus(_ context.Context, _ uint64, _ string) error  { return nil }
func (r *fakeAdminInvRepo) FindPendingByEmail(_ context.Context, _ string) (*domain.AdminInvitation, error) {
	return nil, nil
}

func (r *fakeAdminInvRepo) FindPendingByToken(_ context.Context, _ string) (*domain.AdminInvitation, error) {
	return nil, nil
}

func (r *fakeAdminInvRepo) FindByID(_ context.Context, id uint64) (*domain.AdminInvitation, error) {
	for i := range r.all {
		if r.all[i].ID == id {
			return &r.all[i], nil
		}
	}
	return nil, nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

// invWorkspaceID はテスト用のワークスペース ID を連番から組み立てる。
// invitations が持つのは workspace_id だけなので、テストも UUID 形の値だけで話を通す。
func invWorkspaceID(n uint64) string {
	return fmt.Sprintf("0198a000-0000-7000-8000-%012d", n)
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

func Test_招待ハンドラ_一覧_運営管理者_全件(t *testing.T) {
	repo := &fakeAdminInvRepo{
		all:       []domain.AdminInvitation{{ID: 1}, {ID: 2}},
		workspace: []domain.AdminInvitation{{ID: 99}},
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

func Test_招待ハンドラ_一覧_ワークスペース管理者_自分の所属に自動絞り込み(t *testing.T) {
	repo := &fakeAdminInvRepo{workspace: []domain.AdminInvitation{{ID: 7}}}
	wid := invWorkspaceID(123)
	_, r := newTestHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: &wid})

	req := httptest.NewRequest(http.MethodGet, "/admin/invitations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if repo.called != "workspace" {
		t.Fatalf("expected ListByWorkspaceID, got %q", repo.called)
	}
	if repo.workspaceID != wid {
		t.Fatalf("expected workspaceID %q to be delegated, got %q", wid, repo.workspaceID)
	}
}

func Test_招待ハンドラ_一覧_ワークスペース管理者_ワークスペース未所属は禁止(t *testing.T) {
	repo := &fakeAdminInvRepo{}
	_, r := newTestHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: nil})

	req := httptest.NewRequest(http.MethodGet, "/admin/invitations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func Test_招待ハンドラ_一覧_traineeは禁止(t *testing.T) {
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

func Test_招待ハンドラ_一覧_未認証(t *testing.T) {
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
// 招待できるのは「ワークスペース管理者が自分の所属へ trainee を招く」場合だけ:
//   - SuperAdmin → 招待できない（テナントを横断して招ける中央管理者は置かない）
//   - CompanyAdmin → trainee の招待のみ可、招待先は自分の所属に固定
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
// newTestCreateHandler は Create handler 用 harness。usecase は本物を使う。
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

// 運営管理者は所属を持たないので招待先が決まらない。招待先をリクエストで指定させる
// 入口も無いため、どのロール宛でも 403 で止める（テナントを横断した招待を残さない）。
func Test_招待ハンドラ_作成_運営管理者は招待できない(t *testing.T) {
	for _, role := range []string{"company_admin", "trainee"} {
		t.Run(role+" 宛でも 403", func(t *testing.T) {
			repo := &fakeAdminInvRepoWithCreate{}
			r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleSuperAdmin})

			w := postJSON(t, r, fmt.Sprintf(`{"email":"a@b","role":%q}`, role))

			if w.Code != http.StatusForbidden {
				t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), "forbidden") {
				t.Errorf("expected error code, got %s", w.Body.String())
			}
			if repo.createCalls != 0 {
				t.Errorf("create must not be called, got %d", repo.createCalls)
			}
		})
	}
}

func Test_招待ハンドラ_作成_ワークスペース管理者_trainee_正常系かつ自分の所属に固定(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	wid := invWorkspaceID(42)
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: &wid})

	// 招待先を指定する入口は無いので、actor 自身の所属がそのまま招待先になること。
	w := postJSON(t, r, `{"email":"t@b","role":"trainee"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if repo.lastCreate == nil || repo.lastCreate.WorkspaceID == nil || *repo.lastCreate.WorkspaceID != wid {
		t.Errorf("expected workspaceID forced to %q, got %+v", wid, repo.lastCreate)
	}
}

func Test_招待ハンドラ_作成_ワークスペース管理者_ワークスペース未所属は禁止(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: nil})

	w := postJSON(t, r, `{"email":"t@b","role":"trainee"}`)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if repo.createCalls != 0 {
		t.Errorf("create must not be called, got %d", repo.createCalls)
	}
}

func Test_招待ハンドラ_作成_ワークスペース管理者_管理者の招待は禁止(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	wid := invWorkspaceID(42)
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: &wid})

	w := postJSON(t, r, `{"email":"a@b","role":"company_admin"}`)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "company_admin_can_only_invite_trainee") {
		t.Errorf("expected error code, got %s", w.Body.String())
	}
}

func Test_招待ハンドラ_作成_traineeは禁止(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleTrainee})

	w := postJSON(t, r, `{"email":"a@b","role":"trainee"}`)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
}

func Test_招待ハンドラ_作成_未認証(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, nil)

	w := postJSON(t, r, `{"email":"a@b","role":"trainee"}`)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func Test_招待ハンドラ_未知のmethodは400(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	wid := invWorkspaceID(42)
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: &wid})

	w := postJSON(t, r, `{"email":"t@b","role":"trainee","method":"bogus"}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	// binding:"oneof=..." による 400。エラー本文は binding のメッセージ（method フィールドに言及）。
	if !strings.Contains(w.Body.String(), "Method") && !strings.Contains(w.Body.String(), "method") {
		t.Errorf("body = %s", w.Body.String())
	}
	if repo.createCalls != 0 {
		t.Error("未知 method で招待行が作られた")
	}
}
