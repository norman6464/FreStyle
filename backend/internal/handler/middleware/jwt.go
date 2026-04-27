package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyCognitoSub = "cognitoSub"
	ContextKeyEmail      = "email"
	CookieAccessToken    = "access_token"
)

// JWTAuth は HttpOnly Cookie の access_token を検証する Gin middleware。
// Phase 2 ではトークン存在チェックのみ行い、JWKS 検証は Phase 2.1 (別 PR) で実装する。
// TODO: AWS Cognito JWKS から公開鍵を取得して署名検証する。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(CookieAccessToken)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		// プレースホルダ: 実際の検証は後続 PR で実装する
		c.Set(ContextKeyCognitoSub, "placeholder-sub")
		c.Next()
	}
}
