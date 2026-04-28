package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
)

// AiChatWsHandler は AI チャット用 WebSocket。
// 現状は user 認証 + echo の最小実装。Bedrock 連携は別 PR。
// SockJS / STOMP には依存せず、native WebSocket + JSON でフロントと通信する。
type AiChatWsHandler struct {
	upgrader websocket.Upgrader
}

func NewAiChatWsHandler() *AiChatWsHandler {
	return &AiChatWsHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return middleware.IsAllowedOrigin(r.Header.Get("Origin"))
			},
		},
	}
}

func (h *AiChatWsHandler) Handle(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("AiChat ws upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		// Bedrock 連携前のスケルトン: クライアント発話をそのまま echo する。
		// プロトコル本体は ws_protocol.go と同じ形式に揃える予定。
		if err := conn.WriteMessage(mt, msg); err != nil {
			return
		}
	}
}
