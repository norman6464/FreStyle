package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// stubUsers は UserRepository の最小 stub。FindByCognitoSub だけ返す。
type stubUsers struct{ user *domain.User }

func (s *stubUsers) FindByCognitoSub(context.Context, string) (*domain.User, error) {
	return s.user, nil
}

func (s *stubUsers) FindByID(context.Context, uint64) (*domain.User, error) { return s.user, nil }

func (s *stubUsers) ListByRole(context.Context, domain.RoleName) ([]domain.User, error) {
	return nil, nil
}

func (s *stubUsers) ListByWorkspaceID(context.Context, string) ([]domain.User, error) {
	return nil, nil
}

func (s *stubUsers) CreateWithOidcIdentity(context.Context, *domain.User, string, string) error {
	return nil
}

func (s *stubUsers) CreateFirstSuperAdminWithOidcIdentity(
	context.Context, *domain.User, string, string,
) (bool, error) {
	return true, nil
}
func (s *stubUsers) EnsureOidcIdentity(context.Context, uint64, string, string) error { return nil }
func (s *stubUsers) FindActiveByEmail(context.Context, string) (*domain.User, error)  { return nil, nil }

func (s *stubUsers) CognitoSubjectByUserID(context.Context, uint64) (string, error) { return "", nil }
func (s *stubUsers) UpdateName(context.Context, uint64, string) error               { return nil }
func (s *stubUsers) UpdateRole(context.Context, uint64, domain.RoleName) error      { return nil }
func (s *stubUsers) UpdateWorkspaceID(context.Context, uint64, *string) error       { return nil }
func (s *stubUsers) UpdateActive(context.Context, uint64, bool) error               { return nil }
func (s *stubUsers) SoftDelete(context.Context, uint64) error                       { return nil }

// stubCompanies は CompanyRepository の最小 stub。FindByWorkspaceID で company / err を返し、
// 問い合わせに使われた workspace_id を記録する。
type stubCompanies struct {
	company *domain.Company
	err     error

	gotWorkspaceID string
}

func (s *stubCompanies) ListAll(context.Context) ([]domain.Company, error) { return nil, nil }
func (s *stubCompanies) FindByID(context.Context, uint64) (*domain.Company, error) {
	return s.company, s.err
}

func (s *stubCompanies) FindByWorkspaceID(_ context.Context, workspaceID string) (*domain.Company, error) {
	s.gotWorkspaceID = workspaceID
	return s.company, s.err
}
func (s *stubCompanies) UpdateActive(context.Context, uint64, bool) error { return nil }

func runCurrentUser(t *testing.T, users *stubUsers, companies *stubCompanies) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(ContextKeyCognitoSub, "sub-123")
	CurrentUser(users, companies)(c)
	return w, c
}

func strPtr(v string) *string { return &v }

func Test_カレントユーザー_無効な会社を遮断(t *testing.T) {
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, WorkspaceID: strPtr("ws-7")}}
	companies := &stubCompanies{company: &domain.Company{ID: 7, IsActive: false, WorkspaceID: strPtr("ws-7")}}

	w, c := runCurrentUser(t, users, companies)

	if companies.gotWorkspaceID != "ws-7" {
		t.Fatalf("会社はユーザーの workspace_id で引くべき: got %q", companies.gotWorkspaceID)
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
	if !c.IsAborted() {
		t.Fatal("無効会社のユーザーは abort されるべき")
	}
}

func Test_カレントユーザー_有効な会社は許可(t *testing.T) {
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, WorkspaceID: strPtr("ws-7")}}
	companies := &stubCompanies{company: &domain.Company{ID: 7, IsActive: true, WorkspaceID: strPtr("ws-7")}}

	_, c := runCurrentUser(t, users, companies)

	if c.IsAborted() {
		t.Fatal("有効会社のユーザーは通すべき")
	}
	if CurrentUserFromContext(c) == nil {
		t.Fatal("currentUser が context にセットされるべき")
	}
}

func Test_カレントユーザー_運営管理者はワークスペースなしでも許可(t *testing.T) {
	// super_admin は workspace_id なし → 会社チェックをスキップして通す。
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleSuperAdmin, IsActive: true, WorkspaceID: nil}}
	companies := &stubCompanies{err: domain.ErrNotFound}

	_, c := runCurrentUser(t, users, companies)

	if companies.gotWorkspaceID != "" {
		t.Fatalf("未所属ユーザーで会社を引くべきではない: got %q", companies.gotWorkspaceID)
	}
	if c.IsAborted() {
		t.Fatal("ワークスペースなしの super_admin は通すべき")
	}
}

func Test_カレントユーザー_会社が見つからなくても許可(t *testing.T) {
	// workspace_id はあるが対応する会社行が無い（データ不整合）→ 弾かない。
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, WorkspaceID: strPtr("ws-99")}}
	companies := &stubCompanies{err: domain.ErrNotFound}

	_, c := runCurrentUser(t, users, companies)

	if c.IsAborted() {
		t.Fatal("会社行なしは弾かない")
	}
}

func Test_カレントユーザー_無効なユーザーを遮断(t *testing.T) {
	// IsActive=false のユーザーは会社が有効でも弾く（即時に利用不可）。
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: false, WorkspaceID: strPtr("ws-7")}}
	companies := &stubCompanies{company: &domain.Company{ID: 7, IsActive: true, WorkspaceID: strPtr("ws-7")}}

	w, c := runCurrentUser(t, users, companies)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
	if !c.IsAborted() {
		t.Fatal("無効ユーザーは abort されるべき")
	}
}
