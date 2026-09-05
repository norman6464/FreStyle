package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/infra/oidc"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// newRefreshHandler は token endpoint を模したサーバに向けた AuthHandler を返す。
// respond が token 応答の中身を決める。
func newRefreshHandler(t *testing.T, idp *testIdP, users *fakeUserRepo, respond func() map[string]any) (*AuthHandler, *int) {
	t.Helper()
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(respond())
	}))
	t.Cleanup(srv.Close)

	cfg := &config.OIDCConfig{
		ClientID:       testClientID,
		TokenURI:       srv.URL,
		AdminRoleClaim: testRolesClaim,
		AdminRole:      "admin",
	}
	h := &AuthHandler{
		oidcCfg:  cfg,
		verifier: idp.verifier(t),
		tokens: oidc.NewTokenExchanger(oidc.ExchangerConfig{
			ClientID: cfg.ClientID, TokenURI: cfg.TokenURI,
		}),
		upsertUser: usecase.NewUpsertUserFromIDTokenUseCase(
			users, fakeOidcIdentityRepo{}, fakeTxManager{},
		),
	}
	return h, &calls
}

// doRefresh は refresh_token Cookie を積んで Refresh を呼び、応答を返す。
func doRefresh(h *AuthHandler, cookieValue string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/auth/refresh", nil)
	if cookieValue != "" {
		c.Request.AddCookie(&http.Cookie{Name: middleware.CookieRefreshToken, Value: cookieValue})
	}
	h.Refresh(c)
	return rec
}

// setCookieValue は Set-Cookie 群から name の値を取り出す（無ければ空文字と false）。
func setCookieValue(rec *httptest.ResponseRecorder, name string) (string, bool) {
	for _, raw := range rec.Result().Cookies() {
		if raw.Name == name {
			return raw.Value, true
		}
	}
	return "", false
}

// **この PR の要。**
// 発行者は更新のたびに refresh_token を入れ替える（回転）。書き戻さないと Cookie には
// 使用済みの値が残り、2 回目の更新で必ず失敗する。多くの実装は使い回しを窃取の兆候と
// みなしてトークン系列ごと失効させるので、全員がログイン画面へ飛ばされる。
func Test_リフレッシュ_回転した値をCookieに書き戻す(t *testing.T) {
	idp := newTestIdP(t)
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 1, Email: "u@example.com"},
	}}
	h, _ := newRefreshHandler(t, idp, users, func() map[string]any {
		return map[string]any{
			"access_token":  idp.sign(t, map[string]any{"sub": "u1"}),
			"refresh_token": "rotated-refresh-token",
			"expires_in":    3600,
		}
	})

	rec := doRefresh(h, "old-refresh-token")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	got, ok := setCookieValue(rec, middleware.CookieRefreshToken)
	if !ok {
		t.Fatal("refresh_token の Cookie が書き戻されていない（2 回目の更新で必ず失敗する）")
	}
	if got != "rotated-refresh-token" {
		t.Fatalf("refresh_token = %q, want rotated-refresh-token", got)
	}
}

// access_token は毎回書き直す。
func Test_リフレッシュ_アクセストークンを書き直す(t *testing.T) {
	idp := newTestIdP(t)
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 1, Email: "u@example.com"},
	}}
	at := idp.sign(t, map[string]any{"sub": "u1"})
	h, _ := newRefreshHandler(t, idp, users, func() map[string]any {
		return map[string]any{"access_token": at, "refresh_token": "rt2", "expires_in": 3600}
	})

	rec := doRefresh(h, "old")
	got, ok := setCookieValue(rec, middleware.CookieAccessToken)
	if !ok || got != at {
		t.Fatalf("access_token Cookie = %q (ok=%t)", got, ok)
	}
}

// 発行者が refresh_token を返さなかったときは、手元の値を消さない。
// 空文字で上書きすると、まだ使えるはずのトークンを自分で捨てることになる。
func Test_リフレッシュ_新しい値が無ければCookieを消さない(t *testing.T) {
	idp := newTestIdP(t)
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 1, Email: "u@example.com"},
	}}
	h, _ := newRefreshHandler(t, idp, users, func() map[string]any {
		return map[string]any{
			"access_token": idp.sign(t, map[string]any{"sub": "u1"}),
			"expires_in":   3600,
		}
	})

	rec := doRefresh(h, "old")
	if _, ok := setCookieValue(rec, middleware.CookieRefreshToken); ok {
		t.Fatal("refresh_token を空で上書きしてしまっている")
	}
}

func Test_リフレッシュ_Cookieが無ければ401(t *testing.T) {
	idp := newTestIdP(t)
	h, calls := newRefreshHandler(t, idp, &fakeUserRepo{}, func() map[string]any { return nil })

	rec := doRefresh(h, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if *calls != 0 {
		t.Errorf("Cookie が無いのに発行者を呼んだ（%d 回）", *calls)
	}
}

// newFailingRefreshHandler は、発行者が指定の状態コードを返す状況を作る。
func newFailingRefreshHandler(t *testing.T, idp *testIdP, status int, body string) *AuthHandler {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	cfg := &config.OIDCConfig{ClientID: testClientID, TokenURI: srv.URL, AdminRoleClaim: testRolesClaim, AdminRole: "admin"}
	return &AuthHandler{
		oidcCfg:  cfg,
		verifier: idp.verifier(t),
		tokens:   oidc.NewTokenExchanger(oidc.ExchangerConfig{ClientID: cfg.ClientID, TokenURI: cfg.TokenURI}),
	}
}

// clearsAuthCookies は、両方の認証 Cookie が **失効させられている** かを返す。
//
// 名前があるかだけを見てはいけない。通常の値で再発行する回帰でも通ってしまい、
// 「消したつもりで消えていない」を捕まえられない。
func clearsAuthCookies(rec *httptest.ResponseRecorder) bool {
	expired := map[string]bool{}
	for _, c := range rec.Result().Cookies() {
		if c.Name != middleware.CookieAccessToken && c.Name != middleware.CookieRefreshToken {
			continue
		}
		if c.MaxAge < 0 || (!c.Expires.IsZero() && c.Expires.Before(time.Now())) {
			expired[c.Name] = true
		}
	}
	return expired[middleware.CookieAccessToken] && expired[middleware.CookieRefreshToken]
}

// その refresh_token がもう使えないと分かったときは、Cookie を消してログインへ戻す。
// 残すと、無効な値で叩き続けることになる。
func Test_リフレッシュ_grantが無効ならCookieを消す(t *testing.T) {
	idp := newTestIdP(t)
	h := newFailingRefreshHandler(t, idp, http.StatusBadRequest, `{"error":"invalid_grant"}`)

	rec := doRefresh(h, "expired")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !clearsAuthCookies(rec) {
		t.Fatalf("Cookie の消去が出ていない: %v", rec.Result().Header.Values("Set-Cookie"))
	}
}

// **発行者の一時的な不調で全員をログアウトさせない。**
// 429 や 5xx は「今は無理」であって「もう使えない」ではない。ここで Cookie を消すと、
// 発行者の短い不調が全利用者の強制ログアウトに化ける。
func Test_リフレッシュ_一時的な失敗ではCookieを消さない(t *testing.T) {
	for _, c := range []struct {
		name   string
		status int
		body   string
	}{
		{"絞られた（429）", http.StatusTooManyRequests, `{"error":"slow_down"}`},
		{"発行者が落ちている（500）", http.StatusInternalServerError, `{"error":"server_error"}`},
		{"上流が詰まっている（503）", http.StatusServiceUnavailable, ``},
		// 400 / 401 でも、原因が設定なら利用者のトークンは生きている。
		// 状態コードだけで判断すると、設定を 1 つ間違えた瞬間に全員がログアウトする。
		{"要求の組み立てが悪い（400 invalid_request）", http.StatusBadRequest, `{"error":"invalid_request"}`},
		{"クライアント設定が悪い（401 invalid_client）", http.StatusUnauthorized, `{"error":"invalid_client"}`},
		{"本文が読めない（400）", http.StatusBadRequest, `<html>gateway error</html>`},
	} {
		t.Run(c.name, func(t *testing.T) {
			idp := newTestIdP(t)
			h := newFailingRefreshHandler(t, idp, c.status, c.body)

			rec := doRefresh(h, "still-valid")
			if rec.Code != http.StatusBadGateway {
				t.Fatalf("status = %d, want 502（後で再試行できる形）body=%s", rec.Code, rec.Body.String())
			}
			if clearsAuthCookies(rec) {
				t.Fatal("一時的な失敗なのに Cookie を消してしまっている")
			}
		})
	}
}
