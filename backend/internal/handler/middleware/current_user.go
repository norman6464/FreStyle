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
// これにより handler は `middleware.CurrentUserIDOrZero(c)` で
// 認証済みユーザーの users.id を取り出せる。
// 個別 handler が `:userId` path / `?userId=` クエリを要求しなくても済む。
func CurrentUser(users repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(ContextKeyCognitoSub)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		sub, ok := raw.(string)
		if !ok || sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_sub"})
			return
		}
		user, err := users.FindByCognitoSub(c.Request.Context(), sub)
		if err != nil {
			// repo / DB エラーは認証問題ではなくサーバ側障害なので 500。
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user_lookup_failed"})
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_not_found"})
			return
		}
		c.Set(ContextKeyCurrentUserID, user.ID)
		c.Next()
	}
}

// CurrentUserIDOrZero は middleware.CurrentUser がセットした users.id を取り出す。
// 設定されていなければ 0 を返す（呼び出し側でガードする）。
// 名前で「未設定なら 0」と明示することで、Must プレフィックスから連想される
// 「不在なら panic / abort」挙動と取り違えないようにする意図。
func CurrentUserIDOrZero(c *gin.Context) uint64 {
	v, ok := c.Get(ContextKeyCurrentUserID)
	if !ok {
		return 0
	}
	id, _ := v.(uint64)
	return id
}
