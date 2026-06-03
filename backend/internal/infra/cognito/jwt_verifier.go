package cognito

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Verifier は Cognito が発行した JWT (access_token / id_token) の署名と
// 標準クレームを検証する。JWKS を取得して kid ごとの RSA 公開鍵で RS256 署名を検証し、
// exp（有効期限）と iss（発行者）も確認する。
//
// 署名未検証の base64 デコードだけでは sub / cognito:groups を偽造できてしまうため、
// 保護ルートの認可前に必ずこの Verify を通す。
type Verifier struct {
	jwksURI    string
	issuer     string
	httpClient *http.Client

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	// refreshMu は JWKS 再取得を 1 本に直列化し、未知 kid 同時多発時のスパイクを防ぐ。
	refreshMu sync.Mutex
	// refreshCooldown は未知 kid によるリフェッチ連打 (DoS) を防ぐ最小間隔。
	refreshCooldown time.Duration
	// leeway は時計ずれを吸収する exp の許容誤差。
	leeway time.Duration
}

// 検証失敗の sentinel エラー。呼び出し側は errors.Is で分岐できる。
var (
	ErrJWTMalformed     = errors.New("cognito: malformed jwt")
	ErrJWTBadAlg        = errors.New("cognito: unexpected signing alg")
	ErrJWTUnknownKey    = errors.New("cognito: signing key not found")
	ErrJWTBadSignature  = errors.New("cognito: signature verification failed")
	ErrJWTExpired       = errors.New("cognito: token expired")
	ErrJWTBadIssuer     = errors.New("cognito: unexpected issuer")
	ErrJWKSUnavailable  = errors.New("cognito: jwks fetch failed")
	ErrVerifierDisabled = errors.New("cognito: verifier not configured (empty jwks uri)")
)

// NewVerifier は JWKS URI から Verifier を組み立てる。
// issuer は JWKS URI から `/.well-known/jwks.json` を除いて導出する
// (Cognito の iss = https://cognito-idp.<region>.amazonaws.com/<userPoolId>)。
func NewVerifier(jwksURI string) *Verifier {
	issuer := strings.TrimSuffix(jwksURI, "/.well-known/jwks.json")
	return &Verifier{
		jwksURI:         jwksURI,
		issuer:          issuer,
		httpClient:      &http.Client{Timeout: 5 * time.Second},
		keys:            map[string]*rsa.PublicKey{},
		refreshCooldown: 1 * time.Minute,
		leeway:          60 * time.Second,
	}
}

// NewVerifierWithClient はテスト用に http.Client / issuer を差し替えるコンストラクタ。
func NewVerifierWithClient(jwksURI, issuer string, client *http.Client) *Verifier {
	v := NewVerifier(jwksURI)
	v.issuer = issuer
	if client != nil {
		v.httpClient = client
	}
	return v
}

// Verify は token の署名・exp・iss を検証し、検証済みの claims を返す。
func (v *Verifier) Verify(ctx context.Context, token string) (map[string]any, error) {
	if v.jwksURI == "" {
		return nil, ErrVerifierDisabled
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrJWTMalformed
	}

	header, err := decodeJSONSegment(parts[0])
	if err != nil {
		return nil, ErrJWTMalformed
	}
	// alg confusion 攻撃 (none / HS256 で公開鍵を秘密鍵扱い) を防ぐため RS256 を強制。
	if alg, _ := header["alg"].(string); alg != "RS256" {
		return nil, ErrJWTBadAlg
	}
	kid, _ := header["kid"].(string)
	if kid == "" {
		return nil, ErrJWTMalformed
	}

	key, err := v.keyForKid(ctx, kid)
	if err != nil {
		return nil, err
	}

	if err := verifyRS256(parts[0]+"."+parts[1], parts[2], key); err != nil {
		return nil, err
	}

	claims, err := decodeJSONSegment(parts[1])
	if err != nil {
		return nil, ErrJWTMalformed
	}
	if err := v.verifyClaims(claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// verifyClaims は exp と iss を検証する。
func (v *Verifier) verifyClaims(claims map[string]any) error {
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().After(time.Unix(int64(exp), 0).Add(v.leeway)) {
			return ErrJWTExpired
		}
	} else {
		return ErrJWTMalformed
	}
	if v.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != v.issuer {
			return ErrJWTBadIssuer
		}
	}
	return nil
}

// keyForKid は kid に対応する RSA 公開鍵を返す。キャッシュに無ければ JWKS を再取得する。
func (v *Verifier) keyForKid(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if key, ok := v.lookup(kid); ok {
		return key, nil
	}

	// 未知 kid。refresh を 1 本に直列化し、待機中に他 goroutine が更新済みか / cooldown 内かを再確認する。
	v.refreshMu.Lock()
	defer v.refreshMu.Unlock()

	if key, ok := v.lookup(kid); ok {
		return key, nil
	}
	v.mu.RLock()
	stale := time.Since(v.fetchedAt) > v.refreshCooldown
	v.mu.RUnlock()
	if !stale {
		// 直前に別 goroutine が refresh 済み（かつ kid は依然見つからない）。連打せず未知扱い。
		return nil, ErrJWTUnknownKey
	}
	if err := v.refresh(ctx); err != nil {
		return nil, err
	}
	if key, ok := v.lookup(kid); ok {
		return key, nil
	}
	return nil, ErrJWTUnknownKey
}

// lookup は kid に対応する鍵をキャッシュから返す。
func (v *Verifier) lookup(kid string) (*rsa.PublicKey, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	key, ok := v.keys[kid]
	return key, ok
}

// jwk は JWKS 内の 1 鍵 (RSA) を表す。
type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// refresh は JWKS を取得してキャッシュを差し替える。
func (v *Verifier) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURI, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrJWKSUnavailable, err)
	}
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrJWKSUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrJWKSUnavailable, resp.StatusCode)
	}
	var doc struct {
		Keys []jwk `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return fmt.Errorf("%w: %w", ErrJWKSUnavailable, err)
	}
	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, k := range doc.Keys {
		if k.Kty != "RSA" || k.Kid == "" {
			continue
		}
		pub, err := k.toRSAPublicKey()
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	// 空 / 壊れた JWKS で既存の有効キャッシュを潰さない（認証全断の回避）。
	if len(keys) == 0 {
		return fmt.Errorf("%w: no usable rsa keys", ErrJWKSUnavailable)
	}
	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = time.Now()
	v.mu.Unlock()
	return nil
}

// toRSAPublicKey は JWK の n / e (base64url) から rsa.PublicKey を組み立てる。
func (k jwk) toRSAPublicKey() (*rsa.PublicKey, error) {
	nBytes, err := base64URLDecode(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64URLDecode(k.E)
	if err != nil {
		return nil, err
	}
	// 外部入力なので exponent の範囲を検証する（int 変換のオーバーフロー / 異常値を弾く）。
	eBig := new(big.Int).SetBytes(eBytes)
	if !eBig.IsInt64() {
		return nil, errors.New("cognito: jwk exponent too large")
	}
	e := eBig.Int64()
	if e <= 0 || e > math.MaxInt32 {
		return nil, errors.New("cognito: invalid jwk exponent")
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(e),
	}, nil
}

// verifyRS256 は signingInput (header.payload) の RS256 署名を公開鍵で検証する。
func verifyRS256(signingInput, sigSegment string, key *rsa.PublicKey) error {
	sig, err := base64URLDecode(sigSegment)
	if err != nil {
		return ErrJWTMalformed
	}
	hashed := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], sig); err != nil {
		return ErrJWTBadSignature
	}
	return nil
}

// decodeJSONSegment は base64url セグメントを JSON object にデコードする。
func decodeJSONSegment(seg string) (map[string]any, error) {
	raw, err := base64URLDecode(seg)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// base64URLDecode は JWT の URL-safe base64 (パディング省略可) をデコードする。
func base64URLDecode(s string) ([]byte, error) {
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.URLEncoding.DecodeString(s)
}
