package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// newTokenServer は token endpoint を模したサーバを立て、受け取ったフォームを記録する。
func newTokenServer(t *testing.T, status int, body string) (*httptest.Server, *url.Values) {
	t.Helper()
	got := &url.Values{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		parsed, err := url.ParseQuery(string(raw))
		if err != nil {
			t.Errorf("フォームを読めない: %v", err)
		}
		*got = parsed
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(s.Close)
	return s, got
}

func okTokenBody(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"access_token": "at", "id_token": "it", "refresh_token": "rt",
		"expires_in": 3600, "token_type": "Bearer",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

// PKCE の検証値を送っていること。公開クライアントではこれが唯一の
// 「認可を始めた本人か」の確かめ方なので、送り忘れると交換自体が通らない。
func Test_トークン交換_code_verifierを送る(t *testing.T) {
	srv, got := newTokenServer(t, http.StatusOK, okTokenBody(t))
	ex := NewTokenExchanger(ExchangerConfig{
		ClientID: "c1", RedirectURI: "http://localhost/cb", TokenURI: srv.URL,
	})

	if _, err := ex.ExchangeAuthorizationCode(context.Background(), "the-code", "the-verifier"); err != nil {
		t.Fatalf("交換に失敗: %v", err)
	}
	if got.Get("code_verifier") != "the-verifier" {
		t.Errorf("code_verifier = %q", got.Get("code_verifier"))
	}
	if got.Get("grant_type") != "authorization_code" {
		t.Errorf("grant_type = %q", got.Get("grant_type"))
	}
	if got.Get("code") != "the-code" {
		t.Errorf("code = %q", got.Get("code"))
	}
	if got.Get("redirect_uri") != "http://localhost/cb" {
		t.Errorf("redirect_uri = %q", got.Get("redirect_uri"))
	}
}

// 秘密を持たない公開クライアントでは client_secret を送らない。
// 空文字で送ると、発行者によっては invalid_client で弾かれる。
func Test_トークン交換_秘密が無ければclient_secretを送らない(t *testing.T) {
	srv, got := newTokenServer(t, http.StatusOK, okTokenBody(t))
	ex := NewTokenExchanger(ExchangerConfig{ClientID: "c1", TokenURI: srv.URL})

	if _, err := ex.ExchangeAuthorizationCode(context.Background(), "code", "verifier"); err != nil {
		t.Fatalf("交換に失敗: %v", err)
	}
	if _, ok := (*got)["client_secret"]; ok {
		t.Error("client_secret を送ってしまっている")
	}
}

func Test_トークン交換_秘密があれば送る(t *testing.T) {
	srv, got := newTokenServer(t, http.StatusOK, okTokenBody(t))
	ex := NewTokenExchanger(ExchangerConfig{ClientID: "c1", ClientSecret: "s3cret", TokenURI: srv.URL})

	if _, err := ex.ExchangeAuthorizationCode(context.Background(), "code", ""); err != nil {
		t.Fatalf("交換に失敗: %v", err)
	}
	if got.Get("client_secret") != "s3cret" {
		t.Errorf("client_secret = %q", got.Get("client_secret"))
	}
}

// 検証値が空なら送らない（PKCE を使わない構成で、空の code_verifier を送って弾かれないように）。
func Test_トークン交換_検証値が空なら送らない(t *testing.T) {
	srv, got := newTokenServer(t, http.StatusOK, okTokenBody(t))
	ex := NewTokenExchanger(ExchangerConfig{ClientID: "c1", TokenURI: srv.URL})

	if _, err := ex.ExchangeAuthorizationCode(context.Background(), "code", ""); err != nil {
		t.Fatalf("交換に失敗: %v", err)
	}
	if _, ok := (*got)["code_verifier"]; ok {
		t.Error("空の code_verifier を送ってしまっている")
	}
}

func Test_トークン更新_リフレッシュトークンを送る(t *testing.T) {
	srv, got := newTokenServer(t, http.StatusOK, okTokenBody(t))
	ex := NewTokenExchanger(ExchangerConfig{ClientID: "c1", TokenURI: srv.URL})

	tok, err := ex.RefreshAccessToken(context.Background(), "old-rt")
	if err != nil {
		t.Fatalf("更新に失敗: %v", err)
	}
	if got.Get("grant_type") != "refresh_token" || got.Get("refresh_token") != "old-rt" {
		t.Errorf("form = %v", *got)
	}
	// 回転した新しい値が呼び元へ返ること（呼び元はこれを保存し直す責任がある）。
	if tok.RefreshToken != "rt" {
		t.Errorf("refresh_token = %q", tok.RefreshToken)
	}
}

func Test_トークン交換_未設定なら呼ぶ前に落とす(t *testing.T) {
	ex := NewTokenExchanger(ExchangerConfig{})
	if _, err := ex.ExchangeAuthorizationCode(context.Background(), "code", "v"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err = %v, want ErrNotConfigured", err)
	}
}

// 発行者が返した本当の理由（invalid_grant 等）を保って返す。
// ここを潰すと、切り分けのときに「何が悪いのか分からない」状態になる。
func Test_トークン交換_発行者のエラーを保って返す(t *testing.T) {
	srv, _ := newTokenServer(t, http.StatusBadRequest, `{"error":"invalid_grant"}`)
	ex := NewTokenExchanger(ExchangerConfig{ClientID: "c1", TokenURI: srv.URL})

	_, err := ex.ExchangeAuthorizationCode(context.Background(), "code", "v")
	var exErr *TokenExchangeError
	if !errors.As(err, &exErr) {
		t.Fatalf("err = %v, want *TokenExchangeError", err)
	}
	if exErr.HTTPStatus != http.StatusBadRequest {
		t.Errorf("status = %d", exErr.HTTPStatus)
	}
	if !errors.Is(err, ErrTokenExchangeFailed) {
		t.Error("sentinel で分岐できない")
	}
}
