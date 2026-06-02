package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rlBucket は 1 クライアント分のトークンバケット状態。
type rlBucket struct {
	tokens float64
	last   time.Time
}

// ipRateLimiter は IP ごとのトークンバケットで流量を制限する。
//
// 単一 ECS タスク (desiredCount=1) 前提の in-memory 実装。スケールアウトしたら
// インスタンスごとに別カウントになるため、その際は共有ストア (Redis 等) が必要。
// キーは gin の ClientIP()。ALB / CloudFront 経由の X-Forwarded-For を信用するため
// 厳密な詐称防止にはならない（フォームスパムや軽い総当たりの緩和が目的）。
type ipRateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*rlBucket
	rate        float64 // 毎秒の補充トークン数
	burst       float64 // バケット上限
	idleTTL     time.Duration
	lastCleanup time.Time
	now         func() time.Time // テスト差し替え用
}

func newIPRateLimiter(perMinute float64, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		buckets:     map[string]*rlBucket{},
		rate:        perMinute / 60.0,
		burst:       float64(burst),
		idleTTL:     10 * time.Minute,
		lastCleanup: time.Now(),
		now:         time.Now,
	}
}

// allow は key の 1 リクエスト分のトークンを消費できれば true を返す。
func (l *ipRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.cleanupLocked(now)

	b, ok := l.buckets[key]
	if !ok {
		b = &rlBucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	// 経過時間ぶんトークンを補充する。
	b.tokens += now.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// cleanupLocked は idleTTL を超えて使われていないバケットを掃除する（メモリ肥大化防止）。
func (l *ipRateLimiter) cleanupLocked(now time.Time) {
	if now.Sub(l.lastCleanup) < l.idleTTL {
		return
	}
	for k, b := range l.buckets {
		if now.Sub(b.last) > l.idleTTL {
			delete(l.buckets, k)
		}
	}
	l.lastCleanup = now
}

// RateLimitPerMinute は IP あたり perMinute 回（短期 burst まで許容）に制限する middleware を返す。
// 超過時は 429 + Retry-After を返す。各呼び出しが独立した limiter を持つ。
func RateLimitPerMinute(perMinute float64, burst int) gin.HandlerFunc {
	limiter := newIPRateLimiter(perMinute, burst)
	return func(c *gin.Context) {
		if !limiter.allow(c.ClientIP()) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "リクエストが多すぎます。しばらく時間をおいて再度お試しください。",
			})
			return
		}
		c.Next()
	}
}
