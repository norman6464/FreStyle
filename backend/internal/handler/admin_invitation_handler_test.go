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
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
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

// newTestHandler は List handler をテストするための薄い harness。
// CurrentUser middleware の代わりに context に *domain.User を直接 set する。
func newTestHandler(repo *fakeAdminInvRepo, currentUser *domain.User) (*AdminInvitationHandler, *gin.Engine) {
	h := NewAdminInvitationHandler(
		usecase.NewListAdminInvitationsUseCase(repo),
		nil,
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

func Test_招待ハンドラ_一覧_運営管理者_会社IDクエリ付き(t *testing.T) {
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

func Test_招待ハンドラ_一覧_会社管理者_自社に自動絞り込み(t *testing.T) {
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

func Test_招待ハンドラ_一覧_会社管理者_会社IDなしは禁止(t *testing.T) {
	repo := &fakeAdminInvRepo{}
	_, r := newTestHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: nil})

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
	return newTestCreateHandlerWithTemp(repo, currentUser, nil)
}

// fakeTempPasswordCreator は TemporaryPasswordCreator のテスト用スタブ。
type fakeTempPasswordCreator struct {
	pw       string
	err      error
	gotEmail string
	gotName  string
}

func (f *fakeTempPasswordCreator) CreateWithTemporaryPassword(_ context.Context, email, name string) (string, error) {
	f.gotEmail, f.gotName = email, name
	return f.pw, f.err
}

func newTestCreateHandlerWithTemp(repo *fakeAdminInvRepoWithCreate, currentUser *domain.User, creator usecase.TemporaryPasswordCreator) *gin.Engine {
	var tempUC *usecase.CreateTemporaryPasswordInvitationUseCase
	if creator != nil {
		tempUC = usecase.NewCreateTemporaryPasswordInvitationUseCase(repo, creator)
	}
	h := NewAdminInvitationHandler(
		usecase.NewListAdminInvitationsUseCase(repo),
		usecase.NewCreateAdminInvitationUseCase(repo, nil, nil, nil),
		tempUC,
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

func Test_招待ハンドラ_作成_運営管理者_会社管理者_正常系(t *testing.T) {
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

func Test_招待ハンドラ_作成_運営管理者_trainee_禁止(t *testing.T) {
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

func Test_招待ハンドラ_作成_会社管理者_trainee_正常系かつ自社に強制(t *testing.T) {
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

func Test_招待ハンドラ_作成_会社管理者_会社管理者_禁止(t *testing.T) {
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

func Test_招待ハンドラ_作成_traineeは禁止(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleTrainee})

	w := postJSON(t, r, `{"companyId":1,"email":"a@b","role":"trainee"}`)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
}

func Test_招待ハンドラ_作成_未認証(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	r := newTestCreateHandler(repo, nil)

	w := postJSON(t, r, `{"companyId":1,"email":"a@b","role":"trainee"}`)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func Test_招待ハンドラ_初期パスワード方式_成功で一時パスワードを返す(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	cid := uint64(42)
	creator := &fakeTempPasswordCreator{pw: "Temp-Pass-9!"}
	r := newTestCreateHandlerWithTemp(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid}, creator)

	w := postJSON(t, r, `{"companyId":42,"email":"np@b","role":"trainee","method":"temporary_password"}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	var got struct {
		Invitation        *domain.AdminInvitation `json:"invitation"`
		TemporaryPassword string                  `json:"temporaryPassword"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("応答が期待の JSON 形ではない: %v body = %s", err, w.Body.String())
	}
	if got.TemporaryPassword != "Temp-Pass-9!" {
		t.Errorf("temporaryPassword = %q", got.TemporaryPassword)
	}
	if got.Invitation == nil {
		t.Error("invitation が欠落している")
	}
	if creator.gotEmail != "np@b" {
		t.Errorf("creator got email %q", creator.gotEmail)
	}
	// 招待行も pending で作られる（ログイン時の招待ゲート用）。
	if repo.lastCreate == nil || repo.lastCreate.Role != domain.RoleTrainee || repo.lastCreate.CompanyID != cid {
		t.Errorf("invitation row not created correctly: %+v", repo.lastCreate)
	}
}

func Test_招待ハンドラ_初期パスワード方式_未構成なら400(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	cid := uint64(42)
	// creator=nil → tempCreate usecase が nil → 未構成。usecase の ErrUnavailable と同じ 400。
	r := newTestCreateHandlerWithTemp(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid}, nil)

	w := postJSON(t, r, `{"companyId":42,"email":"np@b","role":"trainee","method":"temporary_password"}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "temporary_password_not_configured") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func Test_招待ハンドラ_初期パスワード方式_既存ユーザーは409(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	cid := uint64(42)
	creator := &fakeTempPasswordCreator{err: cognito.ErrUserAlreadyExists}
	r := newTestCreateHandlerWithTemp(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid}, creator)

	w := postJSON(t, r, `{"companyId":42,"email":"dup@b","role":"trainee","method":"temporary_password"}`)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "user_already_exists") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func Test_招待ハンドラ_初期パスワード方式もSoDを守る(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	cid := uint64(42)
	creator := &fakeTempPasswordCreator{pw: "x"}
	r := newTestCreateHandlerWithTemp(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid}, creator)

	// company_admin が company_admin を初期パスワードで招待 → 方式に関わらず 403。
	w := postJSON(t, r, `{"companyId":42,"email":"a@b","role":"company_admin","method":"temporary_password"}`)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	if creator.gotEmail != "" {
		t.Error("SoD 違反なのに Cognito ユーザーが作られた")
	}
}

func Test_招待ハンドラ_未知のmethodは400(t *testing.T) {
	repo := &fakeAdminInvRepoWithCreate{}
	cid := uint64(42)
	r := newTestCreateHandler(repo, &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid})

	w := postJSON(t, r, `{"companyId":42,"email":"t@b","role":"trainee","method":"bogus"}`)

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
