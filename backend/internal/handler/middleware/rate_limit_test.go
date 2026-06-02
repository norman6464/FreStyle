package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestIPRateLimiter_BurstThenDeny(t *testing.T) {
	l := newIPRateLimiter(60, 3) // 3 burst
	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}
	if l.allow("1.2.3.4") {
		t.Fatal("4th request should be denied after burst is exhausted")
	}
}

func TestIPRateLimiter_PerKeyIndependent(t *testing.T) {
	l := newIPRateLimiter(60, 1)
	if !l.allow("a") || !l.allow("b") {
		t.Fatal("different IPs should have independent buckets")
	}
	if l.allow("a") {
		t.Fatal("same IP should be limited after its own burst")
	}
}

func TestIPRateLimiter_RefillsOverTime(t *testing.T) {
	cur := time.Now()
	l := newIPRateLimiter(60, 1) // 1 token/sec
	l.now = func() time.Time { return cur }

	if !l.allow("ip") {
		t.Fatal("first request should pass")
	}
	if l.allow("ip") {
		t.Fatal("immediate second request should be denied")
	}
	// 1 秒経過させると 1 トークン補充される。
	cur = cur.Add(1100 * time.Millisecond)
	if !l.allow("ip") {
		t.Fatal("request after refill window should pass")
	}
}

func TestRateLimitPerMinute_Returns429(t *testing.T) {
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
