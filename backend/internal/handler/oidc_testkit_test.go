package handler

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/infra/oidc"
)

// テスト用の発行者。実際に鍵を作り、JWKS を配り、その鍵で署名する。
//
// 以前のテストは `alg: none` の署名なしトークンを作って渡していた。当時の本体が
// 署名を検証していなかったので通っていたが、その状態では「署名を検証している」
// ことをテストで確かめられない。検証を入れた以上、テストも本物の鍵で署名する。
const (
	testIssuer   = "https://issuer.test"
	testClientID = "test-client-id"
	// testRolesClaim は役割の一覧が入るクレーム名。発行者ごとに違うので設定で指す。
	testRolesClaim = "urn:zitadel:iam:org:project:roles"
)

type testIdP struct {
	key    *rsa.PrivateKey
	kid    string
	server *httptest.Server
}

// newTestIdP は鍵を作って JWKS を配るテスト用サーバを立てる。
func newTestIdP(t *testing.T) *testIdP {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("鍵を作れない: %v", err)
	}
	idp := &testIdP{key: key, kid: "test-kid"}
	idp.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]string{
				{"kid": idp.kid, "kty": "RSA", "use": "sig", "alg": "RS256", "n": n, "e": e},
			},
		})
	}))
	t.Cleanup(idp.server.Close)
	return idp
}

// verifier はこの発行者を信頼する検証器を返す。
func (i *testIdP) verifier(t *testing.T) *oidc.Verifier {
	t.Helper()
	v, err := oidc.NewVerifier(oidc.Config{
		Issuer:   testIssuer,
		JWKSURI:  i.server.URL,
		ClientID: testClientID,
	})
	if err != nil {
		t.Fatalf("検証器を作れない: %v", err)
	}
	return v
}

// sign は claims に既定の iss / aud / exp / iat を補って署名する。
// 呼び出し側が同じ鍵を渡せば、その値が優先される。
func (i *testIdP) sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	full := map[string]any{
		"iss": testIssuer,
		"aud": testClientID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Add(-time.Minute).Unix(),
	}
	for k, v := range claims {
		full[k] = v
	}
	return i.signExact(t, full)
}

// signExact は渡された claims をそのまま署名する（既定を補わない）。
func (i *testIdP) signExact(t *testing.T, claims map[string]any) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT", "kid": i.kid})
	if err != nil {
		t.Fatalf("ヘッダを作れない: %v", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("ペイロードを作れない: %v", err)
	}
	signingInput := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(payload)
	sum := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, i.key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatalf("署名できない: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}
