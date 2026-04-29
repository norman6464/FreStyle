package middleware

import (
	"encoding/base64"
	"encoding/json"
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

// AdminGroupName は Cognito User Pool 上の admin グループ名。
// resjimkalto89890@gmail.com など管理者ユーザーが所属する。
// AWS Cognito の group 名は case-sensitive。Pool 2 (ap-northeast-1_TkRen4lyD) の
// 既存 group `admin` (小文字) を Spring Boot 時代から踏襲している。
const AdminGroupName = "admin"

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
		// cognito:groups は []string として claim に入る。Spring Boot 時代と同じ
		// "ADMIN" group を見て管理者判定する想定。
		if raw, ok := claims["cognito:groups"]; ok {
			groups := ToStringSliceFromClaim(raw)
			c.Set(ContextKeyCognitoGroups, groups)
		}
		c.Next()
	}
}

// ToStringSliceFromClaim は claim の `cognito:groups` を []string に変換する。
// JSON unmarshal 結果が []any なので逐次 string assert する。
// auth_handler 等の外部からも使うため exported。
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
