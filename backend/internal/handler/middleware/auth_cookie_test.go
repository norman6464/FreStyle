package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// runHandler は与えた handler を 1 回呼び出してレスポンスを返す簡易ヘルパー。
func runHandler(handler gin.HandlerFunc) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	handler(c)
	return w
}

func TestSetAccessTokenCookie_DefaultsMaxAge(t *testing.T) {
	w := runHandler(func(c *gin.Context) {
		SetAccessTokenCookie(c, "AT", 0)
	})
	cookie := findCookie(t, w, CookieAccessToken)
	if cookie.Value != "AT" {
		t.Fatalf("value = %q", cookie.Value)
	}
	if cookie.MaxAge != AccessTokenDefaultMaxAgeSeconds {
		t.Fatalf("max-age = %d, want %d", cookie.MaxAge, AccessTokenDefaultMaxAgeSeconds)
	}
	if !cookie.HttpOnly {
		t.Fatal("HttpOnly should be true")
	}
	if !cookie.Secure {
		t.Fatal("Secure should be true")
	}
	if cookie.SameSite != http.SameSiteNoneMode {
		t.Fatalf("SameSite = %v", cookie.SameSite)
	}
}

func TestSetAccessTokenCookie_RespectsExplicitMaxAge(t *testing.T) {
	w := runHandler(func(c *gin.Context) {
		SetAccessTokenCookie(c, "AT", 7200)
	})
	cookie := findCookie(t, w, CookieAccessToken)
	if cookie.MaxAge != 7200 {
		t.Fatalf("max-age = %d, want 7200", cookie.MaxAge)
	}
}

func TestSetRefreshTokenCookie_SkipsEmpty(t *testing.T) {
	w := runHandler(func(c *gin.Context) {
		SetRefreshTokenCookie(c, "")
	})
	for _, h := range w.Header().Values("Set-Cookie") {
		if strings.HasPrefix(h, CookieRefreshToken+"=") {
			t.Fatalf("empty refresh_token should not produce a cookie, got %q", h)
		}
	}
}

func TestSetRefreshTokenCookie_LongMaxAge(t *testing.T) {
	w := runHandler(func(c *gin.Context) {
		SetRefreshTokenCookie(c, "RT")
	})
	cookie := findCookie(t, w, CookieRefreshToken)
	if cookie.Value != "RT" {
		t.Fatalf("value = %q", cookie.Value)
	}
	if cookie.MaxAge != RefreshTokenMaxAgeSeconds {
		t.Fatalf("max-age = %d, want %d", cookie.MaxAge, RefreshTokenMaxAgeSeconds)
	}
}

func TestClearAuthCookies_BothExpired(t *testing.T) {
	w := runHandler(ClearAuthCookies)
	for _, name := range []string{CookieAccessToken, CookieRefreshToken} {
		cookie := findCookie(t, w, name)
		if cookie.MaxAge != -1 {
			t.Fatalf("%s max-age = %d, want -1", name, cookie.MaxAge)
		}
		if cookie.Value != "" {
			t.Fatalf("%s value = %q, want empty", name, cookie.Value)
		}
	}
}

// findCookie は ResponseRecorder の Set-Cookie ヘッダーから指定名の Cookie を取り出す。
func findCookie(t *testing.T, w *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	resp := http.Response{Header: w.Header()}
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("cookie %q not found in %v", name, w.Header())
	return nil
}
