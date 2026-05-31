package middleware

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyCognitoSub    = "cognitoSub"
	ContextKeyEmail         = "email"
	ContextKeyCognitoGroups = "cognitoGroups"
	CookieAccessToken       = "access_token"
)

// AdminGroupName は Cognito User Pool 上の admin グループ名（case-sensitive）。
const AdminGroupName = "admin"

// JWTAuth は HttpOnly Cookie の access_token を検証する Gin middleware。
// 現状は payload の base64 デコードのみで、署名（JWKS）検証は別 issue。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(CookieAccessToken)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		claims, err := DecodeClaims(token)
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
		// cognito:groups は admin 判定に使う。
		if raw, ok := claims["cognito:groups"]; ok {
			groups := ToStringSliceFromClaim(raw)
			c.Set(ContextKeyCognitoGroups, groups)
		}
		c.Next()
	}
}

// ToStringSliceFromClaim は claim の cognito:groups を []string に変換する。
func ToStringSliceFromClaim(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// CognitoGroupsFromContext は context にセットされた cognito:groups を返す。
// 未設定 / 不正型の場合は nil。
func CognitoGroupsFromContext(c *gin.Context) []string {
	v, ok := c.Get(ContextKeyCognitoGroups)
	if !ok {
		return nil
	}
	groups, _ := v.([]string)
	return groups
}

// IsAdminFromGroups は groups に AdminGroupName が含まれているかを判定する。
func IsAdminFromGroups(groups []string) bool {
	return slices.Contains(groups, AdminGroupName)
}

// DecodeClaims は JWT の payload 部を base64url デコードして claim マップに変換する（署名検証はしない）。
func DecodeClaims(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidJWT
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

// base64URLDecode は JWT で使われる URL-safe base64 (パディング省略) を復元してデコードする。
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

// ErrInvalidJWT は token の形式 (3 セグメント) が壊れているときに返る。
var ErrInvalidJWT = errors.New("middleware: invalid jwt format")
