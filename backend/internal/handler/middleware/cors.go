package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 本番フロントは Firebase Hosting（frestyle.dev がカスタムドメイン、frestyle-dev.web.app が
// 既定 URL、frestyle-dev.firebaseapp.com は GCIP の authDomain としても使われる既定ドメインで、
// 同じ内容が配信される）。旧 frestyle.jp（DNS 削除済み）・旧 CloudFront・旧 S3 website
// エンドポイントは AWS 撤去に伴い実在しないため、許可リストから外してある
// （到達不能な origin を残す意味が無い）。
var allowedOrigins = map[string]struct{}{
	"https://frestyle.dev":                 {},
	"https://frestyle-dev.web.app":         {},
	"https://frestyle-dev.firebaseapp.com": {},
	"http://localhost:5173":                {},
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
