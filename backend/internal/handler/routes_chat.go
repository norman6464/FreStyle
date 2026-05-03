package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerChatRoutes は AI チャットセッションの REST エンドポイントを登録する。
// WebSocket は registerWebSocketRoutes 側で別途登録する。
func registerChatRoutes(g *gin.RouterGroup, deps *routeDeps) {
	aiSessionRepo := repository.NewAiChatSessionRepository(deps.db)
	aiHandler := NewAiChatHandler(
		usecase.NewGetAiChatSessionsByUserIDUseCase(aiSessionRepo),
		usecase.NewCreateAiChatSessionUseCase(aiSessionRepo),
		usecase.NewGetAiChatSessionUseCase(aiSessionRepo),
		usecase.NewUpdateAiChatSessionTitleUseCase(aiSessionRepo),
		usecase.NewDeleteAiChatSessionUseCase(aiSessionRepo),
		usecase.NewGetAiChatMessagesUseCase(deps.msgRepo),
	)
	g.GET("/ai-chat/sessions", aiHandler.GetSessions)
	g.POST("/ai-chat/sessions", aiHandler.CreateSession)
	g.GET("/ai-chat/sessions/:id", aiHandler.GetSession)
	g.PUT("/ai-chat/sessions/:id", aiHandler.UpdateSessionTitle)
	g.DELETE("/ai-chat/sessions/:id", aiHandler.DeleteSession)
	g.GET("/ai-chat/sessions/:id/messages", aiHandler.GetMessages)
}
