package cognito

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// testSigner は RSA 鍵と JWKS を提供するテスト用 IdP。
type testSigner struct {
	key *rsa.PrivateKey
	kid string
}

func newTestSigner(t *testing.T) *testSigner {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen key: %v", err)
	}
	return &testSigner{key: key, kid: "test-kid-1"}
}

// jwksJSON は signer の公開鍵を JWKS 形式で返す。
func (s *testSigner) jwksJSON() []byte {
	n := base64.RawURLEncoding.EncodeToString(s.key.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.key.E)).Bytes())
	doc := map[string]any{"keys": []map[string]any{
		{"kid": s.kid, "kty": "RSA", "alg": "RS256", "use": "sig", "n": n, "e": e},
	}}
	b, _ := json.Marshal(doc)
	return b
}

// sign は claims を RS256 で署名した JWT を返す。kid / alg は引数で差し替え可能。
func (s *testSigner) sign(t *testing.T, claims map[string]any, alg, kid string) string {
	t.Helper()
	header := map[string]any{"alg": alg, "typ": "JWT", "kid": kid}
	seg := func(v any) string {
		b, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	signingInput := seg(header) + "." + seg(claims)
	hashed := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, hashed[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

const testIssuer = "https://cognito-idp.ap-northeast-1.amazonaws.com/pool"

// newVerifier は JWKS を返す httptest サーバに紐づく Verifier を作る。
func newVerifier(t *testing.T, s *testSigner) *Verifier {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(s.jwksJSON())
	}))
	t.Cleanup(srv.Close)
	return NewVerifierWithClient(srv.URL, testIssuer, srv.Client())
}

func validClaims() map[string]any {
	return map[string]any{
		"sub":            "user-1",
		"iss":            testIssuer,
		"exp":            float64(time.Now().Add(time.Hour).Unix()),
		"cognito:groups": []any{"admin"},
	}
}

func TestVerify_Success(t *testing.T) {
	s := newTestSigner(t)
	v := newVerifier(t, s)
	tok := s.sign(t, validClaims(), "RS256", s.kid)
	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if claims["sub"] != "user-1" {
		t.Fatalf("sub mismatch: %v", claims["sub"])
	}
}

func TestVerify_TamperedSignature(t *testing.T) {
	s := newTestSigner(t)
	v := newVerifier(t, s)
	tok := s.sign(t, validClaims(), "RS256", s.kid)
	// 署名を別の鍵で作り直す（payload はそのまま）→ 検証は失敗するべき。
	other := newTestSigner(t)
	other.kid = s.kid
	forged := other.sign(t, validClaims(), "RS256", s.kid)
	if _, err := v.Verify(context.Background(), forged); !errors.Is(err, ErrJWTBadSignature) {
		t.Fatalf("expected ErrJWTBadSignature, got %v", err)
	}
	// 正規トークンは通る（前段の偽造でキャッシュが壊れていないこと）。
	if _, err := v.Verify(context.Background(), tok); err != nil {
		t.Fatalf("valid token should pass, got %v", err)
	}
}

func TestVerify_Expired(t *testing.T) {
	s := newTestSigner(t)
	v := newVerifier(t, s)
	c := validClaims()
	c["exp"] = float64(time.Now().Add(-2 * time.Hour).Unix())
	tok := s.sign(t, c, "RS256", s.kid)
	if _, err := v.Verify(context.Background(), tok); !errors.Is(err, ErrJWTExpired) {
		t.Fatalf("expected ErrJWTExpired, got %v", err)
	}
}

func TestVerify_BadIssuer(t *testing.T) {
	s := newTestSigner(t)
	v := newVerifier(t, s)
	c := validClaims()
	c["iss"] = "https://evil.example.com"
	tok := s.sign(t, c, "RS256", s.kid)
	if _, err := v.Verify(context.Background(), tok); !errors.Is(err, ErrJWTBadIssuer) {
		t.Fatalf("expected ErrJWTBadIssuer, got %v", err)
	}
}

func TestVerify_RejectNonRS256(t *testing.T) {
	s := newTestSigner(t)
	v := newVerifier(t, s)
	tok := s.sign(t, validClaims(), "none", s.kid)
	if _, err := v.Verify(context.Background(), tok); !errors.Is(err, ErrJWTBadAlg) {
		t.Fatalf("expected ErrJWTBadAlg, got %v", err)
	}
}

func TestVerify_UnknownKid(t *testing.T) {
	s := newTestSigner(t)
	v := newVerifier(t, s)
	tok := s.sign(t, validClaims(), "RS256", "some-other-kid")
	if _, err := v.Verify(context.Background(), tok); !errors.Is(err, ErrJWTUnknownKey) {
		t.Fatalf("expected ErrJWTUnknownKey, got %v", err)
	}
}

func TestVerify_Malformed(t *testing.T) {
	v := newVerifier(t, newTestSigner(t))
	if _, err := v.Verify(context.Background(), "not-a-jwt"); !errors.Is(err, ErrJWTMalformed) {
		t.Fatalf("expected ErrJWTMalformed, got %v", err)
	}
}

func TestVerify_DisabledWhenNoJWKS(t *testing.T) {
	v := NewVerifier("")
	if _, err := v.Verify(context.Background(), "a.b.c"); !errors.Is(err, ErrVerifierDisabled) {
		t.Fatalf("expected ErrVerifierDisabled, got %v", err)
	}
}

// TestVerify_EmptyJWKSDoesNotWipeCache は空 JWKS が来ても既存の有効キャッシュを保持し、
// 既知トークンが通り続けることを確認する。
func TestVerify_EmptyJWKSDoesNotWipeCache(t *testing.T) {
	s := newTestSigner(t)
	empty := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if empty {
			_, _ = w.Write([]byte(`{"keys":[]}`))
			return
		}
		_, _ = w.Write(s.jwksJSON())
	}))
	t.Cleanup(srv.Close)
	v := NewVerifierWithClient(srv.URL, testIssuer, srv.Client())
	v.refreshCooldown = 0 // 毎回 refresh を許可

	tok := s.sign(t, validClaims(), "RS256", s.kid)
	if _, err := v.Verify(context.Background(), tok); err != nil {
		t.Fatalf("first verify should pass: %v", err)
	}

	// JWKS が空を返す状態に切替。未知 kid を引くと refresh は ErrJWKSUnavailable で失敗するが、
	// 既存キャッシュは空に潰されない。
	empty = true
	if _, err := v.Verify(context.Background(), s.sign(t, validClaims(), "RS256", "unknown")); !errors.Is(err, ErrJWKSUnavailable) {
		t.Fatalf("unknown kid with empty jwks should be ErrJWKSUnavailable, got %v", err)
	}
	// 既存 kid のトークンは引き続き通る（キャッシュが空に潰されていない）。
	if _, err := v.Verify(context.Background(), tok); err != nil {
		t.Fatalf("known token should still pass after empty jwks: %v", err)
	}
}
