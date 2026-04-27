package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// AiChatWsHandler は AI チャット用 WebSocket。
// Spring Boot の AiChatWebSocketController に相当。
// Phase 27 ではエコーサーバ + 接続管理のスケルトンのみ。
// Bedrock 連携は Phase 27.1 で別 PR。
type AiChatWsHandler struct {
	upgrader websocket.Upgrader
}

func NewAiChatWsHandler() *AiChatWsHandler {
	return &AiChatWsHandler{
		upgrader: websocket.Upgrader{
			// 本番では origin チェックを有効化する。
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *AiChatWsHandler) Handle(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("AiChat ws upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// 暫定: クライアントから受け取ったメッセージをそのまま返す（接続疎通確認用）
		if err := conn.WriteMessage(mt, msg); err != nil {
			break
		}
	}
}
