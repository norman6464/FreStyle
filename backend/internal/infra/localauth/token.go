package localauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// mintKey はローカル開発専用の署名鍵。VerifyToken がこの鍵で HMAC を再計算して照合するため、
// 固定値をソースに置くと鍵を知る誰でもトークンを偽造できてしまう。既定はプロセス起動ごとの
// ランダム値（バックエンド再起動でトークンが失効し再ログインが必要になるが、ローカル開発では
// 許容）。挙動を固定したい場合のみ LOCAL_AUTH_SIGNING_KEY で上書きできる。
var mintKey = initMintKey()

func initMintKey() []byte {
	if k := os.Getenv("LOCAL_AUTH_SIGNING_KEY"); k != "" {
		return []byte(k)
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		// エントロピー枯渇は現実的にほぼ起きない。起きた場合も localauth はローカル専用のため
		// panic で気づけるようにする（本番経路には存在しない）。
		panic(fmt.Sprintf("localauth: 署名鍵の生成に失敗しました: %v", err))
	}
	return key
}

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
