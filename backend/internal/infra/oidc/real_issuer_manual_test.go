//go:build manual

package oidc

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// 実際に動いている発行者が出したトークンを、この検証器が受け取れるかを確かめる。
//
// 単体テストは自分で作った鍵で署名しているので、「実装が想定している形」しか試せない。
// 発行者が実際に返す形（aud に何が入るか・azp が付くか・役割がどう入るか）は、
// 本物を通すまで分からない。推測で書いた部分が食い違っていると、
// 単体テストが全部緑のまま本番で全員 401 になる。
//
// 走らせ方（GCIP なら Firebase JS SDK でクライアント側サインインし、得られた ID トークンを
// JSON で保存してから。access_token は GCIP には無いので id_token と同じ値を入れてよい）:
//
//	# ブラウザコンソール等でサインインし、user.getIdToken() の値を JSON で保存する
//	OIDC_TOKENS_FILE=/path/to/tokens.json \
//	OIDC_ISSUER=<発行者の issuer URL> \
//	OIDC_JWKS_URI=<発行者の JWKS URL> \
//	OIDC_CLIENT_ID=<client id。GCIP なら空でよく、代わりに検証器の Audiences にプロジェクト ID を渡す> \
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

	var audiences []string
	if a := os.Getenv("OIDC_AUDIENCES"); a != "" {
		audiences = strings.Split(a, ",")
	}
	v, err := NewVerifier(Config{
		Issuer:    os.Getenv("OIDC_ISSUER"),
		JWKSURI:   os.Getenv("OIDC_JWKS_URI"),
		ClientID:  os.Getenv("OIDC_CLIENT_ID"),
		Audiences: audiences,
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

	// nonce は空で照合する。GCIP はクライアント SDK が直接発行者とやり取りするため、
	// リダイレクトを介した認可要求（nonce で守る対象）自体が存在せず、
	// 実際の id_token にも nonce クレームは入らない。
	idClaims, err := v.VerifyIDToken(ctx, tok.IDToken, "")
	if err != nil {
		t.Fatalf("実際の id_token が落ちた: %v", err)
	}
	t.Logf("id_token 受理: sub=%v email=%v azp=%v", idClaims["sub"], idClaims["email"], idClaims["azp"])

	// GCIP は refresh_token をこの形では公開しない（クライアント SDK が内部で保持し、
	// getIdToken() の再呼び出しで自動更新する）。入っていなくても異常ではないので情報として残すだけ。
	if tok.RefreshToken == "" {
		t.Log("refresh_token は空（GCIP はこの形では公開しないため想定内）")
	}
}
