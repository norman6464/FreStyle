package handler

import (
	"github.com/gin-gonic/gin"
)

// registerWebSocketRoutes は AI チャットの WebSocket エンドポイントを登録する。
// SockJS / STOMP は廃止し、フロントは ブラウザ標準 `WebSocket` で接続する。
//
// 認証は親 group の middleware.JWTAuth + middleware.CurrentUser を継承する
// （Cookie の access_token をそのまま WebSocket upgrade リクエストでも検証）。
func registerWebSocketRoutes(g *gin.RouterGroup) {
	aiChatWsHandler := NewAiChatWsHandler()
	g.GET("/ws/ai-chat", aiChatWsHandler.Handle)
}
