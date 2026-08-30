package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/norman6464/FreStyle/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
)

// actorWorkspaceFromContext は middleware が注入した current user から
// (userID, workspace, role) を取り出す。未認証なら 401 を書き込んで ok=false を返すので、
// 呼び出し側は ok を見て早期 return する。各 handler が同じ「user 取得 + 401」を書かずに
// 済むための共通小道具。workspace は未所属(workspace_id = NULL)を表せる
// domain.WorkspaceRef で、空文字には潰さない。
func actorWorkspaceFromContext(c *gin.Context) (userID uint64, workspace domain.WorkspaceRef, role domain.RoleName, ok bool) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, domain.NoWorkspace(), "", false
	}
	return user.ID, user.WorkspaceRef(), user.Role, true
}

// respondEntityErr は usecase が返したエラーを HTTP ステータスへ振り分ける共通処理。
// レコード未検出は 404(notFoundMsg)、認可エラー(forbidden* / ワークスペース未所属)は 403、
// それ以外は 500(fallback) を返す。エンティティ別の文言は notFoundMsg / fallback で渡す。
func respondEntityErr(c *gin.Context, err error, notFoundMsg, fallback string) {
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": notFoundMsg})
		return
	}
	if strings.HasPrefix(err.Error(), "forbidden") || err.Error() == "actor must belong to a workspace" {
		c.JSON(http.StatusForbidden, gin.H{"error": "操作権限がありません"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
}
