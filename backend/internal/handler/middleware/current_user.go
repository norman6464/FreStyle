package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

const (
	ContextKeyCurrentUserID = "currentUserID"
	// ContextKeyCurrentUser は handler が role / company_id を見るための *domain.User。
	ContextKeyCurrentUser = "currentUser"
)

// CurrentUser は cognito sub から users 行を引いて currentUserID / currentUser を context にセットする。
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
		c.Set(ContextKeyCurrentUser, user)
		c.Next()
	}
}

// CurrentUserIDOrZero は CurrentUser がセットした users.id を返す（未設定なら 0）。
func CurrentUserIDOrZero(c *gin.Context) uint64 {
	v, ok := c.Get(ContextKeyCurrentUserID)
	if !ok {
		return 0
	}
	id, _ := v.(uint64)
	return id
}

// CurrentUserFromContext は CurrentUser が保存した *domain.User を返す（未セット時は nil）。
func CurrentUserFromContext(c *gin.Context) *domain.User {
	v, ok := c.Get(ContextKeyCurrentUser)
	if !ok {
		return nil
	}
	u, _ := v.(*domain.User)
	return u
}
