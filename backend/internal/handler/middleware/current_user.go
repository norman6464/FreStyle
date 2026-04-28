package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

const ContextKeyCurrentUserID = "currentUserID"

// CurrentUser は JWTAuth の後段で動く middleware。
// JWTAuth が context に詰めた cognito sub から users 行を引き、
// `currentUserID` を context にセットする。
//
// これにより handler は `c.Get(ContextKeyCurrentUserID)` で
// 認証済みユーザーの users.id を取り出せる。
// 個別 handler が `:userId` path / `?userId=` クエリを要求しなくても済む。
func CurrentUser(users repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(ContextKeyCognitoSub)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		sub, _ := raw.(string)
		user, err := users.FindByCognitoSub(c.Request.Context(), sub)
		if err != nil || user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_not_found"})
			return
		}
		c.Set(ContextKeyCurrentUserID, user.ID)
		c.Next()
	}
}

// MustCurrentUserID は middleware.CurrentUser が必須の handler 内で使うヘルパ。
// 設定されていなければ 0 を返す（呼び出し側でガードする）。
func MustCurrentUserID(c *gin.Context) uint64 {
	v, ok := c.Get(ContextKeyCurrentUserID)
	if !ok {
		return 0
	}
	id, _ := v.(uint64)
	return id
}
