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
	)
	g.GET("/ai-chat/sessions", aiHandler.GetSessions)
	g.POST("/ai-chat/sessions", aiHandler.CreateSession)
}
