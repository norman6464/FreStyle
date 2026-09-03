package oidc

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

// defaultHTTPTimeout は token endpoint への通信が無限に待たないための上限。
const defaultHTTPTimeout = 10 * time.Second

// maxTokenResponseBytes は token 応答の読み取り上限。
const maxTokenResponseBytes = 1 << 20 // 1 MiB

// Token は OAuth2 の token endpoint が返す応答。
type Token struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// ExchangerConfig は TokenExchanger が必要とする設定。
type ExchangerConfig struct {
	ClientID string
	// ClientSecret は機密クライアントのときだけ設定する。
	// 空なら公開クライアントとして扱い、client_secret を送らない。
	// ブラウザで動くアプリは秘密を持てないので、こちらが既定の形になる。
	ClientSecret string
	RedirectURI  string
	TokenURI     string
}

// TokenExchanger は認可コードとリフレッシュトークンを token に交換する。
type TokenExchanger struct {
	cfg        ExchangerConfig
	httpClient *http.Client
}

// NewTokenExchanger は既定のタイムアウトつきで TokenExchanger を組み立てる。
func NewTokenExchanger(cfg ExchangerConfig) *TokenExchanger {
	return &TokenExchanger{cfg: cfg, httpClient: &http.Client{Timeout: defaultHTTPTimeout}}
}

// NewTokenExchangerWithClient はテストで通信先を差し替えるためのコンストラクタ。
func NewTokenExchangerWithClient(cfg ExchangerConfig, client *http.Client) *TokenExchanger {
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return &TokenExchanger{cfg: cfg, httpClient: client}
}

// 呼び元がラップ越しに分岐できるよう sentinel として公開する。
var (
	// ErrNotConfigured は ClientID / TokenURI が空の状態で呼ばれたとき。
	ErrNotConfigured = errors.New("oidc: not configured")

	// ErrUnreachable は token endpoint への HTTP リクエスト自体が失敗したとき。
	ErrUnreachable = errors.New("oidc: token endpoint unreachable")

	// ErrInvalidResponse は 200 だが JSON が壊れているとき。
	ErrInvalidResponse = errors.New("oidc: invalid token response")

	// ErrTokenExchangeFailed は発行者が 4xx/5xx を返したとき。
	ErrTokenExchangeFailed = errors.New("oidc: token exchange failed")
)

// TokenExchangeError は発行者が non-2xx を返したときの詳細を保持する。
type TokenExchangeError struct {
	HTTPStatus int
	Body       string
}

func (e *TokenExchangeError) Error() string {
	return fmt.Sprintf("oidc: token exchange failed: status=%d body=%s", e.HTTPStatus, e.Body)
}

func (e *TokenExchangeError) Unwrap() error { return ErrTokenExchangeFailed }

// ExchangeAuthorizationCode は認可コードを token に交換する。
//
// codeVerifier は PKCE の検証値。ブラウザが認可を始めるときに乱数を作って手元に置き、
// その要約（code_challenge）だけを認可要求に載せる。交換のときに元の値を出すことで、
// 「この認可を始めたのと同じ相手か」を発行者が確かめられる。
//
// これが要るのは、認可コードが URL に載って戻ってくるから。履歴・ログ・別アプリの
// リダイレクト先の乗っ取りなど、コードが他人の手に渡る経路がいくつもある。
// 秘密を持てるサーバなら client_secret がその役目を果たすが、ブラウザで動くアプリは
// 秘密を持てない。PKCE は「秘密を毎回作って 1 回だけ使う」ことで同じ働きをさせる。
//
// 空文字なら code_verifier を送らない（PKCE を使わない発行者・設定のため）。
func (t *TokenExchanger) ExchangeAuthorizationCode(ctx context.Context, code, codeVerifier string) (*Token, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", t.cfg.RedirectURI)
	if codeVerifier != "" {
		form.Set("code_verifier", codeVerifier)
	}
	return t.exchange(ctx, form)
}

// RefreshAccessToken は refresh_token で access_token を再発行する。
//
// 発行者によっては、交換のたびに refresh_token 自体も新しいものへ入れ替わる（回転）。
// 呼び元は返ってきた RefreshToken が空でなければ必ず保存し直すこと。古いものを
// 持ち続けると、次の交換で「使用済みのトークンの再利用」と見なされて失敗する。
func (t *TokenExchanger) RefreshAccessToken(ctx context.Context, refreshToken string) (*Token, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	return t.exchange(ctx, form)
}

// exchange は grant 種別の差を吸収した内部実装。
// クライアント認証は body 方式に統一する（Basic ヘッダと併用すると
// invalid_client を返す発行者があるため）。
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
		return nil, fmt.Errorf("oidc: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnreachable, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxTokenResponseBytes))
	if resp.StatusCode != http.StatusOK {
		// 本当の理由（invalid_grant / redirect_uri_mismatch 等）を持って返す。
		return nil, &TokenExchangeError{HTTPStatus: resp.StatusCode, Body: string(body)}
	}

	var tok Token
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}
	// 200 でも中身が空のことはある（発行者の不調・前段のプロキシが本文を落とす等）。
	// そのまま返すと、空文字を Cookie に書いて「ログインできたのに全部 401」になる。
	// 落ちる側に倒して、原因が分かる形で止める。
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("%w: access_token が空", ErrInvalidResponse)
	}
	return &tok, nil
}
