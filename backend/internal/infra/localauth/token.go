package localauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// mintKey はローカル開発専用の署名鍵。検証側（router.buildJWTVerify の local 経路）は
// 署名を見ないため秘匿性は不要だが、JWT の形（header.payload.signature）を正しく保つために使う。
var mintKey = []byte("frestyle-localauth-dev-only")

// mintJWT は claims から HS256 署名付き JWT を組み立てる。
func mintJWT(claims map[string]any) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	h, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	p, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding
	signingInput := enc.EncodeToString(h) + "." + enc.EncodeToString(p)
	mac := hmac.New(sha256.New, mintKey)
	mac.Write([]byte(signingInput))
	return signingInput + "." + enc.EncodeToString(mac.Sum(nil)), nil
}

// Issuer は localauth が発行するトークンの iss クレーム値。
const Issuer = "frestyle-localauth"

// ErrNotLocalToken は localauth 発行のトークンではない（他の検証経路に回すべき）ことを表す。
var ErrNotLocalToken = errors.New("localauth: not a local token")

// VerifyToken は localauth が発行したトークンを検証して claims を返す。
// HMAC 署名・iss・exp を確認する。localauth 発行でないトークン（署名不一致・iss 違い）は
// ErrNotLocalToken を返し、呼び出し側が JWKS など通常の検証経路へフォールバックする。
func VerifyToken(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrNotLocalToken
	}
	enc := base64.RawURLEncoding
	mac := hmac.New(sha256.New, mintKey)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	want := enc.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(want), []byte(parts[2])) {
		return nil, ErrNotLocalToken
	}
	payload, err := enc.DecodeString(parts[1])
	if err != nil {
		return nil, ErrNotLocalToken
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrNotLocalToken
	}
	if iss, _ := claims["iss"].(string); iss != Issuer {
		return nil, ErrNotLocalToken
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("localauth: exp クレームがありません")
	}
	if time.Now().Unix() >= int64(exp) {
		return nil, fmt.Errorf("localauth: トークンの有効期限が切れています（再ログインしてください）")
	}
	return claims, nil
}
