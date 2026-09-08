package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// VerifyFunc は ID トークンを検証して claims を返す関数。
// 本番は infra/oidc.Verifier.Verify（JWKS 署名検証）を注入する。
type VerifyFunc func(ctx context.Context, token string) (map[string]any, error)

const (
	// ContextKeySubject は発行者が付けた本人の識別子（sub）。
	ContextKeySubject = "subject"
)

// JWTAuth は Authorization: Bearer <ID トークン> を verify で検証する Gin middleware。
//
// GCIP はサーバー側で発行するセッション Cookie を持たない。クライアント SDK が保持する
// ID トークンをリクエストのたびに Bearer で送る前提で、backend はここで毎回検証する。
func JWTAuth(verify VerifyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := BearerToken(c.GetHeader("Authorization"))
		if !ok {
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

// BearerToken は "Bearer <token>" の形の Authorization ヘッダからトークン本体を取り出す。
// 形が違う・本体が空ならトークン無しとして扱う。
//
// JWTAuth（毎リクエストの検証）と、ログイン直後に一度だけ呼ばれる session 確立ハンドラの
// 両方が同じ形のヘッダを読むため、ここに切り出して共有する。
func BearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimPrefix(header, prefix)
	if token == "" {
		return "", false
	}
	return token, true
}
