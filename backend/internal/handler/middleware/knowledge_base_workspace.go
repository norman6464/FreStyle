package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

const (
	// ContextKeyKnowledgeBaseWorkspace は KnowledgeBaseWorkspace が入れる *domain.Workspace。
	ContextKeyKnowledgeBaseWorkspace = "kbWorkspace"
	// ParamKnowledgeBaseWorkspaceSlug はワークスペースを指す URL パラメータ名。
	// ルート登録側とこの middleware で同じ綴りを使うため定数にする。
	ParamKnowledgeBaseWorkspaceSlug = "workspaceSlug"
)

// KnowledgeBaseWorkspace はナレッジ基盤の各エンドポイントの入口でテナントを確定させる。
//
// workspace_id をリクエストボディやクエリで受け取らないのが要点。受け取る形にすると
// 「クライアントが指定した任意のテナントで処理する」経路が API 契約として固まってしまい、
// handler が 1 箇所検証を忘れた時点でテナント越えの読み書きが通る。URL の slug から引き直し、
// principals による所属確認まで済ませた *domain.Workspace だけを context に載せる。
//
// 存在しない slug と、存在するが所属していない slug は同じ 404 にする（判定は usecase 側で
// 潰してあり、ここでは撃ち分けようがない）。403 を返すと slug の総当たりで
// 他社のワークスペースの実在が分かってしまうため。
//
// 前提として、先に CurrentUser middleware が currentUserID を context に入れている必要がある。
func KnowledgeBaseWorkspace(resolve *usecase.ResolveWorkspaceUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := CurrentUserIDOrZero(c)
		if uid == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		ws, err := resolve.Execute(c.Request.Context(), usecase.ResolveWorkspaceInput{
			Slug:   c.Param(ParamKnowledgeBaseWorkspaceSlug),
			UserID: uid,
		})
		switch {
		case errors.Is(err, repository.ErrWorkspaceNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		case err != nil:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		c.Set(ContextKeyKnowledgeBaseWorkspace, ws)
		c.Next()
	}
}

// KnowledgeBaseWorkspaceFromContext は KnowledgeBaseWorkspace が保存した *domain.Workspace を返す
// （未セット時は nil）。
func KnowledgeBaseWorkspaceFromContext(c *gin.Context) *domain.Workspace {
	v, ok := c.Get(ContextKeyKnowledgeBaseWorkspace)
	if !ok {
		return nil
	}
	ws, _ := v.(*domain.Workspace)
	return ws
}

// KnowledgeBaseWorkspaceIDOrEmpty は確定済みワークスペースの ID を返す（未セットなら空文字）。
func KnowledgeBaseWorkspaceIDOrEmpty(c *gin.Context) string {
	if ws := KnowledgeBaseWorkspaceFromContext(c); ws != nil {
		return ws.ID
	}
	return ""
}
