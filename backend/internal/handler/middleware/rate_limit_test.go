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

// IP 単位の上限は RealClientIP（X-Forwarded-For の末尾 = Cloud Run 自身が観測した
// 接続元）を鍵にする。要求元が書ける先頭側の要素をいくら変えても、Cloud Run が
// 追記する末尾は変わらないので、この上限は抜けられない。
func Test_分間レートリミット_XFFの先頭を変えても抜けられない(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimitPerMinute(60, 2), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	codes := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		// 先頭（要求元が自由に書ける部分）だけを毎回変える。末尾は固定
		// （Cloud Run が実際に観測した接続元を模している）。
		req.Header.Set("X-Forwarded-For", "203.0.113."+strconv.Itoa(i)+", 198.51.100.7")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}
	if codes[0] != http.StatusOK || codes[1] != http.StatusOK {
		t.Fatalf("burst 内は通るはず: %v", codes)
	}
	if codes[2] != http.StatusTooManyRequests {
		t.Fatalf("先頭を変えても末尾（鍵）が同じなら 429 になるはず: %v", codes)
	}
}

// 末尾（Cloud Run が観測した接続元）が本当に違う相手どうしは、互いの上限を巻き添えにしない。
func Test_分間レートリミット_XFFの末尾が違えば別の鍵になる(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", RateLimitPerMinute(60, 2), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	call := func(realIP string) int {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		req.Header.Set("X-Forwarded-For", "203.0.113.9, "+realIP)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}
	for i := 0; i < 2; i++ {
		if got := call("198.51.100.1"); got != http.StatusOK {
			t.Fatalf("1人目の burst 内は通るはず: %d 回目で %d", i+1, got)
		}
	}
	if got := call("198.51.100.1"); got != http.StatusTooManyRequests {
		t.Fatalf("1人目は 3 回目で 429 になるはず: %d", got)
	}
	if got := call("198.51.100.2"); got != http.StatusOK {
		t.Fatalf("末尾が違う2人目は巻き添えにしないはず: %d", got)
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
