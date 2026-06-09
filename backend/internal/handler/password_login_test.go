package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
)

// fakePasswordAuth は passwordAuthenticator のテスト用スタブ。
type fakePasswordAuth struct {
	token       *cognito.Token
	err         error
	gotEmail    string
	gotPassword string
}

func (f *fakePasswordAuth) Authenticate(_ context.Context, email, password string) (*cognito.Token, error) {
	f.gotEmail, f.gotPassword = email, password
	return f.token, f.err
}

func postLoginCtx(body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/cognito/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, rec
}

func TestLogin_Success_ExistingUser(t *testing.T) {
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"sub-1": {ID: 1, CognitoSub: "sub-1", Email: "u@example.com", Role: domain.RoleTrainee},
	}}
	idTok := makeIDToken(t, map[string]any{"sub": "sub-1", "email": "u@example.com"})
	pw := &fakePasswordAuth{token: &cognito.Token{AccessToken: "AT", IDToken: idTok, RefreshToken: "RT", ExpiresIn: 3600}}
	h := &AuthHandler{users: users, invitations: &fakeInvitationRepo{}, passwordAuth: pw}

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

func TestLogin_InvalidCredentials_401(t *testing.T) {
	h := &AuthHandler{
		users:        &fakeUserRepo{},
		invitations:  &fakeInvitationRepo{},
		passwordAuth: &fakePasswordAuth{err: cognito.ErrInvalidCredentials},
	}
	c, rec := postLoginCtx(`{"email":"u@example.com","password":"wrong"}`)
	h.Login(c)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_NewUserWithoutInvitation_403(t *testing.T) {
	idTok := makeIDToken(t, map[string]any{"sub": "new-sub", "email": "new@example.com"})
	h := &AuthHandler{
		users:        &fakeUserRepo{},       // 既存ユーザーなし
		invitations:  &fakeInvitationRepo{}, // pending 招待なし
		passwordAuth: &fakePasswordAuth{token: &cognito.Token{AccessToken: "AT", IDToken: idTok, RefreshToken: "RT"}},
	}
	c, rec := postLoginCtx(`{"email":"new@example.com","password":"secret123"}`)
	h.Login(c)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_BadRequest_MissingPassword_400(t *testing.T) {
	h := &AuthHandler{passwordAuth: &fakePasswordAuth{}}
	c, rec := postLoginCtx(`{"email":"u@example.com"}`)
	h.Login(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_NotConfigured_500(t *testing.T) {
	h := &AuthHandler{} // passwordAuth nil
	c, rec := postLoginCtx(`{"email":"u@example.com","password":"secret123"}`)
	h.Login(c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// 認証は成功したが upsert が DB 失敗で落ちたケースは 403 ではなく 500（招待拒否と切り分け）。
func TestLogin_UpsertInternalError_500(t *testing.T) {
	idTok := makeIDToken(t, map[string]any{"sub": "s1", "email": "u@example.com"})
	users := &fakeUserRepo{createErr: errors.New("db down")}
	inv := &fakeInvitationRepo{pendingByEmail: map[string]*domain.AdminInvitation{
		"u@example.com": {ID: 1, Role: domain.RoleTrainee, CompanyID: 1},
	}}
	h := &AuthHandler{
		users:        users,
		invitations:  inv,
		passwordAuth: &fakePasswordAuth{token: &cognito.Token{AccessToken: "AT", IDToken: idTok, RefreshToken: "RT"}},
	}
	c, rec := postLoginCtx(`{"email":"u@example.com","password":"secret123"}`)
	h.Login(c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d body=%s", rec.Code, rec.Body.String())
	}
}
