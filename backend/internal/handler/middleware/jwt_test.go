package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// makeJWT は header.payload.signature 形式のダミー JWT を組み立てる。
// 署名検証はしないので signature 部はプレースホルダで良い。
func makeJWT(t *testing.T, payload map[string]any) string {
	t.Helper()
	header := encodeSegment(t, map[string]any{"alg": "RS256", "typ": "JWT"})
	body := encodeSegment(t, payload)
	return header + "." + body + ".sig"
}

func encodeSegment(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// base64URL (パディング省略) で encode
	s := base64.StdEncoding.EncodeToString(b)
	s = strings.TrimRight(s, "=")
	s = strings.NewReplacer("+", "-", "/", "_").Replace(s)
	return s
}

func Test_クレームデコード_成功(t *testing.T) {
	want := map[string]any{
		"sub":            "abc-123",
		"email":          "u@example.com",
		"cognito:groups": []any{"admin"},
	}
	tok := makeJWT(t, want)
	got, err := DecodeClaims(tok)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["sub"] != "abc-123" || got["email"] != "u@example.com" {
		t.Fatalf("unexpected claims: %+v", got)
	}
	if groups := ToStringSliceFromClaim(got["cognito:groups"]); len(groups) != 1 || groups[0] != "admin" {
		t.Fatalf("groups: %+v", groups)
	}
}

func Test_クレームデコード_不正な形式(t *testing.T) {
	cases := []string{"", "only.two", "a.b.c.d"}
	for _, c := range cases {
		if _, err := DecodeClaims(c); !errors.Is(err, ErrInvalidJWT) {
			t.Errorf("token=%q want ErrInvalidJWT, got %v", c, err)
		}
	}
}

func Test_クレームデコード_不正なBase64(t *testing.T) {
	tok := "header.!!!notbase64!!!.sig"
	if _, err := DecodeClaims(tok); err == nil {
		t.Fatal("expected error")
	}
}

func Test_クレームデコード_不正なJSON(t *testing.T) {
	header := encodeSegment(t, map[string]any{"alg": "RS256"})
	bogus := strings.NewReplacer("+", "-", "/", "_").Replace(
		strings.TrimRight(base64.StdEncoding.EncodeToString([]byte("not-json")), "="),
	)
	tok := header + "." + bogus + ".sig"
	if _, err := DecodeClaims(tok); err == nil {
		t.Fatal("expected error")
	}
}

func Test_グループからadmin判定(t *testing.T) {
	if !IsAdminFromGroups([]string{"trainee", "admin"}) {
		t.Fatal("admin should be detected")
	}
	if IsAdminFromGroups([]string{"trainee"}) {
		t.Fatal("non-admin should not be detected")
	}
	if IsAdminFromGroups(nil) {
		t.Fatal("nil should not be detected")
	}
}

func Test_クレームから文字列スライス変換(t *testing.T) {
	got := ToStringSliceFromClaim([]any{"a", "b", 42, "c"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %v want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	if ToStringSliceFromClaim("not-an-array") != nil {
		t.Fatal("non-array should return nil")
	}
}

// runJWTAuth は JWTAuth を 1 リクエスト分実行し、status と context にセットされた sub を返す。
func runJWTAuth(t *testing.T, verify VerifyFunc, cookie string) (int, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var gotSub string
	r.GET("/x", JWTAuth(verify), func(c *gin.Context) {
		if v, ok := c.Get(ContextKeyCognitoSub); ok {
			gotSub, _ = v.(string)
		}
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: CookieAccessToken, Value: cookie})
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code, gotSub
}

func Test_JWT認証_Cookieなし(t *testing.T) {
	verify := func(context.Context, string) (map[string]any, error) { return nil, nil }
	if code, _ := runJWTAuth(t, verify, ""); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func Test_JWT認証_検証失敗(t *testing.T) {
	// 偽造トークン相当: verify がエラーを返したら 401 で弾く。
	verify := func(context.Context, string) (map[string]any, error) {
		return nil, errors.New("bad signature")
	}
	if code, _ := runJWTAuth(t, verify, "forged"); code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on verify failure, got %d", code)
	}
}

func Test_JWT認証_検証成功(t *testing.T) {
	verify := func(context.Context, string) (map[string]any, error) {
		return map[string]any{"sub": "user-9"}, nil
	}
	code, sub := runJWTAuth(t, verify, "good")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if sub != "user-9" {
		t.Fatalf("expected sub user-9, got %q", sub)
	}
}

func Test_JWT認証_sub欠落(t *testing.T) {
	verify := func(context.Context, string) (map[string]any, error) {
		return map[string]any{"email": "u@example.com"}, nil
	}
	if code, _ := runJWTAuth(t, verify, "good"); code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when sub missing, got %d", code)
	}
}
