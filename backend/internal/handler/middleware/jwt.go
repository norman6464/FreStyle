package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// VerifyFunc は access_token を検証して claims を返す関数。
// 本番は infra/cognito.Verifier.Verify（JWKS 署名検証）を注入する。
type VerifyFunc func(ctx context.Context, token string) (map[string]any, error)

const (
	ContextKeyCognitoSub    = "cognitoSub"
	ContextKeyEmail         = "email"
	ContextKeyCognitoGroups = "cognitoGroups"
	CookieAccessToken       = "access_token"
)

// AdminGroupName は Cognito User Pool 上の admin グループ名（case-sensitive）。
// 正本は domain 側（運営権限の規則がそこにあるため）。ここは既存の呼び出しのための別名。
const AdminGroupName = domain.CognitoAdminGroupName

// JWTAuth は HttpOnly Cookie の access_token を verify で検証する Gin middleware。
// verify は JWKS 署名検証を行う関数を注入する（偽造トークンを弾く）。
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
		c.Set(ContextKeyCognitoSub, sub)
		if email, ok := claims["email"].(string); ok {
			c.Set(ContextKeyEmail, email)
		}
		// cognito:groups は admin 判定に使う。配列として読めたときだけ置く。
		// 「置いてある」ことが claim の存在の印になり、運営権限の失効判定はそれを見る
		// （読めない形の claim を「グループに居ない」と誤読して権限を剥がさないため）。
		if groups := ToStringSliceFromClaim(claims["cognito:groups"]); groups != nil {
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
	groups, _ := CognitoGroupsClaimFromContext(c)
	return groups
}

// CognitoGroupsClaimFromContext は cognito:groups と、その claim が token に「在ったか」を返す。
// JWTAuth は claim キーが存在したときだけ context へ置くので、present はキーの有無そのもの。
// 運営権限の失効判定は present を必ず見ること（claim 欠落を「グループに居ない」と読むと、
// groups claim が載らない federated ユーザーの権限を誤って剥がす）。
func CognitoGroupsClaimFromContext(c *gin.Context) (groups []string, present bool) {
	v, ok := c.Get(ContextKeyCognitoGroups)
	if !ok {
		return nil, false
	}
	groups, _ = v.([]string)
	return groups, true
}

// PlatformAdminClaimFromContext は JWTAuth が置いた cognito:groups を運営権限の事実へ畳む。
func PlatformAdminClaimFromContext(c *gin.Context) domain.PlatformAdminClaim {
	groups, present := CognitoGroupsClaimFromContext(c)
	return domain.PlatformAdminFromGroups(present, groups)
}

// PlatformAdminClaimFromClaims は decode 済み claim マップを運営権限の事実へ畳む。
// id_token / access_token のどちらからでも同じ規則で読むための唯一の入口。
// キーが無い、または配列として読めない場合は Absent（何も判断しない）。
func PlatformAdminClaimFromClaims(claims map[string]any) domain.PlatformAdminClaim {
	groups := ToStringSliceFromClaim(claims["cognito:groups"])
	return domain.PlatformAdminFromGroups(groups != nil, groups)
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
