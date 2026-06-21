package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// stubUsers は UserRepository の最小 stub。FindByCognitoSub だけ返す。
type stubUsers struct{ user *domain.User }

func (s *stubUsers) FindByCognitoSub(context.Context, string) (*domain.User, error) {
	return s.user, nil
}

func (s *stubUsers) FindByID(context.Context, uint64) (*domain.User, error) { return s.user, nil }

func (s *stubUsers) ListByRole(context.Context, string) ([]domain.User, error) { return nil, nil }

func (s *stubUsers) ListByCompanyID(context.Context, uint64) ([]domain.User, error) { return nil, nil }
func (s *stubUsers) Create(context.Context, *domain.User) error                     { return nil }
func (s *stubUsers) UpdateAiChatEnabled(context.Context, uint64, *bool) error       { return nil }
func (s *stubUsers) UpdateName(context.Context, uint64, string) error               { return nil }
func (s *stubUsers) UpdateRole(context.Context, uint64, string) error               { return nil }
func (s *stubUsers) UpdateCompanyID(context.Context, uint64, uint64) error          { return nil }
func (s *stubUsers) UpdateActive(context.Context, uint64, bool) error               { return nil }
func (s *stubUsers) SoftDelete(context.Context, uint64) error                       { return nil }
func (s *stubUsers) MarkOnboarded(context.Context, uint64) error                    { return nil }

// stubCompanies は CompanyRepository の最小 stub。FindByID で company / err を返す。
type stubCompanies struct {
	company *domain.Company
	err     error
}

func (s *stubCompanies) ListAll(context.Context) ([]domain.Company, error) { return nil, nil }
func (s *stubCompanies) FindByID(context.Context, uint64) (*domain.Company, error) {
	return s.company, s.err
}
func (s *stubCompanies) UpdateAiChatEnabled(context.Context, uint64, bool) error { return nil }
func (s *stubCompanies) UpdateActive(context.Context, uint64, bool) error        { return nil }

func runCurrentUser(t *testing.T, users *stubUsers, companies *stubCompanies) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(ContextKeyCognitoSub, "sub-123")
	CurrentUser(users, companies)(c)
	return w, c
}

func uintPtr(v uint64) *uint64 { return &v }

func Test_カレントユーザー_無効な会社を遮断(t *testing.T) {
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, CompanyID: uintPtr(7)}}
	companies := &stubCompanies{company: &domain.Company{ID: 7, IsActive: false}}

	w, c := runCurrentUser(t, users, companies)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
	if !c.IsAborted() {
		t.Fatal("無効会社のユーザーは abort されるべき")
	}
}

func Test_カレントユーザー_有効な会社は許可(t *testing.T) {
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, CompanyID: uintPtr(7)}}
	companies := &stubCompanies{company: &domain.Company{ID: 7, IsActive: true}}

	_, c := runCurrentUser(t, users, companies)

	if c.IsAborted() {
		t.Fatal("有効会社のユーザーは通すべき")
	}
	if CurrentUserFromContext(c) == nil {
		t.Fatal("currentUser が context にセットされるべき")
	}
}

func Test_カレントユーザー_運営管理者は会社なしでも許可(t *testing.T) {
	// super_admin は company_id なし → 会社チェックをスキップして通す。
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleSuperAdmin, IsActive: true, CompanyID: nil}}
	companies := &stubCompanies{err: gorm.ErrRecordNotFound}

	_, c := runCurrentUser(t, users, companies)

	if c.IsAborted() {
		t.Fatal("company なしの super_admin は通すべき")
	}
}

func Test_カレントユーザー_会社が見つからなくても許可(t *testing.T) {
	// company_id はあるが会社行が無い（データ不整合）→ 弾かない。
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, CompanyID: uintPtr(99)}}
	companies := &stubCompanies{err: gorm.ErrRecordNotFound}

	_, c := runCurrentUser(t, users, companies)

	if c.IsAborted() {
		t.Fatal("会社行なしは弾かない")
	}
}

func Test_カレントユーザー_無効なユーザーを遮断(t *testing.T) {
	// IsActive=false のユーザーは会社が有効でも弾く（即時に利用不可）。
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: false, CompanyID: uintPtr(7)}}
	companies := &stubCompanies{company: &domain.Company{ID: 7, IsActive: true}}

	w, c := runCurrentUser(t, users, companies)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
	if !c.IsAborted() {
		t.Fatal("無効ユーザーは abort されるべき")
	}
}
