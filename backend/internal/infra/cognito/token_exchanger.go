// Package cognito は AWS Cognito User Pool との通信を Go の純粋なコードに閉じ込める。
// handler 層はこのパッケージのみに依存し、url.Values / http.Client / OAuth2 token endpoint
// の URL 組み立てといった低レベルな詳細は知らないようにする（Clean Architecture の Infra 層）。
package cognito

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 既定の HTTP 通信タイムアウト。Cognito token endpoint への通信が
// 無限待ちにならないよう必ず timeout を設定する。
const defaultHTTPTimeout = 10 * time.Second

// Token は Cognito の OAuth2 token endpoint レスポンスを表現する。
// 各フィールド名は AWS Cognito Hosted UI / OIDC Discovery 仕様の token response に対応。
//
// 参考: https://docs.aws.amazon.com/cognito/latest/developerguide/token-endpoint.html
type Token struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Config は TokenExchanger が必要とする Cognito 設定。
// infra/config.CognitoConfig から必要部分のみコピーして渡す。
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	TokenURI     string
}

// TokenExchanger は Cognito User Pool の OAuth2 token endpoint を叩いて
// authorization_code / refresh_token を access/id/refresh token に交換する。
//
// 旧実装は handler/auth_handler.go の Callback / Refresh に同じ HTTP 構築・
// レスポンス処理が ~140 行重複していたが、grant_type だけ違う構造だったため
// このパッケージに集約した。
type TokenExchanger struct {
	cfg        Config
	httpClient *http.Client
}

// NewTokenExchanger は HTTP timeout を含めた標準クライアントで TokenExchanger を組み立てる。
func NewTokenExchanger(cfg Config) *TokenExchanger {
	return &TokenExchanger{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// NewTokenExchangerWithClient はテスト時に http.Client を差し替えるための DI コンストラクタ。
func NewTokenExchangerWithClient(cfg Config, client *http.Client) *TokenExchanger {
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return &TokenExchanger{cfg: cfg, httpClient: client}
}

// 既定で発生し得るエラー値。エラーラップ越しに呼び元が分岐できるよう sentinel として公開する。
var (
	// ErrNotConfigured は ClientID / TokenURI が空の状態で呼ばれたときに返る。
	// handler 側で 500 (cognito_not_configured) に変換することを想定。
	ErrNotConfigured = errors.New("cognito: not configured")

	// ErrUnreachable は token endpoint への HTTP リクエスト自体が失敗したとき。
	// (DNS / TLS / connect / read deadline 等) handler 側で 502 を返す目安。
	ErrUnreachable = errors.New("cognito: token endpoint unreachable")

	// ErrInvalidResponse はステータス 200 だが JSON が壊れているとき。
	ErrInvalidResponse = errors.New("cognito: invalid token response")

	// ErrTokenExchangeFailed は Cognito が 4xx/5xx を返したとき。
	// HTTPStatus / Body を見たい場合は TokenExchangeError 型で wrap される。
	ErrTokenExchangeFailed = errors.New("cognito: token exchange failed")
)

// TokenExchangeError は Cognito が non-2xx を返したときの詳細を保持する。
// errors.Is(err, ErrTokenExchangeFailed) で判定可能。
// CloudWatch Logs への記録用に HTTPStatus / Body を残す（handler 側でログ済）。
type TokenExchangeError struct {
	HTTPStatus int
	Body       string
}

func (e *TokenExchangeError) Error() string {
	return fmt.Sprintf("cognito: token exchange failed: status=%d body=%s", e.HTTPStatus, e.Body)
}

func (e *TokenExchangeError) Unwrap() error { return ErrTokenExchangeFailed }

// ExchangeAuthorizationCode は Cognito Hosted UI から戻った認可コードを token に交換する。
// Spring Boot CognitoAuthController#callback と等価。
func (t *TokenExchanger) ExchangeAuthorizationCode(ctx context.Context, code string) (*Token, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", t.cfg.RedirectURI)
	return t.exchange(ctx, form)
}

// RefreshAccessToken は HttpOnly Cookie の refresh_token を使って access_token を再発行する。
func (t *TokenExchanger) RefreshAccessToken(ctx context.Context, refreshToken string) (*Token, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	return t.exchange(ctx, form)
}

// exchange は authorization_code / refresh_token grant の差を吸収した内部実装。
//
// Cognito の OAuth2 token endpoint には「Authorization Basic header 方式」と
// 「body 方式 (client_id + client_secret を form に入れる)」があるが、両方送ると
// invalid_client を返すケースがある。本実装では body 方式に統一する（AWS docs 推奨）。
func (t *TokenExchanger) exchange(ctx context.Context, form url.Values) (*Token, error) {
	if t.cfg.TokenURI == "" || t.cfg.ClientID == "" {
		return nil, ErrNotConfigured
	}

	form.Set("client_id", t.cfg.ClientID)
	if t.cfg.ClientSecret != "" {
		form.Set("client_secret", t.cfg.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.cfg.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		// NewRequestWithContext は URL parse 不能などで失敗する。
		// 設定不備に近いので handler 側で 500 にする想定。
		return nil, fmt.Errorf("cognito: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		// 4xx/5xx は本物の理由 (invalid_grant / redirect_uri_mismatch 等) を保持して返す。
		return nil, &TokenExchangeError{HTTPStatus: resp.StatusCode, Body: string(body)}
	}

	var tok Token
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	return &tok, nil
}
