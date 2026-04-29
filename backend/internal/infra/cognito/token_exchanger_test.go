package cognito

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestServer は与えた handler でリクエストを受ける httptest.Server を立てる。
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, Config) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	cfg := Config{
		ClientID:     "client-xyz",
		ClientSecret: "secret-abc",
		RedirectURI:  "https://normanblog.com/auth/callback",
		TokenURI:     srv.URL,
	}
	return srv, cfg
}

func TestExchangeAuthorizationCode_Success(t *testing.T) {
	_, cfg := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		form := string(body)
		for _, want := range []string{"grant_type=authorization_code", "code=auth-code-123", "client_id=client-xyz", "client_secret=secret-abc"} {
			if !strings.Contains(form, want) {
				t.Errorf("body missing %q: %s", want, form)
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"AT","id_token":"IT","refresh_token":"RT","expires_in":3600,"token_type":"Bearer"}`))
	})

	tx := NewTokenExchanger(cfg)
	tok, err := tx.ExchangeAuthorizationCode(context.Background(), "auth-code-123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tok.AccessToken != "AT" || tok.RefreshToken != "RT" || tok.ExpiresIn != 3600 {
		t.Fatalf("unexpected token: %+v", tok)
	}
}

func TestRefreshAccessToken_Success(t *testing.T) {
	_, cfg := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form := string(body)
		if !strings.Contains(form, "grant_type=refresh_token") {
			t.Errorf("expected refresh_token grant: %s", form)
		}
		if !strings.Contains(form, "refresh_token=RT") {
			t.Errorf("missing refresh_token in body: %s", form)
		}
		_, _ = w.Write([]byte(`{"access_token":"AT2","expires_in":1800}`))
	})

	tx := NewTokenExchanger(cfg)
	tok, err := tx.RefreshAccessToken(context.Background(), "RT")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if tok.AccessToken != "AT2" || tok.ExpiresIn != 1800 {
		t.Fatalf("unexpected: %+v", tok)
	}
}

func TestExchange_NotConfigured(t *testing.T) {
	tx := NewTokenExchanger(Config{}) // ClientID / TokenURI 共に空
	_, err := tx.ExchangeAuthorizationCode(context.Background(), "code")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestExchange_TokenExchangeFailedWraps(t *testing.T) {
	_, cfg := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	})

	tx := NewTokenExchanger(cfg)
	_, err := tx.ExchangeAuthorizationCode(context.Background(), "bad-code")
	if !errors.Is(err, ErrTokenExchangeFailed) {
		t.Fatalf("want ErrTokenExchangeFailed, got %v", err)
	}
	var exErr *TokenExchangeError
	if !errors.As(err, &exErr) {
		t.Fatalf("expected *TokenExchangeError, got %T", err)
	}
	if exErr.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("HTTPStatus = %d", exErr.HTTPStatus)
	}
	if !strings.Contains(exErr.Body, "invalid_grant") {
		t.Fatalf("Body = %q", exErr.Body)
	}
}

func TestExchange_InvalidResponse(t *testing.T) {
	_, cfg := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	})

	tx := NewTokenExchanger(cfg)
	_, err := tx.ExchangeAuthorizationCode(context.Background(), "code")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("want ErrInvalidResponse, got %v", err)
	}
}

func TestExchange_Unreachable(t *testing.T) {
	cfg := Config{ClientID: "x", TokenURI: "http://127.0.0.1:1"} // 接続失敗
	tx := NewTokenExchangerWithClient(cfg, &http.Client{Timeout: 100 * time.Millisecond})
	_, err := tx.ExchangeAuthorizationCode(context.Background(), "code")
	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("want ErrUnreachable, got %v", err)
	}
}

func TestExchange_OmitsSecretWhenEmpty(t *testing.T) {
	_, cfg := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "client_secret=") {
			t.Errorf("client_secret should be omitted when empty: %s", string(body))
		}
		_, _ = w.Write([]byte(`{"access_token":"AT"}`))
	})
	cfg.ClientSecret = ""
	tx := NewTokenExchanger(cfg)
	if _, err := tx.ExchangeAuthorizationCode(context.Background(), "code"); err != nil {
		t.Fatalf("err: %v", err)
	}
}
