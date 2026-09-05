//go:build manual

package oidc

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

// 実際に動いている発行者が出したトークンを、この検証器が受け取れるかを確かめる。
//
// 単体テストは自分で作った鍵で署名しているので、「実装が想定している形」しか試せない。
// 発行者が実際に返す形（aud に何が入るか・azp が付くか・役割がどう入るか）は、
// 本物を通すまで分からない。推測で書いた部分が食い違っていると、
// 単体テストが全部緑のまま本番で全員 401 になる。
//
// 走らせ方（手元で使える OIDC 発行者を立てて認可フローを 1 回通してから）:
//
//	# 認可フローを 1 回通して token を JSON で保存する
//	OIDC_TOKENS_FILE=/path/to/tokens.json \
//	OIDC_ISSUER=<発行者の issuer URL> \
//	OIDC_JWKS_URI=<発行者の JWKS URL> \
//	OIDC_CLIENT_ID=<client id> \
//	go test -tags=manual ./internal/infra/oidc/ -run 実際の発行者 -v
//
// 手元の発行者が要るので build tag で普段は外してある（CI では走らない）。
func Test_実際の発行者が出したトークンを受け取れる(t *testing.T) {
	path := os.Getenv("OIDC_TOKENS_FILE")
	if path == "" {
		t.Skip("OIDC_TOKENS_FILE が無いので飛ばす")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("トークンを読めない: %v", err)
	}
	var tok struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(raw, &tok); err != nil {
		t.Fatalf("トークンを解釈できない: %v", err)
	}

	v, err := NewVerifier(Config{
		Issuer:   os.Getenv("OIDC_ISSUER"),
		JWKSURI:  os.Getenv("OIDC_JWKS_URI"),
		ClientID: os.Getenv("OIDC_CLIENT_ID"),
	})
	if err != nil {
		t.Fatalf("検証器を作れない: %v", err)
	}

	ctx := context.Background()

	claims, err := v.Verify(ctx, tok.AccessToken)
	if err != nil {
		t.Fatalf("実際の access_token が落ちた: %v", err)
	}
	t.Logf("access_token 受理: sub=%v aud=%v", claims["sub"], claims["aud"])

	idClaims, err := v.VerifyIDToken(ctx, tok.IDToken, "")
	if err != nil {
		t.Fatalf("実際の id_token が落ちた: %v", err)
	}
	t.Logf("id_token 受理: sub=%v email=%v azp=%v", idClaims["sub"], idClaims["email"], idClaims["azp"])

	// nonce を照合する経路も、実物で通しておく。
	nonce, _ := idClaims["nonce"].(string)
	if nonce == "" {
		t.Fatal("実際の id_token に nonce が入っていない（認可要求に nonce を載せていない）")
	}
	if _, err := v.VerifyIDToken(ctx, tok.IDToken, nonce); err != nil {
		t.Fatalf("正しい nonce なのに落ちた: %v", err)
	}
	if _, err := v.VerifyIDToken(ctx, tok.IDToken, nonce+"x"); err == nil {
		t.Fatal("違う nonce が通ってしまった")
	}

	if tok.RefreshToken == "" {
		t.Error("refresh_token が発行されていない（scope に offline_access が無い可能性）")
	}

	// 役割の読み取りも実物で確かめる。ここは形を推測しやすく、
	// 間違えると弾かれずに「権限が静かに消える」ので、本物で見ておく価値が高い。
	claimName := os.Getenv("OIDC_ROLES_CLAIM")
	if claimName == "" {
		claimName = "roles"
	}
	roles := RolesFromClaim(claims[claimName])
	t.Logf("access_token の役割: %v（クレーム名 %s）", roles, claimName)
	if want := os.Getenv("OIDC_EXPECT_ROLE"); want != "" && !HasRole(roles, want) {
		t.Fatalf("役割 %q を読めていない: %v（生の値: %#v）", want, roles, claims[claimName])
	}
}
