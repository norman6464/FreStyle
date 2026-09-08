package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_許可オリジン判定(t *testing.T) {
	cases := []struct {
		origin string
		want   bool
	}{
		{"https://frestyle.dev", true},
		{"https://frestyle-dev.web.app", true},
		{"https://frestyle-dev.firebaseapp.com", true},
		{"http://localhost:5173", true},
		// 旧ドメインはすべて撤去済み（DNS 削除 / AWS 撤去）のため許可しない。
		{"https://frestyle.jp", false},
		{"https://dcd3m6lwt0z8u.cloudfront.net", false},
		{"http://fre-style-bucket.s3-website-ap-northeast-1.amazonaws.com", false},
		{"https://normanblog.com", false},
		{"http://normanblog.com", false},
		{"https://evil.example.com", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.origin, func(t *testing.T) {
			if got := IsAllowedOrigin(tc.origin); got != tc.want {
				t.Fatalf("IsAllowedOrigin(%q) = %v, want %v", tc.origin, got, tc.want)
			}
		})
	}
}

// IsAllowedOrigin だけでなく、実際に gin のミドルウェアとして動かしたときに
// ヘッダが正しく付く/付かないところまで見る。判定と応答の配線が食い違っている
// バグ（判定は正しいのにヘッダを付け忘れる等）は前者だけでは捕まえられない。
func Test_CORSミドルウェア_許可オリジンにはヘッダを付け未許可には付けない(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name        string
		origin      string
		wantAllowed bool
	}{
		{"本番カスタムドメイン", "https://frestyle.dev", true},
		{"Firebase Hosting既定URL", "https://frestyle-dev.web.app", true},
		{"Firebase既定authDomain", "https://frestyle-dev.firebaseapp.com", true},
		{"ローカル開発", "http://localhost:5173", true},
		{"撤去済みの旧ドメイン", "https://frestyle.jp", false},
		{"無関係なオリジン", "https://evil.example.com", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORS())
			router.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req.Header.Set("Origin", tc.origin)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			got := rec.Header().Get("Access-Control-Allow-Origin")
			if tc.wantAllowed {
				if got != tc.origin {
					t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, tc.origin)
				}
				if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
					t.Fatalf("Access-Control-Allow-Credentials が付いていない")
				}
			} else {
				if got != "" {
					t.Fatalf("Access-Control-Allow-Origin が付いてはいけない（許可していないオリジン）が %q が付いた", got)
				}
			}
		})
	}
}

// Preflight（OPTIONS）は許可オリジンかどうかによらず 204 で終端する
// （許可オリジンでなければヘッダは付かないが、ブラウザ側の実リクエストは
// そのヘッダ欠如自体で弾かれるため、ここで 403 等に分ける必要はない）。
func Test_CORSミドルウェア_Preflightは204で終端する(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	called := false
	router.OPTIONS("/ping", func(c *gin.Context) { called = true })

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "https://frestyle.dev")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if called {
		t.Fatal("OPTIONS は後続ハンドラまで到達せず、CORS ミドルウェアで終端すること")
	}
}
