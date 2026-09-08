package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VerifyFunc は access_token を検証して claims を返す関数。
// 本番は infra/oidc.Verifier.Verify（JWKS 署名検証）を注入する。
type VerifyFunc func(ctx context.Context, token string) (map[string]any, error)

const (
	// ContextKeySubject は発行者が付けた本人の識別子（sub）。
	ContextKeySubject = "subject"
	CookieAccessToken = "access_token"
)

// JWTAuth は HttpOnly Cookie の access_token を verify で検証する Gin middleware。
func JWTAuth(verify VerifyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(CookieAccessToken)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		claims, err := verify(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		sub, _ := claims["sub"].(string)
		if sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_sub"})
			return
		}
		c.Set(ContextKeySubject, sub)
		c.Next()
	}
}
