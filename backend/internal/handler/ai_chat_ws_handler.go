package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// WebSocket keepalive 周りの定数。値の根拠:
//   - ALB / CloudFront のアイドルタイムアウト既定値は 60 秒。pingPeriod=54s でその手前に server → client ping を打つ
//   - pongWait は ALB の 60 秒に合わせ、それを超えても pong が来なければ切断する
//   - writeWait は遅い回線で 10 秒。これを超えても書き込めなければ切断
//
// WebSocket は単一接続を複数 goroutine から書き込めない（gorilla/websocket の制約）ため、
// すべての書き込みを 1 つの writer goroutine に集約する。
const (
	aiChatWriteWait      = 10 * time.Second
	aiChatPongWait       = 60 * time.Second
	aiChatPingPeriod     = (aiChatPongWait * 9) / 10
	aiChatMaxMessageSize = 64 * 1024
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
	// Bedrock / DynamoDB の初期化に失敗していると sendMsg は nil。upgrade 前に弾いて
	// クライアントに 503 を返す（接続だけ確立して即無言切断するより明示的）。
	if h.sendMsg == nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "ai_chat_unavailable"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("AiChat ws upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	conn.SetReadLimit(aiChatMaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(aiChatPongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(aiChatPongWait))
		return nil
	})

	// gin の c.Request.Context() は handler return で cancel される。WebSocket は
	// 長時間生き続けるため、独立した context を ws のライフサイクルに紐付ける。
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writes := make(chan any, 16)
	go h.writePump(conn, writes, cancel)

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

		// usecase 呼び出しに ws の context を渡す。Bedrock 呼び出しが詰まっても
		// 接続切断時に cancel されて goroutine が漏れない。
		out, err := h.sendMsg.Execute(ctx, usecase.SendAiMessageInput{
			UserID:      uid,
			SessionID:   sessionID,
			Content:     msg.Content,
			Scene:       msg.Scene,
			SessionType: msg.SessionType,
			ScenarioID:  msg.ScenarioID,
		})
		if err != nil {
			log.Printf("AiChat send message failed: %v", err)
			enqueueWS(writes, map[string]string{"type": "error", "message": "メッセージの送信に失敗しました"})
			continue
		}

		if out.NewSession != nil {
			enqueueWS(writes, wsOutSession{
				Type:        "session",
				ID:          out.NewSession.ID,
				Title:       out.NewSession.Title,
				SessionType: out.NewSession.SessionType,
				ScenarioID:  out.NewSession.ScenarioID,
				CreatedAt:   out.NewSession.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		enqueueWS(writes, wsOutMessage{
			Type:      "message",
			SessionID: out.AiMsg.SessionID,
			ID:        out.AiMsg.MessageID,
			Content:   out.AiMsg.Content,
			Role:      out.AiMsg.Role,
			CreatedAt: out.AiMsg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

// writePump は ws への書き込みを単一 goroutine に集約する。
//   - pingPeriod ごとに server → client へ ping を送り、ALB のアイドルタイムアウトを回避
//   - writes チャネルから受け取った payload を JSON で送信
//   - 書き込み失敗 / writes クローズで cancel を呼んで read ループも終了させる
func (h *AiChatWsHandler) writePump(conn *websocket.Conn, writes <-chan any, cancel context.CancelFunc) {
	ticker := time.NewTicker(aiChatPingPeriod)
	defer func() {
		ticker.Stop()
		cancel()
	}()

	for {
		select {
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(aiChatWriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case payload, ok := <-writes:
			if !ok {
				return
			}
			b, err := json.Marshal(payload)
			if err != nil {
				log.Printf("AiChat ws marshal failed: %v", err)
				continue
			}
			_ = conn.SetWriteDeadline(time.Now().Add(aiChatWriteWait))
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return
			}
		}
	}
}

// enqueueWS は writes チャネルへ非ブロッキング送信する。
// 送信先がバッファ満杯ならドロップしてログに残す（外部要因で詰まったときに read ループを止めない）。
func enqueueWS(writes chan<- any, payload any) {
	select {
	case writes <- payload:
	default:
		log.Printf("AiChat ws writes channel full — dropping payload: %T", payload)
	}
}
