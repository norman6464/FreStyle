package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
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
	tests := []struct {
		name string
		in   any
		want []string // nil は「読めなかった」。長さ 0 の非 nil は「読めた上で空」。
	}{
		{"すべて string", []any{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"空配列は読めた上で空", []any{}, []string{}},
		{"非 string が混ざれば読めない", []any{"admin", 42}, nil},
		{"非 string だけでも読めない", []any{42}, nil},
		{"配列でなければ読めない", "not-an-array", nil},
		{"キーが無ければ読めない", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToStringSliceFromClaim(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v want %#v", got, tt.want)
			}
			if tt.want != nil && got == nil {
				t.Fatal("読めた配列を nil で返してはならない（欠落と区別が付かなくなる）")
			}
		})
	}
}

// 壊れた claim を「グループに居ない」と読むと、正当な運営管理者の権限を剥がしてしまう。
// 読めない形は Absent（何も判断しない）へ倒れることを固定する。
func Test_運営権限クレーム_読めない形は判断しない(t *testing.T) {
	tests := []struct {
		name   string
		claims map[string]any
		want   domain.PlatformAdminClaim
	}{
		{"キーが無い", map[string]any{}, domain.PlatformAdminClaimAbsent},
		{"配列でない", map[string]any{"cognito:groups": "admin"}, domain.PlatformAdminClaimAbsent},
		{"非 string だけ", map[string]any{"cognito:groups": []any{42}}, domain.PlatformAdminClaimAbsent},
		{"admin と非 string の混在", map[string]any{"cognito:groups": []any{"admin", 42}}, domain.PlatformAdminClaimAbsent},
		{"空配列は失効", map[string]any{"cognito:groups": []any{}}, domain.PlatformAdminClaimRevoked},
		{"admin を含まない", map[string]any{"cognito:groups": []any{"users"}}, domain.PlatformAdminClaimRevoked},
		{"admin を含む", map[string]any{"cognito:groups": []any{"users", "admin"}}, domain.PlatformAdminClaimGranted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PlatformAdminClaimFromClaims(tt.claims); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

// JWTAuth 経由でも同じ結論になること（context に置くか否かが claim の存在の印なので、
// 読めない claim を置いてしまうと失効判定が「グループに居ない」と誤読する）。
func Test_JWT認証_運営権限クレームの存在判定(t *testing.T) {
	tests := []struct {
		name   string
		groups any
		want   domain.PlatformAdminClaim
	}{
		{"キーが無い", nil, domain.PlatformAdminClaimAbsent},
		{"非 string だけ", []any{42}, domain.PlatformAdminClaimAbsent},
		{"admin と非 string の混在", []any{"admin", 42}, domain.PlatformAdminClaimAbsent},
		{"空配列", []any{}, domain.PlatformAdminClaimRevoked},
		{"admin を含む", []any{"admin"}, domain.PlatformAdminClaimGranted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			claims := map[string]any{"sub": "user-9"}
			if tt.groups != nil {
				claims["cognito:groups"] = tt.groups
			}
			verify := func(context.Context, string) (map[string]any, error) { return claims, nil }

			var got domain.PlatformAdminClaim
			r := gin.New()
			r.GET("/x", JWTAuth(verify), func(c *gin.Context) {
				got = PlatformAdminClaimFromContext(c)
				c.Status(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.AddCookie(&http.Cookie{Name: CookieAccessToken, Value: "good"})
			r.ServeHTTP(httptest.NewRecorder(), req)

			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
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
