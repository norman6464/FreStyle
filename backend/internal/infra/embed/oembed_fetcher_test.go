package embed

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestFetcher(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Fetcher) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	// httptest.NewTLSServer は自己署名なので InsecureSkipVerify を有効にした client を渡す。
	// 本番では https + 公開証明書を前提とするため通常の http.Client を使う。
	tx := NewFetcherWithClient(&http.Client{
		Timeout:   3 * time.Second,
		Transport: srv.Client().Transport,
	})
	return srv, tx
}

func Test_解決_成功(t *testing.T) {
	srv, tx := newTestFetcher(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`
			<html>
			<head>
				<meta property="og:title" content="Hello Title">
				<meta property="og:description" content="Hello Desc">
				<meta property="og:image" content="https://example.invalid/img.png">
				<meta property="og:site_name" content="Example">
			</head>
			<body></body>
			</html>
		`))
	})

	card, err := tx.Resolve(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if card.Title != "Hello Title" || card.Description != "Hello Desc" {
		t.Fatalf("unexpected card: %+v", card)
	}
	if card.ImageURL != "https://example.invalid/img.png" || card.SiteName != "Example" {
		t.Fatalf("unexpected card: %+v", card)
	}
	if card.Provider != "ogp" {
		t.Fatalf("provider = %q", card.Provider)
	}
}

func Test_解決_titleタグにフォールバック(t *testing.T) {
	srv, tx := newTestFetcher(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Plain Title</title></head><body></body></html>`))
	})

	card, err := tx.Resolve(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if card.Title != "Plain Title" {
		t.Fatalf("title = %q", card.Title)
	}
}

func Test_解決_非HTTPSを拒否(t *testing.T) {
	tx := NewFetcher()
	_, err := tx.Resolve(context.Background(), "http://example.com")
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("want ErrInvalidURL, got %v", err)
	}
}

func Test_解決_不正なURLを拒否(t *testing.T) {
	tx := NewFetcher()
	_, err := tx.Resolve(context.Background(), "::not a url::")
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("want ErrInvalidURL, got %v", err)
	}
}

func Test_解決_プライベートホストを拒否(t *testing.T) {
	tx := NewFetcher()
	for _, host := range []string{
		"https://localhost/",
		"https://127.0.0.1/",
		"https://10.0.0.1/",
		"https://192.168.1.1/",
		"https://169.254.169.254/latest/meta-data/",
	} {
		if _, err := tx.Resolve(context.Background(), host); !errors.Is(err, ErrUnsupportedHost) {
			t.Errorf("host=%s want ErrUnsupportedHost, got %v", host, err)
		}
	}
}

func Test_解決_到達不能(t *testing.T) {
	srv, tx := newTestFetcher(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := tx.Resolve(context.Background(), srv.URL)
	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("want ErrUnreachable, got %v", err)
	}
}

func Test_解決_キャッシュ(t *testing.T) {
	calls := 0
	srv, tx := newTestFetcher(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`<html><head><title>X</title></head></html>`))
	})

	for i := 0; i < 3; i++ {
		if _, err := tx.Resolve(context.Background(), srv.URL); err != nil {
			t.Fatalf("err: %v", err)
		}
	}
	if calls != 1 {
		t.Fatalf("expected single backend call (cached), got %d", calls)
	}
}

func Test_解決_ホスト名をタイトルにフォールバック(t *testing.T) {
	srv, tx := newTestFetcher(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body>no title</body></html>`))
	})

	card, err := tx.Resolve(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// host fallback (httptest URL の host 部分) を含むことを確認
	if !strings.Contains(card.Title, "127.0.0.1") {
		t.Fatalf("expected host fallback title, got %q", card.Title)
	}
}
