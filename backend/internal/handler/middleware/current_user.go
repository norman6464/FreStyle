package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

const (
	ContextKeyCurrentUserID = "currentUserID"
	// ContextKeyCurrentUser は handler が role / company_id を見るための *domain.User。
	ContextKeyCurrentUser = "currentUser"
)

// CurrentUser は cognito sub から users 行を引いて currentUserID / currentUser を context にセットする。
// 併せて、所属会社が無効化されている場合はその会社の全ユーザーを弾く（即時にログイン/利用不可）。
func CurrentUser(users repository.UserRepository, companies repository.CompanyRepository) gin.HandlerFunc {
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

		// ユーザーアカウントが無効化されていれば利用不可（有効な JWT でも即時に弾く）。
		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user_disabled"})
			return
		}

		// 会社アカウントが無効化されていれば、その会社のユーザーは利用不可。
		// 会社行が無い（データ不整合）場合は素通り、DB エラーは 500。
		if user.CompanyID != nil {
			company, err := companies.FindByID(c.Request.Context(), *user.CompanyID)
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				// 会社行なし: 何もしない（弾かない）。
			case err != nil:
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "company_lookup_failed"})
				return
			case company != nil && !company.IsActive:
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "company_disabled"})
				return
			}
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
