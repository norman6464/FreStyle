package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// allowedOrigins は CORS で許可するオリジン。
var allowedOrigins = map[string]struct{}{
	"https://normanblog.com":                                          {},
	"http://normanblog.com":                                           {},
	"http://localhost:5173":                                           {},
	"https://dcd3m6lwt0z8u.cloudfront.net":                            {},
	"http://fre-style-bucket.s3-website-ap-northeast-1.amazonaws.com": {},
}

const (
	allowMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
	allowHeaders = "Content-Type, Authorization, X-Requested-With"
	maxAge       = "3600"
)

// IsAllowedOrigin は CORS middleware 外からも同じ allowlist を使えるようにするヘルパ。
func IsAllowedOrigin(origin string) bool {
	_, ok := allowedOrigins[origin]
	return ok
}

// CORS は許可リストの Origin のみ credentials 付きで許可し、Preflight は 204 で終端する。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if IsAllowedOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Max-Age", maxAge)
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
