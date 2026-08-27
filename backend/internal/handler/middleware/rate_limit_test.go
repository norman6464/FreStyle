package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_分間レートリミット_429を返す(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimitPerMinute(60, 2), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	codes := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}
	if codes[0] != http.StatusOK || codes[1] != http.StatusOK {
		t.Fatalf("first 2 should be 200, got %v", codes)
	}
	if codes[2] != http.StatusTooManyRequests {
		t.Fatalf("3rd should be 429, got %d", codes[2])
	}
}

// IP 単位の上限は XFF を変えるだけで抜けられる（gin の ClientIP が最左を読み、
// このリポジトリは SetTrustedProxies を呼んでいないため）。**塞げていないことを
// 明示的に固定しておく** — 秘密を守る上限をここに置いてはいけない、という根拠になる。
func Test_分間レートリミット_XFFを変えると効かない(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimitPerMinute(60, 2), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		req.Header.Set("X-Forwarded-For", "203.0.113."+strconv.Itoa(i%256))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("XFF を変えれば鍵が変わるので 429 にはならないはず: %d 回目で %d", i+1, w.Code)
		}
	}
}

func Test_分間レートリミット_鍵を差し替えられる(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 鍵をヘッダから作る（IP ではなく「守る対象」を鍵にする形の縮図）。
	r.GET("/x", RateLimitPerMinuteBy(60, 2, func(c *gin.Context) string {
		return c.GetHeader("X-Target")
	}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	call := func(target, xff string) int {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		req.Header.Set("X-Target", target)
		req.Header.Set("X-Forwarded-For", xff)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	// 同じ対象なら、IP をいくら変えても頭打ちになる。
	if call("a", "1.1.1.1") != http.StatusOK || call("a", "2.2.2.2") != http.StatusOK {
		t.Fatal("burst 内は通るはず")
	}
	if got := call("a", "3.3.3.3"); got != http.StatusTooManyRequests {
		t.Fatalf("IP を変えても同じ対象なら 429 になるはず: %d", got)
	}
	// 別の対象は巻き添えにしない。
	if got := call("b", "3.3.3.3"); got != http.StatusOK {
		t.Fatalf("別の鍵は独立しているはず: %d", got)
	}
}

func Test_分間レートリミット_鍵が空なら数えない(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimitPerMinuteBy(60, 1, func(_ *gin.Context) string { return "" }),
		func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("鍵が決まらない要求は数えない: %d 回目で %d", i+1, w.Code)
		}
	}
}

func Test_分間レートリミット_ユーザー単位はXFFで抜けられない(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", func(c *gin.Context) {
		c.Set(ContextKeyCurrentUserID, uint64(7))
		c.Next()
	}, RateLimitPerMinutePerUser(60, 2), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	codes := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		req.Header.Set("X-Forwarded-For", "203.0.113."+strconv.Itoa(i))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}
	if codes[0] != http.StatusOK || codes[1] != http.StatusOK {
		t.Fatalf("burst 内は通るはず: %v", codes)
	}
	if codes[2] != http.StatusTooManyRequests {
		t.Fatalf("同じユーザーなら XFF を変えても 429 になるはず: %v", codes)
	}
}
