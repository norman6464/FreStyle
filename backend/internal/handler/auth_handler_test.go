package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// fakeUserRepo は AuthHandler.upsertUserFromIDToken のテスト用 stub。
type fakeUserRepo struct {
	existingBySub map[string]*domain.User
	created       *domain.User
	createErr     error
	updateRoleID  uint64
	updateRoleVal string
}

func (r *fakeUserRepo) FindByCognitoSub(_ context.Context, sub string) (*domain.User, error) {
	if u, ok := r.existingBySub[sub]; ok {
		return u, nil
	}
	return nil, nil
}
func (r *fakeUserRepo) FindByID(_ context.Context, _ uint64) (*domain.User, error) {
	return nil, nil
}
func (r *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	u.ID = 7
	r.created = u
	return nil
}
func (r *fakeUserRepo) UpdateDisplayName(_ context.Context, _ uint64, _ string) error { return nil }
func (r *fakeUserRepo) UpdateRole(_ context.Context, id uint64, role string) error {
	r.updateRoleID, r.updateRoleVal = id, role
	return nil
}

// fakeInvitationRepo は AdminInvitationRepository の最小スタブ。
// FindPendingByEmail だけ振る舞いをカスタムにしてテストする。
type fakeInvitationRepo struct {
	pendingByEmail map[string]*domain.AdminInvitation
	updatedID      uint64
	updatedStatus  string
}

func (r *fakeInvitationRepo) ListAll(_ context.Context) ([]domain.AdminInvitation, error) {
	return nil, nil
}
func (r *fakeInvitationRepo) ListByCompanyID(_ context.Context, _ uint64) ([]domain.AdminInvitation, error) {
	return nil, nil
}
func (r *fakeInvitationRepo) FindPendingByEmail(_ context.Context, email string) (*domain.AdminInvitation, error) {
	if v, ok := r.pendingByEmail[email]; ok {
		return v, nil
	}
	return nil, nil
}
func (r *fakeInvitationRepo) FindPendingByToken(_ context.Context, _ string) (*domain.AdminInvitation, error) {
	return nil, nil
}
func (r *fakeInvitationRepo) Create(_ context.Context, _ *domain.AdminInvitation) error { return nil }
func (r *fakeInvitationRepo) UpdateStatus(_ context.Context, id uint64, status string) error {
	r.updatedID, r.updatedStatus = id, status
	return nil
}

// makeIDToken は claims を JSON にして JWT 形式（ヘッダ.ペイロード.署名）にエンコードする。
// 署名は middleware.DecodeClaims が検証しないので空文字でよい。
func makeIDToken(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	body := base64.RawURLEncoding.EncodeToString(payload)
	return header + "." + body + "."
}

// newTestAuthHandler はテスト用 AuthHandler を組み立てる。tokens は使わない。
func newTestAuthHandler(users *fakeUserRepo, invitations *fakeInvitationRepo) *AuthHandler {
	return &AuthHandler{users: users, invitations: invitations}
}

func init() {
	gin.SetMode(gin.TestMode)
}

// テスト用に空の gin.Context を返す（c.Request.Context() が呼ばれるので Request も埋める）。
func newGinCtx() *gin.Context {
	c, _ := gin.CreateTestContext(nil)
	c.Request = mustNewRequest()
	return c
}

func TestUpsertUserFromIDToken_BlocksNewUserWithoutInvitationOrAdmin(t *testing.T) {
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{}
	h := newTestAuthHandler(users, invs)

	idToken := makeIDToken(t, map[string]any{
		"sub":   "new-google-user",
		"email": "stranger@example.com",
		// cognito:groups なし、招待もなし → 拒否されるべき
	})

	allowed := h.upsertUserFromIDToken(newGinCtx(), idToken)
	if allowed {
		t.Fatalf("expected allowed=false for non-invited non-admin signup")
	}
	if users.created != nil {
		t.Fatalf("user must NOT be created when no invitation/admin")
	}
}

func TestUpsertUserFromIDToken_AllowsCognitoAdminWithoutInvitation(t *testing.T) {
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{}
	h := newTestAuthHandler(users, invs)

	idToken := makeIDToken(t, map[string]any{
		"sub":            "first-super-admin",
		"email":          "ops@example.com",
		"cognito:groups": []string{"admin"},
	})

	allowed := h.upsertUserFromIDToken(newGinCtx(), idToken)
	if !allowed {
		t.Fatalf("Cognito group admin must be allowed even without invitation")
	}
	if users.created == nil || users.created.Role != domain.RoleSuperAdmin {
		t.Fatalf("expected super_admin to be created, got %+v", users.created)
	}
}

func TestUpsertUserFromIDToken_AllowsInvitedUser_AppliesRoleAndCompany(t *testing.T) {
	users := &fakeUserRepo{}
	cid := uint64(42)
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"trainee@example.com": {
				ID: 99, CompanyID: cid, Email: "trainee@example.com",
				Role: domain.RoleTrainee, DisplayName: "佐藤", Status: domain.InvitationStatusPending,
			},
		},
	}
	h := newTestAuthHandler(users, invs)

	idToken := makeIDToken(t, map[string]any{
		"sub":   "google-trainee-1",
		"email": "trainee@example.com",
	})

	allowed := h.upsertUserFromIDToken(newGinCtx(), idToken)
	if !allowed {
		t.Fatalf("invited user must be allowed")
	}
	if users.created == nil {
		t.Fatalf("expected user to be created")
	}
	if users.created.Role != domain.RoleTrainee {
		t.Errorf("expected role=trainee, got %q", users.created.Role)
	}
	if users.created.CompanyID == nil || *users.created.CompanyID != cid {
		t.Errorf("expected company_id=%d, got %+v", cid, users.created.CompanyID)
	}
	if users.created.DisplayName != "佐藤" {
		t.Errorf("expected displayName=佐藤, got %q", users.created.DisplayName)
	}
	// 招待は accepted にマークされる
	if invs.updatedID != 99 || invs.updatedStatus != domain.InvitationStatusAccepted {
		t.Errorf("invitation must be marked accepted, got id=%d status=%q", invs.updatedID, invs.updatedStatus)
	}
}

func TestUpsertUserFromIDToken_ExistingUser_AlwaysAllowed(t *testing.T) {
	existing := &domain.User{ID: 5, CognitoSub: "existing", Email: "u@example.com", Role: domain.RoleTrainee}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"existing": existing}}
	invs := &fakeInvitationRepo{}
	h := newTestAuthHandler(users, invs)

	idToken := makeIDToken(t, map[string]any{
		"sub":   "existing",
		"email": "u@example.com",
	})

	allowed := h.upsertUserFromIDToken(newGinCtx(), idToken)
	if !allowed {
		t.Fatalf("existing user must always be allowed (no invitation re-check)")
	}
	if users.created != nil {
		t.Fatalf("existing user must not trigger Create")
	}
}

func TestUpsertUserFromIDToken_ExistingUser_PromotedByCognitoAdmin(t *testing.T) {
	existing := &domain.User{ID: 5, CognitoSub: "existing", Email: "u@example.com", Role: domain.RoleTrainee}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"existing": existing}}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})

	idToken := makeIDToken(t, map[string]any{
		"sub":            "existing",
		"email":          "u@example.com",
		"cognito:groups": []string{"admin"},
	})
	if !h.upsertUserFromIDToken(newGinCtx(), idToken) {
		t.Fatal("must be allowed")
	}
	if users.updateRoleID != 5 || users.updateRoleVal != domain.RoleSuperAdmin {
		t.Fatalf("expected role promoted to super_admin, got id=%d role=%q", users.updateRoleID, users.updateRoleVal)
	}
}

func TestUpsertUserFromIDToken_RejectsMalformedToken(t *testing.T) {
	h := newTestAuthHandler(&fakeUserRepo{}, &fakeInvitationRepo{})
	if h.upsertUserFromIDToken(newGinCtx(), "not-a-jwt") {
		t.Fatal("malformed token must be rejected")
	}
}

// dummy: 使用していないが strings import のリンタを満たすために残す（今後 assert で使う）
var _ = strings.Contains
