package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/infra/ratelimit"
)

// RespondRateLimited は上限超過の応答を書いて処理を止める。
//
// 上限は middleware（IP 単位）だけでなく handler 側（守る対象そのものを鍵にするもの）でも
// 掛けるので、応答の形をこの 1 関数に閉じる。どちらで断られたかで本文が変わると、
// 呼び出し側は 2 通りの 429 を扱わされ、区別に意味が無いのに分岐が増える。
func RespondRateLimited(c *gin.Context) {
	c.Header("Retry-After", "60")
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":   "rate_limited",
		"message": "リクエストが多すぎます。しばらく時間をおいて再度お試しください。",
	})
}

// RateLimitPerMinute は IP あたり perMinute 回（短期 burst まで許容）に制限する middleware を返す。
// 超過時は 429 + Retry-After。各呼び出しが独立した limiter を持つ。
//
// # 鍵は RealClientIP（Cloud Run が追記した末尾の IP）
//
// 鍵は RealClientIP（client_ip.go）で、X-Forwarded-For の末尾（Cloud Run 自身が観測した
// 接続元）を読む。先頭側の要素は要求元が自由に書けるが、末尾だけは詐称できないので、
// XFF を偽って鍵を変え続けることはできない（client_ip.go の doc 参照）。
//
// とはいえ IP は「その場所にいる全員」で共有される値で、家庭やオフィスの NAT の裏に
// いる複数人が同じ IP を踏むこともある。秘密（パスワード等）を守る上限は、鍵を
// **守る対象そのもの** に取ること。共有リンクの検証がその例で、リンク 1 本ごとの
// 上限を handler 側で別に掛けている（kb_share_link_handler.go）。
func RateLimitPerMinute(perMinute float64, burst int) gin.HandlerFunc {
	return RateLimitPerMinuteBy(perMinute, burst, RealClientIP)
}

// RateLimitPerMinuteBy は鍵の作り方を差し替えられる RateLimitPerMinute。
//
// key が空文字を返した要求は数えない（鍵が決まらない ＝ 数える相手が居ない）。
// 鍵に何を選ぶかがこの middleware の効き目のすべてなので、選び方は呼び出し側に置く。
func RateLimitPerMinuteBy(perMinute float64, burst int, key func(c *gin.Context) string) gin.HandlerFunc {
	limiter := ratelimit.New(perMinute, burst)
	return func(c *gin.Context) {
		k := key(c)
		if k != "" && !limiter.Allow(k) {
			RespondRateLimited(c)
			return
		}
		c.Next()
	}
}

// RateLimitPerMinutePerUser はログイン済みユーザー 1 人あたりで制限する middleware を返す。
//
// 認証済みのルートではこちらを使う。ユーザー ID は検証済みの JWT から来るので
// 攻撃者が付け替えられない（IP は XFF で付け替えられる）。
// 認証前に呼ばれた場合は鍵が決まらないので数えない — このルートに認証 middleware が
// 掛かっていること自体が前提で、掛かっていなければ数えるより先に 401 になる。
func RateLimitPerMinutePerUser(perMinute float64, burst int) gin.HandlerFunc {
	return RateLimitPerMinuteBy(perMinute, burst, func(c *gin.Context) string {
		if uid := CurrentUserIDOrZero(c); uid != 0 {
			return "user:" + strconv.FormatUint(uid, 10)
		}
		return ""
	})
}
