package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerWebSocketRoutes は AI チャットの WebSocket エンドポイントを登録する。
// SockJS / STOMP は廃止し、フロントはブラウザ標準 WebSocket で接続する。
// 認証は親 group の JWTAuth + CurrentUser ミドルウェアを継承する。
func registerWebSocketRoutes(g *gin.RouterGroup, deps *routeDeps) {
	sessionRepo := repository.NewAiChatSessionRepository(deps.db)
	sendMsg := usecase.NewSendAiMessageUseCase(sessionRepo, deps.msgRepo, deps.bedrockClient)
	g.GET("/ws/ai-chat", NewAiChatWsHandler(sendMsg).Handle)
}
