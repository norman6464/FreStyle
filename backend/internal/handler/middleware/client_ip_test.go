package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func realClientIPOf(req *http.Request) string {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	return RealClientIP(c)
}

func Test_RealClientIP(t *testing.T) {
	t.Run("XFFが無ければRemoteAddrのホスト部分", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.9:54321"
		if got := realClientIPOf(req); got != "203.0.113.9" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("XFFが1要素だけならその値", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:1"
		req.Header.Set("X-Forwarded-For", "198.51.100.1")
		if got := realClientIPOf(req); got != "198.51.100.1" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("XFFが複数要素なら末尾を使う_先頭は要求元が書けるので信用しない", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:1"
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 203.0.113.2, 198.51.100.9")
		if got := realClientIPOf(req); got != "198.51.100.9" {
			t.Fatalf("got %q, want the last entry", got)
		}
	})

	t.Run("末尾の前後に空白があってもトリムする", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:1"
		req.Header.Set("X-Forwarded-For", "203.0.113.1,  198.51.100.9  ")
		if got := realClientIPOf(req); got != "198.51.100.9" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("末尾がIPとして不正ならRemoteAddrへ倒す", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "9.9.9.9:1"
		req.Header.Set("X-Forwarded-For", "203.0.113.1, not-an-ip")
		if got := realClientIPOf(req); got != "9.9.9.9" {
			t.Fatalf("got %q", got)
		}
	})
}
