package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyCognitoSub = "cognitoSub"
	ContextKeyEmail      = "email"
	CookieAccessToken    = "access_token"
)

// JWTAuth は HttpOnly Cookie の access_token を検証する Gin middleware。
// 現状は payload (claims) の base64 デコードのみで、署名 (JWKS) 検証は別 issue で実装する。
// ハンドラから c.Get(ContextKeyCognitoSub) で本物の Cognito sub を取れる。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(CookieAccessToken)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		claims, err := decodeClaims(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		sub, _ := claims["sub"].(string)
		if sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_sub"})
			return
		}
		c.Set(ContextKeyCognitoSub, sub)
		if email, ok := claims["email"].(string); ok {
			c.Set(ContextKeyEmail, email)
		}
		c.Next()
	}
}

func decodeClaims(jwt string) (map[string]any, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, errInvalidJWT
	}
	payload, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func base64URLDecode(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	s = strings.NewReplacer("-", "+", "_", "/").Replace(s)
	return base64.StdEncoding.DecodeString(s)
}

type jwtError string

func (e jwtError) Error() string { return string(e) }

const errInvalidJWT = jwtError("invalid jwt format")
