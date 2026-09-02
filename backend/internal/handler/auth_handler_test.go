package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeUserRepo は AuthHandler.upsertUserFromIDToken のテスト用 stub。
type fakeUserRepo struct {
	existingBySub      map[string]*domain.User
	findErr            error
	created            *domain.User
	createErr          error
	updateRoleID       uint64
	updateRoleVal      domain.RoleName
	updateRoleErr      error
	updateRoleCalls    int
	superAdmins        []domain.User
	updateWorkspaceID  uint64
	updateWorkspaceVal *string
	updateNameID       uint64
	updateNameVal      string
}

func (r *fakeUserRepo) FindByCognitoSub(_ context.Context, sub string) (*domain.User, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if u, ok := r.existingBySub[sub]; ok {
		return u, nil
	}
	return nil, nil
}

func (r *fakeUserRepo) EnsureOidcIdentity(_ context.Context, _ uint64, _, _ string) error {
	return nil
}

func (r *fakeUserRepo) FindActiveByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, nil
}

func (r *fakeUserRepo) CognitoSubjectByUserID(_ context.Context, _ uint64) (string, error) {
	return "", nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, _ uint64) (*domain.User, error) {
	return nil, nil
}

func (r *fakeUserRepo) ListByRole(_ context.Context, _ domain.RoleName) ([]domain.User, error) {
	return r.superAdmins, nil
}

func (r *fakeUserRepo) CreateWithOidcIdentity(_ context.Context, u *domain.User, _, _ string) error {
	if r.createErr != nil {
		return r.createErr
	}
	u.ID = 7
	r.created = u
	return nil
}

// CreateFirstSuperAdminWithOidcIdentity は本物の repository と同じく「渡された role が
// super_admin でなければ拒否し、super_admin が 0 人のときだけ作る」振る舞いを模す。
func (r *fakeUserRepo) CreateFirstSuperAdminWithOidcIdentity(
	ctx context.Context, u *domain.User, provider, subject string,
) (bool, error) {
	// 本物（persistence.userRepository）はこの経路を super_admin 専用として先頭で弾く。
	// ダブルがこの事前条件を落とすと、bootstrap 経路に trainee が流れる壊れ方を検出できない。
	if u.Role != domain.RoleSuperAdmin {
		return false, fmt.Errorf("最初の運営管理者の作成に role %q が渡されました（super_admin 専用の経路です）", u.Role)
	}
	if len(r.superAdmins) > 0 {
		return false, nil
	}
	if err := r.CreateWithOidcIdentity(ctx, u, provider, subject); err != nil {
		return false, err
	}
	r.superAdmins = append(r.superAdmins, *u)
	return true, nil
}

func (r *fakeUserRepo) UpdateName(_ context.Context, id uint64, name string) error {
	r.updateNameID, r.updateNameVal = id, name
	return nil
}

func (r *fakeUserRepo) UpdateRole(_ context.Context, id uint64, role domain.RoleName) error {
	r.updateRoleCalls++
	if r.updateRoleErr != nil {
		return r.updateRoleErr
	}
	r.updateRoleID, r.updateRoleVal = id, role
	return nil
}

func (r *fakeUserRepo) UpdateWorkspaceID(_ context.Context, id uint64, workspaceID *string) error {
	r.updateWorkspaceID, r.updateWorkspaceVal = id, workspaceID
	return nil
}

func (r *fakeUserRepo) UpdateActive(context.Context, uint64, bool) error { return nil }
func (r *fakeUserRepo) SoftDelete(context.Context, uint64) error         { return nil }
func (r *fakeUserRepo) ListByWorkspaceID(_ context.Context, _ string) ([]domain.User, error) {
	return nil, nil
}

// fakeInvitationRepo は AdminInvitationRepository の最小スタブ。
// FindPendingByEmail / FindPendingByToken の振る舞いをカスタムにしてテストする。
type fakeInvitationRepo struct {
	pendingByEmail map[string]*domain.AdminInvitation
	pendingByToken map[string]*domain.AdminInvitation
	updatedID      uint64
	updatedStatus  string
}

type fakeUserInvitationTransactionRunner struct {
	users       repository.UserWithOidcIdentityCreator
	invitations repository.InvitationStatusUpdater
}

func (r *fakeUserInvitationTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(
		users repository.UserWithOidcIdentityCreator,
		invitations repository.InvitationStatusUpdater,
	) error,
) error {
	return fn(r.users, r.invitations)
}

func (r *fakeInvitationRepo) ListAll(_ context.Context) ([]domain.AdminInvitation, error) {
	return nil, nil
}

func (r *fakeInvitationRepo) ListByWorkspaceID(_ context.Context, _ string) ([]domain.AdminInvitation, error) {
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

func (r *fakeInvitationRepo) FindByID(_ context.Context, _ uint64) (*domain.AdminInvitation, error) {
	return nil, nil
}

func (r *fakeInvitationRepo) Create(_ context.Context, _ *domain.AdminInvitation) error { return nil }

func (r *fakeInvitationRepo) UpdateStatus(_ context.Context, id uint64, status string) error {
	r.updatedID, r.updatedStatus = id, status
	return nil
}

// makeIDToken は claims を本物の鍵で署名した id_token にする。
// 署名を検証する側になったので、テストも実際に署名する（testIdP は oidc_testkit_test.go）。
func makeIDToken(t *testing.T, idp *testIdP, claims map[string]any) string {
	t.Helper()
	return idp.sign(t, claims)
}

// newTestAuthHandler はテスト用 AuthHandler を組み立てる。tokens は使わない。
// ブートストラップ免除は無効（本番の既定）。
func newTestAuthHandler(
	t *testing.T,
	idp *testIdP,
	users *fakeUserRepo,
	invitations *fakeInvitationRepo,
) *AuthHandler {
	return newTestAuthHandlerWithBootstrap(t, idp, users, invitations, "")
}

// newTestAuthHandlerWithBootstrap はブートストラップ用アドレスを設定したテスト用 AuthHandler を返す。
func newTestAuthHandlerWithBootstrap(
	t *testing.T,
	idp *testIdP,
	users *fakeUserRepo,
	invitations *fakeInvitationRepo,
	bootstrapEmail string,
) *AuthHandler {
	t.Helper()
	return &AuthHandler{
		verifier: idp.verifier(t),
		oidcCfg:  &config.OIDCConfig{AdminRoleClaim: testRolesClaim, AdminRole: "admin"},
		upsertUser: usecase.NewUpsertUserFromIDTokenUseCase(
			users,
			invitations,
			bootstrapEmail,
			&fakeUserInvitationTransactionRunner{
				users:       users,
				invitations: invitations,
			},
		),
		promoteAdmin: usecase.NewPromoteCognitoAdminRoleUseCase(users),
	}
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

func Test_IDトークンからユーザー登録_既存ユーザーは常に許可(t *testing.T) {
	existing := &domain.User{ID: 5, Email: "u@example.com", Role: domain.RoleTrainee}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"existing": existing}}
	invs := &fakeInvitationRepo{}
	idp := newTestIdP(t)
	h := newTestAuthHandler(t, idp, users, invs)

	idToken := makeIDToken(t, idp, map[string]any{
		"sub":   "existing",
		"email": "u@example.com",
	})

	allowed := upsertAllowed(h, newGinCtx(), idToken, "")
	if !allowed {
		t.Fatalf("existing user must always be allowed (no invitation re-check)")
	}
	if users.created != nil {
		t.Fatalf("existing user must not trigger Create")
	}
}

func Test_IDトークンからユーザー登録_既存ユーザーはCognito_adminで昇格(t *testing.T) {
	idp := newTestIdP(t)
	existing := &domain.User{ID: 5, Email: "u@example.com", Role: domain.RoleTrainee}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"existing": existing}}
	h := newTestAuthHandler(t, idp, users, &fakeInvitationRepo{})

	idToken := makeIDToken(t, idp, map[string]any{
		"sub":          "existing",
		"email":        "u@example.com",
		testRolesClaim: map[string]any{"admin": map[string]any{}},
	})
	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.updateRoleID != 5 || users.updateRoleVal != domain.RoleSuperAdmin {
		t.Fatalf("expected role promoted to super_admin, got id=%d role=%q", users.updateRoleID, users.updateRoleVal)
	}
}

func Test_IDトークンからユーザー登録_不正な形式のトークンを拒否(t *testing.T) {
	idp := newTestIdP(t)
	h := newTestAuthHandler(t, idp, &fakeUserRepo{}, &fakeInvitationRepo{})
	if upsertAllowed(h, newGinCtx(), "not-a-jwt", "") {
		t.Fatal("malformed token must be rejected")
	}
}

// invitationToken が指定されているとき、email ベースより token ベースが優先されることを確認する。
// email ベースで見つかる古い invitation よりも、token ベースの新しい invitation を採用する。
func Test_IDトークンからユーザー登録_招待tokenはメールより優先(t *testing.T) {
	idp := newTestIdP(t)
	widByToken := "0198a000-0000-7000-8000-000000000099"
	widByEmail := "0198a000-0000-7000-8000-000000000001"
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"u@example.com": {ID: 1, WorkspaceID: &widByEmail, Email: "u@example.com", Role: domain.RoleTrainee},
		},
		pendingByToken: map[string]*domain.AdminInvitation{
			"magic-token-xyz": {ID: 7, WorkspaceID: &widByToken, Email: "u@example.com", Role: domain.RoleCompanyAdmin, Name: "佐藤"},
		},
	}
	h := newTestAuthHandler(t, idp, users, invs)

	idToken := makeIDToken(t, idp, map[string]any{
		"sub":   "google-user-1",
		"email": "u@example.com",
	})
	if !upsertAllowed(h, newGinCtx(), idToken, "magic-token-xyz") {
		t.Fatal("must be allowed when token matches a pending invitation")
	}
	if users.created == nil {
		t.Fatalf("user must be created")
	}
	if users.created.Role != domain.RoleCompanyAdmin {
		t.Errorf("token-based invitation role should win, got %q", users.created.Role)
	}
	if users.created.WorkspaceID == nil || *users.created.WorkspaceID != widByToken {
		t.Errorf("token-based workspaceID should win, got %+v", users.created.WorkspaceID)
	}
	if invs.updatedID != 7 || invs.updatedStatus != domain.InvitationStatusAccepted {
		t.Errorf("token-based invitation must be marked accepted, got id=%d status=%q", invs.updatedID, invs.updatedStatus)
	}
}

// invitationToken が無効でも、email ベースで見つかれば許可する（旧フロー互換）。
func Test_IDトークンからユーザー登録_無効なtokenはメールにフォールバック(t *testing.T) {
	idp := newTestIdP(t)
	wid := "0198a000-0000-7000-8000-000000000001"
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"u@example.com": {ID: 1, WorkspaceID: &wid, Email: "u@example.com", Role: domain.RoleTrainee},
		},
	}
	h := newTestAuthHandler(t, idp, users, invs)

	idToken := makeIDToken(t, idp, map[string]any{
		"sub":   "google-user-2",
		"email": "u@example.com",
	})
	if !upsertAllowed(h, newGinCtx(), idToken, "garbage-token") {
		t.Fatal("must fall back to email-based invitation when token is invalid")
	}
	if users.created == nil || users.created.Role != domain.RoleTrainee {
		t.Errorf("expected trainee from email-based invitation, got %+v", users.created)
	}
}

// 既存 super_admin は招待を受けても降格しない。
func Test_IDトークンからユーザー登録_既存運営管理者は招待で降格しない(t *testing.T) {
	idp := newTestIdP(t)
	existing := &domain.User{ID: 1, Email: "ops@example.com", Role: domain.RoleSuperAdmin}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"ops": existing}}
	wid := "0198a000-0000-7000-8000-000000000001"
	invs := &fakeInvitationRepo{
		pendingByToken: map[string]*domain.AdminInvitation{
			"t": {ID: 1, WorkspaceID: &wid, Email: "ops@example.com", Role: domain.RoleTrainee},
		},
	}
	h := newTestAuthHandler(t, idp, users, invs)
	idToken := makeIDToken(t, idp, map[string]any{"sub": "ops", "email": "ops@example.com"})

	if !upsertAllowed(h, newGinCtx(), idToken, "t") {
		t.Fatal("must be allowed")
	}
	if users.updateRoleVal != "" {
		t.Errorf("super_admin must not be downgraded, but UpdateRole was called with %q", users.updateRoleVal)
	}
}

// 新規ユーザ作成時に id_token の `name` claim が Name に使われる（email にフォールバックしない）。
func Test_IDトークンからユーザー登録_新規はOIDC名をメールより優先(t *testing.T) {
	idp := newTestIdP(t)
	wid := "0198a000-0000-7000-8000-000000000010"
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"taro@example.com": {
				ID: 1, WorkspaceID: &wid, Email: "taro@example.com",
				Role: domain.RoleTrainee, Status: domain.InvitationStatusPending,
				// 招待 displayName が空だと OIDC name が採用される。
			},
		},
	}
	h := newTestAuthHandler(t, idp, users, invs)
	idToken := makeIDToken(t, idp, map[string]any{
		"sub":   "google-1",
		"email": "taro@example.com",
		"name":  "山田 太郎",
	})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.created == nil {
		t.Fatal("expected user created")
	}
	if users.created.Name != "山田 太郎" {
		t.Errorf("Name = %q, want 山田 太郎 (OIDC name)", users.created.Name)
	}
}

// 招待 displayName が指定されているときは招待値が優先で OIDC name は無視。
func Test_IDトークンからユーザー登録_新規は招待名がOIDC名に優先(t *testing.T) {
	idp := newTestIdP(t)
	wid := "0198a000-0000-7000-8000-000000000010"
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"u@example.com": {
				ID: 1, WorkspaceID: &wid, Email: "u@example.com",
				Role: domain.RoleTrainee, Name: "招待された名前", Status: domain.InvitationStatusPending,
			},
		},
	}
	h := newTestAuthHandler(t, idp, users, invs)
	idToken := makeIDToken(t, idp, map[string]any{
		"sub": "g-2", "email": "u@example.com", "name": "Google Name",
	})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.created.Name != "招待された名前" {
		t.Errorf("Name = %q, want 招待された名前", users.created.Name)
	}
}

// name claim が無いケースは email にフォールバックする（後方互換）。
func Test_IDトークンからユーザー登録_新規でOIDC名なしはメールにフォールバック(t *testing.T) {
	idp := newTestIdP(t)
	wid := "0198a000-0000-7000-8000-000000000010"
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"a@example.com": {
				ID: 1, WorkspaceID: &wid, Email: "a@example.com",
				Role: domain.RoleTrainee, Status: domain.InvitationStatusPending,
			},
		},
	}
	h := newTestAuthHandler(t, idp, users, invs)
	idToken := makeIDToken(t, idp, map[string]any{"sub": "g-3", "email": "a@example.com"})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.created.Name != "a@example.com" {
		t.Errorf("Name = %q, want a@example.com (fallback)", users.created.Name)
	}
}

// 既存ユーザの Name が email と一致 + id_token に name → name で上書きされる。
func Test_IDトークンからユーザー登録_既存ユーザーは表示名をOIDCから補完(t *testing.T) {
	idp := newTestIdP(t)
	existing := &domain.User{
		ID:    5,
		Email: "old@example.com", Name: "old@example.com",
		Role: domain.RoleTrainee,
	}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"exists": existing}}
	h := newTestAuthHandler(t, idp, users, &fakeInvitationRepo{})
	idToken := makeIDToken(t, idp, map[string]any{
		"sub": "exists", "email": "old@example.com", "name": "本名 太郎",
	})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.updateNameID != 5 || users.updateNameVal != "本名 太郎" {
		t.Errorf("expected backfill UpdateName(5, '本名 太郎'), got id=%d val=%q",
			users.updateNameID, users.updateNameVal)
	}
}

// 既存ユーザが既にプロフィール編集済（Name != email）なら OIDC name で上書きしない。
func Test_IDトークンからユーザー登録_表示名カスタム済みは補完しない(t *testing.T) {
	idp := newTestIdP(t)
	existing := &domain.User{
		ID:    5,
		Email: "u@example.com", Name: "ユーザ自身が編集した名前",
		Role: domain.RoleTrainee,
	}
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{"exists": existing}}
	h := newTestAuthHandler(t, idp, users, &fakeInvitationRepo{})
	idToken := makeIDToken(t, idp, map[string]any{
		"sub": "exists", "email": "u@example.com", "name": "Google Name",
	})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("must be allowed")
	}
	if users.updateNameVal != "" {
		t.Errorf("expected no backfill, but UpdateName called with %q", users.updateNameVal)
	}
}

func Test_IDトークンからユーザー登録_壊れたトークンを弾く(t *testing.T) {
	idp := newTestIdP(t)
	h := newTestAuthHandler(t, idp, &fakeUserRepo{}, &fakeInvitationRepo{})

	user, err := h.upsertUserFromIDToken(newGinCtx(), "invalid-id-token", "", "")

	if user != nil {
		t.Fatal("不正な id_token を許可してはいけない")
	}
	if !errors.Is(err, errIDTokenRejected) {
		t.Fatalf("id_token の拒否として返っていない: %v", err)
	}
}

// 署名が正しくても、別の発行者・別の宛先のトークンは通さない。
// 1 つの発行者が複数のアプリにトークンを出す構成では、ここを見ないと
// 隣のアプリのトークンがそのまま通る。
func Test_IDトークンからユーザー登録_宛先違いを弾く(t *testing.T) {
	idp := newTestIdP(t)
	h := newTestAuthHandler(t, idp, &fakeUserRepo{}, &fakeInvitationRepo{})

	idToken := idp.signExact(t, map[string]any{
		"iss":   testIssuer,
		"aud":   "someone-elses-client",
		"sub":   "u1",
		"email": "u@example.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	user, err := h.upsertUserFromIDToken(newGinCtx(), idToken, "", "")
	if user != nil {
		t.Fatal("別のアプリ宛のトークンを許可してはいけない")
	}
	if !errors.Is(err, errIDTokenRejected) {
		t.Fatalf("id_token の拒否として返っていない: %v", err)
	}
}

// nonce は「この応答が自分の始めた認可の応答か」を確かめる値。
// 一致しないものを通すと、攻撃者が自分の認可コードを他人に踏ませる筋道が残る。
func Test_IDトークンからユーザー登録_nonce不一致を弾く(t *testing.T) {
	idp := newTestIdP(t)
	h := newTestAuthHandler(t, idp, &fakeUserRepo{}, &fakeInvitationRepo{})

	idToken := makeIDToken(t, idp, map[string]any{
		"sub": "u1", "email": "u@example.com", "nonce": "attacker-nonce",
	})

	user, err := h.upsertUserFromIDToken(newGinCtx(), idToken, "victim-nonce", "")
	if user != nil {
		t.Fatal("nonce が違うトークンを許可してはいけない")
	}
	if !errors.Is(err, errIDTokenRejected) {
		t.Fatalf("id_token の拒否として返っていない: %v", err)
	}
}

// 期待した nonce と一致していれば通る（上のテストが「常に落ちる」だけでないことの裏取り）。
func Test_IDトークンからユーザー登録_nonce一致なら通る(t *testing.T) {
	idp := newTestIdP(t)
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 1, Email: "u@example.com", Role: domain.RoleTrainee},
	}}
	h := newTestAuthHandler(t, idp, users, &fakeInvitationRepo{})

	idToken := makeIDToken(t, idp, map[string]any{
		"sub": "u1", "email": "u@example.com", "nonce": "same-nonce",
	})

	user, err := h.upsertUserFromIDToken(newGinCtx(), idToken, "same-nonce", "")
	if err != nil {
		t.Fatalf("nonce が一致しているのに落ちた: %v", err)
	}
	if user == nil {
		t.Fatal("ユーザーが返っていない")
	}
}
