package middleware

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
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

func TestDecodeClaims_Success(t *testing.T) {
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

func TestDecodeClaims_InvalidFormat(t *testing.T) {
	cases := []string{"", "only.two", "a.b.c.d"}
	for _, c := range cases {
		if _, err := DecodeClaims(c); !errors.Is(err, ErrInvalidJWT) {
			t.Errorf("token=%q want ErrInvalidJWT, got %v", c, err)
		}
	}
}

func TestDecodeClaims_InvalidBase64(t *testing.T) {
	tok := "header.!!!notbase64!!!.sig"
	if _, err := DecodeClaims(tok); err == nil {
		t.Fatal("expected error")
	}
}

func TestDecodeClaims_InvalidJSON(t *testing.T) {
	header := encodeSegment(t, map[string]any{"alg": "RS256"})
	bogus := strings.NewReplacer("+", "-", "/", "_").Replace(
		strings.TrimRight(base64.StdEncoding.EncodeToString([]byte("not-json")), "="),
	)
	tok := header + "." + bogus + ".sig"
	if _, err := DecodeClaims(tok); err == nil {
		t.Fatal("expected error")
	}
}

func TestIsAdminFromGroups(t *testing.T) {
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

func TestToStringSliceFromClaim(t *testing.T) {
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
