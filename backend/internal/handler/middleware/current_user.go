package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

const (
	ContextKeyCurrentUserID = "currentUserID"
	// ContextKeyCurrentUser は handler が role / workspace_id を見るための *domain.User。
	ContextKeyCurrentUser = "currentUser"
)

// CurrentUser は cognito sub から users 行を引いて currentUserID / currentUser を context にセットする。
// 併せて、所属ワークスペースが停止されている場合はその全員を弾く（即時に利用不可）。
func CurrentUser(users repository.UserRepository, workspaces repository.WorkspaceActivationReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(ContextKeySubject)
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

		// 所属ワークスペースが停止されていれば、そこに属する全員が利用不可。
		// 未所属のユーザーは検査対象が無いので素通りする。
		//
		// 行が見つからない場合は素通りさせない。users.workspace_id には FK
		// （fk_users_workspace）が張ってあるので、所属先の行は必ず存在するはず。
		// 無いのはデータ不整合であって「停止されていない」ことの証拠ではないため、
		// 弾く側に倒す（素通りにすると、FK が外れた瞬間に遮断が黙って効かなくなる）。
		if workspaceID, affiliated := user.WorkspaceRef().WorkspaceID(); affiliated {
			workspace, err := workspaces.FindWorkspaceByID(c.Request.Context(), workspaceID)
			switch {
			case err != nil:
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "workspace_lookup_failed"})
				return
			case workspace == nil || !workspace.IsActive:
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "workspace_disabled"})
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
