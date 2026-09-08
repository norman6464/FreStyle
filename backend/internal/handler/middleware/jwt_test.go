package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newAuthedEngine は verify を注入した JWTAuth 付きのルータを返す。
// ハンドラは context に積まれた値をそのまま返すので、middleware が何を渡したかを見られる。
func newAuthedEngine(verify VerifyFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(verify))
	r.GET("/x", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"sub": c.GetString(ContextKeySubject),
		})
	})
	return r
}

func getWithBearer(r *gin.Engine, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func Test_JWT認証_Authorizationヘッダが無ければ401(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		t.Fatal("ヘッダが無いのに検証が呼ばれた")
		return nil, nil
	})
	if got := getWithBearer(r, "").Code; got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", got)
	}
}

// Bearer 以外の形式（例: Basic 認証、値が空の Bearer）は弾く。
func Test_JWT認証_Bearer形式でなければ401(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		t.Fatal("Bearer 形式でないのに検証が呼ばれた")
		return nil, nil
	})
	gin.SetMode(gin.TestMode)
	for _, header := range []string{"Basic dXNlcjpwYXNz", "Bearer ", "bearer tok", "tok"} {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("Authorization", header)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("header = %q: status = %d, want 401", header, w.Code)
		}
	}
}

// 検証が落ちたら通さない。ここが素通りすると、署名検証の意味が無くなる。
func Test_JWT認証_検証が落ちたら401(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		return nil, errors.New("bad token")
	})
	if got := getWithBearer(r, "tok").Code; got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", got)
	}
}

// sub が無いトークンは通さない。誰なのかが決まらないまま先へ進むと、
// 後段が「空文字の利用者」として扱ってしまう。
func Test_JWT認証_subが無ければ401(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		return map[string]any{"email": "u@example.com"}, nil
	})
	w := getWithBearer(r, "tok")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "missing_sub") {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func Test_JWT認証_検証が通ればsubを渡す(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		return map[string]any{"sub": "abc-123"}, nil
	})
	w := getWithBearer(r, "tok")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"sub":"abc-123"`) {
		t.Fatalf("sub が渡っていない: %s", body)
	}
}
