package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/infra/oidc"
)

// VerifyFunc は access_token を検証して claims を返す関数。
// 本番は infra/oidc.Verifier.Verify（JWKS 署名検証）を注入する。
type VerifyFunc func(ctx context.Context, token string) (map[string]any, error)

const (
	// ContextKeySubject は発行者が付けた本人の識別子（sub）。
	ContextKeySubject = "subject"
	// ContextKeyRoles は発行者側の役割の一覧。
	ContextKeyRoles   = "roles"
	CookieAccessToken = "access_token"
)

// JWTAuth は HttpOnly Cookie の access_token を verify で検証する Gin middleware。
//
// rolesClaim は役割の一覧が入っているクレーム名。発行者ごとに違うので設定から渡す。
func JWTAuth(verify VerifyFunc, rolesClaim string) gin.HandlerFunc {
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
		if rolesClaim != "" {
			c.Set(ContextKeyRoles, oidc.RolesFromClaim(claims[rolesClaim]))
		}
		c.Next()
	}
}

// RolesFromContext は context にセットされた役割の一覧を返す。
// 未設定 / 不正型の場合は nil。
func RolesFromContext(c *gin.Context) []string {
	v, ok := c.Get(ContextKeyRoles)
	if !ok {
		return nil
	}
	roles, _ := v.([]string)
	return roles
}
