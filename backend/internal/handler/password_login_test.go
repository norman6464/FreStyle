package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
)

// fakePasswordAuth は passwordAuthenticator のテスト用スタブ。
type fakePasswordAuth struct {
	token       *cognito.Token
	err         error
	gotEmail    string
	gotPassword string

	// 新パスワード応答用（NewPassword ハンドラのテストで使う）。
	respondToken   *cognito.Token
	respondErr     error
	gotSession     string
	gotNewPassword string
}

func (f *fakePasswordAuth) Authenticate(_ context.Context, email, password string) (*cognito.Token, error) {
	f.gotEmail, f.gotPassword = email, password
	return f.token, f.err
}

func (f *fakePasswordAuth) RespondToNewPassword(_ context.Context, email, session, newPassword string) (*cognito.Token, error) {
	f.gotEmail, f.gotSession, f.gotNewPassword = email, session, newPassword
	return f.respondToken, f.respondErr
}

func postLoginCtx(body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/cognito/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, rec
}

func assertAuthCookiesUnchanged(
	t *testing.T,
	rec *httptest.ResponseRecorder,
) {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == middleware.CookieAccessToken ||
			cookie.Name == middleware.CookieRefreshToken {
			t.Fatalf(
				"authentication cookie %q must not be changed on invitation denial",
				cookie.Name,
			)
		}
	}
}

func Test_ログイン_成功_既存ユーザー(t *testing.T) {
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"sub-1": {ID: 1, Email: "u@example.com", Role: domain.RoleTrainee},
	}}
	idTok := makeIDToken(t, map[string]any{"sub": "sub-1", "email": "u@example.com"})
	pw := &fakePasswordAuth{token: &cognito.Token{AccessToken: "AT", IDToken: idTok, RefreshToken: "RT", ExpiresIn: 3600}}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	h.passwordAuth = pw

	c, rec := postLoginCtx(`{"email":"u@example.com","password":"secret123"}`)
	h.Login(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if pw.gotEmail != "u@example.com" || pw.gotPassword != "secret123" {
		t.Errorf("authenticator did not receive credentials: %q / %q", pw.gotEmail, pw.gotPassword)
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Errorf("expected auth cookies to be set on success")
	}
}

func Test_ログイン_認証情報不正_401(t *testing.T) {
	h := newTestAuthHandler(
		&fakeUserRepo{},
		&fakeInvitationRepo{},
	)
	h.passwordAuth = &fakePasswordAuth{
		err: cognito.ErrInvalidCredentials,
	}
	c, rec := postLoginCtx(`{"email":"u@example.com","password":"wrong"}`)
	h.Login(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func Test_ログイン_招待なし新規ユーザーも自己サインアップできる(t *testing.T) {
	users := &fakeUserRepo{}
	idTok := makeIDToken(t, map[string]any{"sub": "new-sub", "email": "new@example.com"})
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	h.passwordAuth = &fakePasswordAuth{
		token: &cognito.Token{
			AccessToken:  "AT",
			IDToken:      idTok,
			RefreshToken: "RT",
		},
	}
	router := gin.New()
	router.POST("/api/v2/auth/cognito/login", h.Login)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v2/auth/cognito/login",
		strings.NewReader(`{"email":"new@example.com","password":"secret123"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if users.created == nil {
		t.Fatal("user was not created")
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Error("expected auth cookies to be set on success")
	}
}

func Test_コールバック_招待なし新規ユーザーも自己サインアップできる(t *testing.T) {
	idToken := makeIDToken(t, map[string]any{
		"sub":   "new-callback-sub",
		"email": "new-callback@example.com",
	})
	tokenServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(cognito.Token{
				AccessToken:  "new-access-token",
				IDToken:      idToken,
				RefreshToken: "new-refresh-token",
				ExpiresIn:    3600,
				TokenType:    "Bearer",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	))
	defer tokenServer.Close()

	users := &fakeUserRepo{}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	h.tokens = cognito.NewTokenExchangerWithClient(
		cognito.Config{
			ClientID:    "test-client",
			RedirectURI: "https://example.com/callback",
			TokenURI:    tokenServer.URL,
		},
		tokenServer.Client(),
	)

	router := gin.New()
	router.POST("/api/v2/auth/login", h.Callback)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v2/auth/login",
		strings.NewReader(`{"code":"authorization-code"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if users.created == nil {
		t.Fatal("user was not created")
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Error("expected auth cookies to be set on success")
	}
}

func Test_ログイン_パスワード欠落_400(t *testing.T) {
	h := &AuthHandler{passwordAuth: &fakePasswordAuth{}}
	c, rec := postLoginCtx(`{"email":"u@example.com"}`)
	h.Login(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func Test_ログイン_未設定_500(t *testing.T) {
	h := &AuthHandler{} // passwordAuth nil
	c, rec := postLoginCtx(`{"email":"u@example.com","password":"secret123"}`)
	h.Login(c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// 認証は成功したが upsert が DB 失敗で落ちたケースは 403 ではなく 500（招待拒否と切り分け）。
func Test_ログイン_upsert内部エラー_500(t *testing.T) {
	idTok := makeIDToken(t, map[string]any{"sub": "s1", "email": "u@example.com"})
	users := &fakeUserRepo{createErr: errors.New("db down")}
	inv := &fakeInvitationRepo{pendingByEmail: map[string]*domain.AdminInvitation{
		"u@example.com": {ID: 1, Role: domain.RoleTrainee, CompanyID: 1},
	}}
	h := newTestAuthHandler(users, inv)
	h.passwordAuth = &fakePasswordAuth{
		token: &cognito.Token{
			AccessToken:  "AT",
			IDToken:      idTok,
			RefreshToken: "RT",
		},
	}
	c, rec := postLoginCtx(`{"email":"u@example.com","password":"secret123"}`)
	h.Login(c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func postCtx(path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, rec
}

func Test_ログイン_NEW_PASSWORD_REQUIRED_はチャレンジを返す(t *testing.T) {
	h := newTestAuthHandler(&fakeUserRepo{}, &fakeInvitationRepo{})
	h.passwordAuth = &fakePasswordAuth{err: &cognito.NewPasswordRequiredError{Session: "sess-xyz"}}

	c, rec := postLoginCtx(`{"email":"u@example.com","password":"temp"}`)
	h.Login(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "NEW_PASSWORD_REQUIRED") ||
		!strings.Contains(rec.Body.String(), "sess-xyz") {
		t.Errorf("challenge/session not in body: %s", rec.Body.String())
	}
	// チャレンジ段階では Cookie を発行しない。
	assertAuthCookiesUnchanged(t, rec)
}

func Test_新パスワード設定_成功でCookie発行(t *testing.T) {
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"sub-np": {ID: 5, Email: "np@example.com", Role: domain.RoleTrainee},
	}}
	idTok := makeIDToken(t, map[string]any{"sub": "sub-np", "email": "np@example.com"})
	pw := &fakePasswordAuth{respondToken: &cognito.Token{AccessToken: "AT", IDToken: idTok, RefreshToken: "RT", ExpiresIn: 3600}}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	h.passwordAuth = pw

	c, rec := postCtx("/api/v2/auth/cognito/new-password",
		`{"email":"np@example.com","session":"sess-xyz","newPassword":"New-Pass-1"}`)
	h.NewPassword(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if pw.gotSession != "sess-xyz" || pw.gotNewPassword != "New-Pass-1" {
		t.Errorf("authenticator did not receive challenge response: session=%q newpw=%q", pw.gotSession, pw.gotNewPassword)
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Errorf("expected auth cookies on success")
	}
}

func Test_新パスワード設定_session失効は401(t *testing.T) {
	h := newTestAuthHandler(&fakeUserRepo{}, &fakeInvitationRepo{})
	h.passwordAuth = &fakePasswordAuth{respondErr: cognito.ErrInvalidCredentials}

	c, rec := postCtx("/api/v2/auth/cognito/new-password",
		`{"email":"np@example.com","session":"expired","newPassword":"New-Pass-1"}`)
	h.NewPassword(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func Test_新パスワード設定_ポリシー違反は400(t *testing.T) {
	h := newTestAuthHandler(&fakeUserRepo{}, &fakeInvitationRepo{})
	h.passwordAuth = &fakePasswordAuth{respondErr: errors.New("InvalidPasswordException")}

	c, rec := postCtx("/api/v2/auth/cognito/new-password",
		`{"email":"np@example.com","session":"sess","newPassword":"weak"}`)
	h.NewPassword(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
