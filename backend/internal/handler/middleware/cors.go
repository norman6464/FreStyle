package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 許可するオリジン。Spring Boot の CorsConfig と同じ値を Go 側でも維持する。
var allowedOrigins = map[string]struct{}{
	"https://normanblog.com":      {},
	"http://normanblog.com":       {},
	"http://localhost:5173":       {},
	"https://dcd3m6lwt0z8u.cloudfront.net": {},
	"http://fre-style-bucket.s3-website-ap-northeast-1.amazonaws.com": {},
}

const (
	allowMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
	allowHeaders = "Content-Type, Authorization, X-Requested-With"
	maxAge       = "3600"
)

// IsAllowedOrigin は WebSocket Upgrader.CheckOrigin など CORS middleware の外でも
// 同じ allowlist を使えるようにするヘルパ。
func IsAllowedOrigin(origin string) bool {
	_, ok := allowedOrigins[origin]
	return ok
}

// CORS は Gin 用の CORS middleware。
// 許可リストにある Origin のみ Access-Control-Allow-Origin を返し、
// allowCredentials=true で動作するようにする。
// Preflight (OPTIONS) は本 middleware 内で 204 で返して終端する。
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
