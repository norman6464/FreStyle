package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AiChatWsHandler は AI チャット用 WebSocket エンドポイント。
// native WebSocket + JSON プロトコルで動作し、SockJS / STOMP には依存しない。
type AiChatWsHandler struct {
	upgrader websocket.Upgrader
	sendMsg  *usecase.SendAiMessageUseCase
}

func NewAiChatWsHandler(sendMsg *usecase.SendAiMessageUseCase) *AiChatWsHandler {
	return &AiChatWsHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return middleware.IsAllowedOrigin(r.Header.Get("Origin"))
			},
		},
		sendMsg: sendMsg,
	}
}

// wsInbound はフロントエンドから受け取るメッセージ形式。
type wsInbound struct {
	Type        string  `json:"type"`
	SessionID   *int64  `json:"sessionId"`
	Content     string  `json:"content"`
	Role        string  `json:"role"`
	Scene       string  `json:"scene"`
	SessionType string  `json:"sessionType"`
	ScenarioID  *uint64 `json:"scenarioId"`
}

// wsOutSession は新規セッション作成時にフロントへ返す形式。
type wsOutSession struct {
	Type        string  `json:"type"`
	ID          uint64  `json:"id"`
	Title       string  `json:"title"`
	SessionType string  `json:"sessionType,omitempty"`
	ScenarioID  *uint64 `json:"scenarioId,omitempty"`
	CreatedAt   string  `json:"createdAt,omitempty"`
}

// wsOutMessage はアシスタント返答をフロントへ返す形式。
type wsOutMessage struct {
	Type      string `json:"type"`
	SessionID uint64 `json:"sessionId"`
	ID        string `json:"id"`
	Content   string `json:"content"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
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
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg wsInbound
		if err := json.Unmarshal(raw, &msg); err != nil || msg.Type != "send" {
			continue
		}

		var sessionID uint64
		if msg.SessionID != nil && *msg.SessionID > 0 {
			sessionID = uint64(*msg.SessionID)
		}

		out, err := h.sendMsg.Execute(c.Request.Context(), usecase.SendAiMessageInput{
			UserID:      uid,
			SessionID:   sessionID,
			Content:     msg.Content,
			Scene:       msg.Scene,
			SessionType: msg.SessionType,
			ScenarioID:  msg.ScenarioID,
		})
		if err != nil {
			log.Printf("AiChat send message failed: %v", err)
			wsWriteJSON(conn, map[string]string{"type": "error", "message": "メッセージの送信に失敗しました"})
			continue
		}

		if out.NewSession != nil {
			wsWriteJSON(conn, wsOutSession{
				Type:        "session",
				ID:          out.NewSession.ID,
				Title:       out.NewSession.Title,
				SessionType: out.NewSession.SessionType,
				ScenarioID:  out.NewSession.ScenarioID,
				CreatedAt:   out.NewSession.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		wsWriteJSON(conn, wsOutMessage{
			Type:      "message",
			SessionID: out.AiMsg.SessionID,
			ID:        out.AiMsg.MessageID,
			Content:   out.AiMsg.Content,
			Role:      out.AiMsg.Role,
			CreatedAt: out.AiMsg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func wsWriteJSON(conn *websocket.Conn, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("AiChat ws marshal failed: %v", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
		log.Printf("AiChat ws write failed: %v", err)
	}
}
