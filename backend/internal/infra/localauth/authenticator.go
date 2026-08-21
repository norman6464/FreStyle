// Package localauth はローカル開発専用のパスワードログイン実装を提供する。
// users.password_hash（bcrypt）を検証し、HMAC 署名付きの JWT を発行する。
// 検証は本パッケージの VerifyToken が HMAC 署名・iss・exp を確認し（router.buildJWTVerify が
// 呼ぶ）、localauth 発行でないトークンは従来経路（JWKS 等）へフォールバックされる。
// **本番では絶対に配線しない**（New が local 以外を拒否し、wiring / verify 側も二重に弾く。
// FRESTYLE-311 / FRESTYLE-249 の決定）。
package localauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"golang.org/x/crypto/bcrypt"
)

// UserSource は authenticator が必要とする users への読み書き境界。
// persistence の UserRepository が構造的に満たす（wiring は router で行う）。
type UserSource interface {
	FindActiveByEmail(ctx context.Context, email string) (*domain.User, error)
	CognitoSubjectByUserID(ctx context.Context, userID uint64) (string, error)
	EnsureOidcIdentity(ctx context.Context, userID uint64, provider, subject string) error
}

// Authenticator は email / password を DB の bcrypt ハッシュで検証し token を返す。
// handler の passwordAuthenticator interface（Authenticate）を実装する。
type Authenticator struct {
	users UserSource
	now   func() time.Time
}

// dummyHash はユーザー不在時にも bcrypt 比較を 1 回走らせるためのダミー（"dummy" のハッシュ）。
// 「存在する email か」を応答時間から推測されにくくする。
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// New はローカル専用 authenticator を作る。APP_ENV が local 以外なら fail closed でエラー。
func New(users UserSource, appEnv string) (*Authenticator, error) {
	if appEnv != "local" {
		return nil, fmt.Errorf("localauth: APP_ENV=%q では使用できません（local 専用）", appEnv)
	}
	if users == nil {
		return nil, fmt.Errorf("localauth: user source が未設定です")
	}
	return &Authenticator{users: users, now: time.Now}, nil
}

// Authenticate は email / password を検証し、Cognito 互換の token 一式を返す。
// 資格情報の誤り（ユーザー不在 / パスワード未設定 / 不一致）は詳細を区別せず
// cognito.ErrInvalidCredentials を返す（存在の有無を漏らさない）。
func (a *Authenticator) Authenticate(ctx context.Context, email, password string) (*cognito.Token, error) {
	user, err := a.users.FindActiveByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("localauth: find user: %w", err)
	}
	// loginable の否定形とハッシュ選択が同じ述語になるよう 1 か所で判定する。
	loginable := user != nil && user.PasswordHash != nil && *user.PasswordHash != ""
	hash := dummyHash
	if loginable {
		hash = *user.PasswordHash
	}
	compareErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if !loginable || compareErr != nil {
		return nil, cognito.ErrInvalidCredentials
	}

	subject, err := a.subjectFor(ctx, user)
	if err != nil {
		return nil, err
	}

	now := a.now()
	const ttl = 24 * time.Hour // ローカル開発用。refresh は Cognito 前提のため失効後は再ログインする。
	claims := map[string]any{
		"sub":   subject,
		"email": user.Email,
		"name":  user.Name,
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
		"iss":   Issuer,
	}
	token, err := mintJWT(claims)
	if err != nil {
		return nil, fmt.Errorf("localauth: mint token: %w", err)
	}
	refresh := make([]byte, 16)
	if _, err := rand.Read(refresh); err != nil {
		return nil, fmt.Errorf("localauth: refresh token: %w", err)
	}
	return &cognito.Token{
		// access / id を同一 claims で発行する。middleware は sub / email / name しか読まず、
		// cognito:groups 無し = グループ由来の権限昇格は起きない（ロールは DB が正）。
		AccessToken:  token,
		IDToken:      token,
		RefreshToken: "local-" + hex.EncodeToString(refresh),
		// access_token Cookie の maxAge を JWT の exp（24h）に合わせる。未設定だと handler 側が
		// 既定 1 時間へ丸め、JWT は生きているのに Cookie だけ切れて 1h で 401 になる。
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}

// subjectFor はトークンの sub を決める。既存の OIDC identity があればその subject
// （seed ユーザーなら seed-sub-N）、無ければ決定的な値を生成して identity を張る。
// これにより下流（upsert の FindByCognitoSub / 招待ゲート / ロール同期）は Cognito ログインと
// 完全に同じ経路を通る。
func (a *Authenticator) subjectFor(ctx context.Context, user *domain.User) (string, error) {
	subject, err := a.users.CognitoSubjectByUserID(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("localauth: subject lookup: %w", err)
	}
	if subject != "" {
		return subject, nil
	}
	subject = fmt.Sprintf("local-pw-%d", user.ID)
	if err := a.users.EnsureOidcIdentity(ctx, user.ID, domain.OidcProviderCognito, subject); err != nil {
		return "", fmt.Errorf("localauth: ensure identity: %w", err)
	}
	return subject, nil
}
