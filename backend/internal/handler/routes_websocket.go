package handler

import (
	"github.com/gin-gonic/gin"
)

// registerWebSocketRoutes は AI チャット / ユーザー間チャットの WebSocket エンドポイントを登録する。
// SockJS / STOMP は廃止し、フロントは ブラウザ標準 `WebSocket` で接続する。
//
// 認証は親 group の middleware.JWTAuth + middleware.CurrentUser を継承する
// （Cookie の access_token をそのまま WebSocket upgrade リクエストでも検証）。
func registerWebSocketRoutes(g *gin.RouterGroup) {
	// AiChat WebSocket (raw / native, Bedrock 連携は別 PR)
	aiChatWsHandler := NewAiChatWsHandler()
	g.GET("/ws/ai-chat", aiChatWsHandler.Handle)

	// Chat WebSocket (ルームごとブロードキャスト)。raw WebSocket + JSON プロトコル。
	chatWsHandler := NewChatWsHandler()
	g.GET("/ws/chat/:roomId", chatWsHandler.Handle)
}
