package handler

import (
	"net/http"

	"github.com/norman6464/FreStyle/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
)

// actorWorkspaceFromContext は middleware が注入した current user から
// (userID, workspace) を取り出す。未認証なら 401 を書き込んで ok=false を返すので、
// 呼び出し側は ok を見て早期 return する。各 handler が同じ「user 取得 + 401」を書かずに
// 済むための共通小道具。workspace は未所属(workspace_id = NULL)を表せる
// domain.WorkspaceRef で、空文字には潰さない。
func actorWorkspaceFromContext(c *gin.Context) (userID uint64, workspace domain.WorkspaceRef, ok bool) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, domain.NoWorkspace(), false
	}
	return user.ID, user.WorkspaceRef(), true
}
