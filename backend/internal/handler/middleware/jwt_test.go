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

const testRolesClaim = "urn:frestyle:test:roles"

// newAuthedEngine は verify を注入した JWTAuth 付きのルータを返す。
// ハンドラは context に積まれた値をそのまま返すので、middleware が何を渡したかを見られる。
func newAuthedEngine(verify VerifyFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(verify, testRolesClaim))
	r.GET("/x", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"sub":   c.GetString(ContextKeySubject),
			"roles": RolesFromContext(c),
		})
	})
	return r
}

func getWithCookie(r *gin.Engine, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: CookieAccessToken, Value: token})
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func Test_JWT認証_Cookieが無ければ401(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		t.Fatal("Cookie が無いのに検証が呼ばれた")
		return nil, nil
	})
	if got := getWithCookie(r, "").Code; got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", got)
	}
}

// 検証が落ちたら通さない。ここが素通りすると、署名検証の意味が無くなる。
func Test_JWT認証_検証が落ちたら401(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		return nil, errors.New("bad token")
	})
	if got := getWithCookie(r, "tok").Code; got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", got)
	}
}

// sub が無いトークンは通さない。誰なのかが決まらないまま先へ進むと、
// 後段が「空文字の利用者」として扱ってしまう。
func Test_JWT認証_subが無ければ401(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		return map[string]any{"email": "u@example.com"}, nil
	})
	w := getWithCookie(r, "tok")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "missing_sub") {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func Test_JWT認証_検証が通ればsubと役割を渡す(t *testing.T) {
	r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
		return map[string]any{
			"sub":          "abc-123",
			testRolesClaim: map[string]any{"admin": map[string]any{"org": "acme"}},
		}, nil
	})
	w := getWithCookie(r, "tok")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"sub":"abc-123"`) {
		t.Fatalf("sub が渡っていない: %s", body)
	}
	// 役割が表（名前を鍵にした map）で来ても読めること。配列だと決めつけていると
	// ここで空になり、権限が黙って消える。
	if !strings.Contains(body, `"roles":["admin"]`) {
		t.Fatalf("役割が渡っていない: %s", body)
	}
}

// 役割は発行者ごとに形が違う。配列だと決めつけると、表で来た瞬間に空になり、
// 弾かれるのではなく静かに権限が消える。どの形でも読めることを固定する。
func Test_JWT認証_役割はどの形でも読める(t *testing.T) {
	for _, c := range []struct {
		name string
		raw  any
		want string
	}{
		{"文字列の配列", []any{"admin", "editor"}, `"roles":["admin","editor"]`},
		{"文字列ひとつ", "admin", `"roles":["admin"]`},
		{"役割名を鍵にした表", map[string]any{"admin": map[string]any{"org": "acme"}}, `"roles":["admin"]`},
		{"空の配列", []any{}, `"roles":[]`},
		{"想定外の型", 42, `"roles":null`},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newAuthedEngine(func(context.Context, string) (map[string]any, error) {
				return map[string]any{"sub": "s", testRolesClaim: c.raw}, nil
			})
			body := getWithCookie(r, "tok").Body.String()
			if !strings.Contains(body, c.want) {
				t.Fatalf("body = %s, want %s", body, c.want)
			}
		})
	}
}

// 役割のクレーム名を空にしたら、役割は積まれない（設定で切れることの確認）。
func Test_JWT認証_役割クレーム名が空なら役割を積まない(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(func(context.Context, string) (map[string]any, error) {
		return map[string]any{"sub": "s", testRolesClaim: []any{"admin"}}, nil
	}, ""))
	r.GET("/x", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"roles": RolesFromContext(c)})
	})
	w := getWithCookie(r, "tok")
	if !strings.Contains(w.Body.String(), `"roles":null`) {
		t.Fatalf("役割が積まれてしまった: %s", w.Body.String())
	}
}
