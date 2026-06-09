package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// RequireAiChatEnabled は ai-chat 系エンドポイントの入口で、会社設定により trainee の AI 利用が
// 無効化されていれば 403 で弾く横断ゲート。管理者・会社未所属は常に通す（判定は usecase に委譲）。
// 認可の前提として、先に CurrentUser ミドルウェアが *domain.User を context に入れている必要がある。
func RequireAiChatEnabled(checker *usecase.AiChatEnabledForUserUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get(ContextKeyCurrentUser)
		u, _ := user.(*domain.User)
		allowed, err := checker.Execute(c.Request.Context(), u)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "ai_chat_disabled_for_company"})
			return
		}
		c.Next()
	}
}
