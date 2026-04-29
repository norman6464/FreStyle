package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerChatRoutes は AI チャットセッションとユーザー間チャットルームの REST エンドポイントを登録する。
// WebSocket は registerWebSocketRoutes 側で別途登録する。
func registerChatRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Phase 3: AI チャット
	aiSessionRepo := repository.NewAiChatSessionRepository(deps.db)
	aiHandler := NewAiChatHandler(
		usecase.NewGetAiChatSessionsByUserIDUseCase(aiSessionRepo),
		usecase.NewCreateAiChatSessionUseCase(aiSessionRepo),
	)
	g.GET("/ai-chat/sessions", aiHandler.GetSessions)
	g.POST("/ai-chat/sessions", aiHandler.CreateSession)

	// Phase 4: ユーザー間チャット (ルーム CRUD)
	chatRoomRepo := repository.NewChatRoomRepository(deps.db)
	chatHandler := NewChatHandler(
		usecase.NewGetChatRoomsByUserIDUseCase(chatRoomRepo),
		usecase.NewCreateChatRoomUseCase(chatRoomRepo),
	)
	g.GET("/chat/rooms", chatHandler.GetRooms)
	g.POST("/chat/rooms", chatHandler.CreateRoom)
}
