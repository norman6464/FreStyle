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
	existingBySub        map[string]*domain.User
	created              *domain.User
	createErr            error
	updateRoleID         uint64
	updateRoleVal        string
	updateCompanyID      uint64
	updateCompanyVal     uint64
	updateDisplayNameID  uint64
	updateDisplayNameVal string
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
func (r *fakeUserRepo) UpdateDisplayName(_ context.Context, id uint64, name string) error {
	r.updateDisplayNameID, r.updateDisplayNameVal = id, name
	return nil
}
func (r *fakeUserRepo) UpdateRole(_ context.Context, id uint64, role string) error {
	r.updateRoleID, r.updateRoleVal = id, role
	return nil
}
func (r *fakeUserRepo) UpdateCompanyID(_ context.Context, id uint64, companyID uint64) error {
	r.updateCompanyID, r.updateCompanyVal = id, companyID
	return nil
}
func (r *fakeUserRepo) MarkOnboarded(_ context.Context, _ uint64) error {
	return nil
}

// fakeInvitationRepo は AdminInvitationRepository の最小スタブ。
// FindPendingByEmail / FindPendingByToken の振る舞いをカスタムにしてテストする。
type fakeInvitationRepo struct {
	pendingByEmail map[string]*domain.AdminInvitation
	pendingByToken map[string]*domain.AdminInvitation
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
func (r *fakeInvitationRepo) FindPendingByToken(_ context.Context, token string) (*domain.AdminInvitation, error) {
	if v, ok := r.pendingByToken[token]; ok {
		return v, nil
	}
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

	allowed := h.upsertUserFromIDToken(newGinCtx(), idToken, "")
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

	allowed := h.upsertUserFromIDToken(newGinCtx(), idToken, "")
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

	allowed := h.upsertUserFromIDToken(newGinCtx(), idToken, "")
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

	allowed := h.upsertUserFromIDToken(newGinCtx(), idToken, "")
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
	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.updateRoleID != 5 || users.updateRoleVal != domain.RoleSuperAdmin {
		t.Fatalf("expected role promoted to super_admin, got id=%d role=%q", users.updateRoleID, users.updateRoleVal)
	}
}

func TestUpsertUserFromIDToken_RejectsMalformedToken(t *testing.T) {
	h := newTestAuthHandler(&fakeUserRepo{}, &fakeInvitationRepo{})
	if h.upsertUserFromIDToken(newGinCtx(), "not-a-jwt", "") {
		t.Fatal("malformed token must be rejected")
	}
}

// invitationToken が指定されているとき、email ベースより token ベースが優先されることを確認する。
// email ベースで見つかる古い invitation よりも、token ベースの新しい invitation を採用する。
func TestUpsertUserFromIDToken_InvitationToken_TakesPrecedenceOverEmail(t *testing.T) {
	cidByToken := uint64(99)
	cidByEmail := uint64(1)
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"u@example.com": {ID: 1, CompanyID: cidByEmail, Email: "u@example.com", Role: domain.RoleTrainee},
		},
		pendingByToken: map[string]*domain.AdminInvitation{
			"magic-token-xyz": {ID: 7, CompanyID: cidByToken, Email: "u@example.com", Role: domain.RoleCompanyAdmin, DisplayName: "佐藤"},
		},
	}
	h := newTestAuthHandler(users, invs)

	idToken := makeIDToken(t, map[string]any{
		"sub":   "google-user-1",
		"email": "u@example.com",
	})
	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "magic-token-xyz") {
		t.Fatal("must be allowed when token matches a pending invitation")
	}
	if users.created == nil {
		t.Fatalf("user must be created")
	}
	if users.created.Role != domain.RoleCompanyAdmin {
		t.Errorf("token-based invitation role should win, got %q", users.created.Role)
	}
	if users.created.CompanyID == nil || *users.created.CompanyID != cidByToken {
		t.Errorf("token-based companyID should win, got %+v", users.created.CompanyID)
	}
	if invs.updatedID != 7 || invs.updatedStatus != domain.InvitationStatusAccepted {
		t.Errorf("token-based invitation must be marked accepted, got id=%d status=%q", invs.updatedID, invs.updatedStatus)
	}
}

// invitationToken が無効でも、email ベースで見つかれば許可する（旧フロー互換）。
func TestUpsertUserFromIDToken_InvalidToken_FallsBackToEmail(t *testing.T) {
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"u@example.com": {ID: 1, CompanyID: 1, Email: "u@example.com", Role: domain.RoleTrainee},
		},
	}
	h := newTestAuthHandler(users, invs)

	idToken := makeIDToken(t, map[string]any{
		"sub":   "google-user-2",
		"email": "u@example.com",
	})
	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "garbage-token") {
		t.Fatal("must fall back to email-based invitation when token is invalid")
	}
	if users.created == nil || users.created.Role != domain.RoleTrainee {
		t.Errorf("expected trainee from email-based invitation, got %+v", users.created)
	}
}

// 既存の trainee ユーザーが company_admin 招待を受けた場合、role を昇格 + company を反映する。
// 過去に signup した既存ユーザーが後から CompanyAdmin として招待されたケースの救済。
func TestUpsertUserFromIDToken_ExistingTrainee_UpgradedByCompanyAdminInvitation(t *testing.T) {
	existing := &domain.User{ID: 5, CognitoSub: "existing-trainee", Email: "u@example.com", Role: domain.RoleTrainee}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"existing-trainee": existing}}
	invs := &fakeInvitationRepo{
		pendingByToken: map[string]*domain.AdminInvitation{
			"magic-xyz": {ID: 9, CompanyID: 42, Email: "u@example.com", Role: domain.RoleCompanyAdmin},
		},
	}
	h := newTestAuthHandler(users, invs)
	idToken := makeIDToken(t, map[string]any{"sub": "existing-trainee", "email": "u@example.com"})

	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "magic-xyz") {
		t.Fatal("must be allowed for existing user with token")
	}
	if users.updateRoleVal != domain.RoleCompanyAdmin || users.updateRoleID != 5 {
		t.Errorf("role must be upgraded to company_admin, got id=%d role=%q", users.updateRoleID, users.updateRoleVal)
	}
	if users.updateCompanyID != 5 || users.updateCompanyVal != 42 {
		t.Errorf("company_id must be updated to 42, got id=%d company=%d", users.updateCompanyID, users.updateCompanyVal)
	}
	if invs.updatedID != 9 || invs.updatedStatus != domain.InvitationStatusAccepted {
		t.Errorf("invitation must be marked accepted, got id=%d status=%q", invs.updatedID, invs.updatedStatus)
	}
}

// 既存 super_admin は招待を受けても降格しない。
func TestUpsertUserFromIDToken_ExistingSuperAdmin_NotDowngradedByInvitation(t *testing.T) {
	existing := &domain.User{ID: 1, CognitoSub: "ops", Email: "ops@example.com", Role: domain.RoleSuperAdmin}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"ops": existing}}
	invs := &fakeInvitationRepo{
		pendingByToken: map[string]*domain.AdminInvitation{
			"t": {ID: 1, CompanyID: 1, Email: "ops@example.com", Role: domain.RoleTrainee},
		},
	}
	h := newTestAuthHandler(users, invs)
	idToken := makeIDToken(t, map[string]any{"sub": "ops", "email": "ops@example.com"})

	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "t") {
		t.Fatal("must be allowed")
	}
	if users.updateRoleVal != "" {
		t.Errorf("super_admin must not be downgraded, but UpdateRole was called with %q", users.updateRoleVal)
	}
}

// dummy: 使用していないが strings import のリンタを満たすために残す（今後 assert で使う）
var _ = strings.Contains

// 新規ユーザ作成時に id_token の `name` claim が DisplayName に使われる（email にフォールバックしない）。
func TestUpsertUserFromIDToken_NewUser_UsesOIDCNameOverEmail(t *testing.T) {
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"taro@example.com": {
				ID: 1, CompanyID: 10, Email: "taro@example.com",
				Role: domain.RoleTrainee, Status: domain.InvitationStatusPending,
				// 招待 displayName が空だと OIDC name が採用される。
			},
		},
	}
	h := newTestAuthHandler(users, invs)
	idToken := makeIDToken(t, map[string]any{
		"sub":   "google-1",
		"email": "taro@example.com",
		"name":  "山田 太郎",
	})

	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.created == nil {
		t.Fatal("expected user created")
	}
	if users.created.DisplayName != "山田 太郎" {
		t.Errorf("DisplayName = %q, want 山田 太郎 (OIDC name)", users.created.DisplayName)
	}
}

// 招待 displayName が指定されているときは招待値が優先で OIDC name は無視。
func TestUpsertUserFromIDToken_NewUser_InvitationNameTrumpsOIDCName(t *testing.T) {
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"u@example.com": {
				ID: 1, CompanyID: 10, Email: "u@example.com",
				Role: domain.RoleTrainee, DisplayName: "招待された名前", Status: domain.InvitationStatusPending,
			},
		},
	}
	h := newTestAuthHandler(users, invs)
	idToken := makeIDToken(t, map[string]any{
		"sub":   "g-2", "email": "u@example.com", "name": "Google Name",
	})

	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.created.DisplayName != "招待された名前" {
		t.Errorf("DisplayName = %q, want 招待された名前", users.created.DisplayName)
	}
}

// name claim が無いケースは email にフォールバックする（後方互換）。
func TestUpsertUserFromIDToken_NewUser_NoOIDCName_FallsBackToEmail(t *testing.T) {
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"a@example.com": {
				ID: 1, CompanyID: 10, Email: "a@example.com",
				Role: domain.RoleTrainee, Status: domain.InvitationStatusPending,
			},
		},
	}
	h := newTestAuthHandler(users, invs)
	idToken := makeIDToken(t, map[string]any{"sub": "g-3", "email": "a@example.com"})

	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.created.DisplayName != "a@example.com" {
		t.Errorf("DisplayName = %q, want a@example.com (fallback)", users.created.DisplayName)
	}
}

// 既存ユーザの DisplayName が email と一致 + id_token に name → name で上書きされる。
func TestUpsertUserFromIDToken_ExistingUser_BackfillDisplayNameFromOIDC(t *testing.T) {
	existing := &domain.User{
		ID: 5, CognitoSub: "exists",
		Email: "old@example.com", DisplayName: "old@example.com",
		Role: domain.RoleTrainee,
	}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"exists": existing}}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	idToken := makeIDToken(t, map[string]any{
		"sub": "exists", "email": "old@example.com", "name": "本名 太郎",
	})

	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.updateDisplayNameID != 5 || users.updateDisplayNameVal != "本名 太郎" {
		t.Errorf("expected backfill UpdateDisplayName(5, '本名 太郎'), got id=%d val=%q",
			users.updateDisplayNameID, users.updateDisplayNameVal)
	}
}

// 既存ユーザが既にプロフィール編集済（DisplayName != email）なら OIDC name で上書きしない。
func TestUpsertUserFromIDToken_ExistingUser_NoBackfillIfDisplayNameCustomized(t *testing.T) {
	existing := &domain.User{
		ID: 5, CognitoSub: "exists",
		Email: "u@example.com", DisplayName: "ユーザ自身が編集した名前",
		Role: domain.RoleTrainee,
	}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"exists": existing}}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	idToken := makeIDToken(t, map[string]any{
		"sub": "exists", "email": "u@example.com", "name": "Google Name",
	})

	if !h.upsertUserFromIDToken(newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.updateDisplayNameVal != "" {
		t.Errorf("expected no backfill, but UpdateDisplayName called with %q", users.updateDisplayNameVal)
	}
}
