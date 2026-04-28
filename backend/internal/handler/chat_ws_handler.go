package handler

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
)

// ChatWsHandler はユーザー間チャット用 WebSocket。
// raw WebSocket + JSON プロトコル。SockJS / STOMP は使わない。
//
// メッセージは {type, content?, createdAtRef?} の JSON で受信し、
// {type, id, roomId, senderId, senderName, content, createdAt} で broadcast する。
// senderId は JWT 由来 (middleware.CurrentUserIDOrZero) でサーバが固定する（IDOR 対策）。
type ChatWsHandler struct {
	upgrader websocket.Upgrader

	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{}
}

func NewChatWsHandler() *ChatWsHandler {
	return &ChatWsHandler{
		upgrader: websocket.Upgrader{
			// 本番 origin allowlist。SockJS 排除と同時に origin 検査を厳格化。
			CheckOrigin: func(r *http.Request) bool {
				return middleware.IsAllowedOrigin(r.Header.Get("Origin"))
			},
		},
		rooms: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (h *ChatWsHandler) Handle(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	roomID := c.Param("roomId")
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Chat ws upgrade failed: %v", err)
		return
	}
	h.join(roomID, conn)
	defer h.leave(roomID, conn)

	// senderName は JWT email を簡易的に使う。永続化導入時に users.display_name へ差し替え。
	senderName, _ := c.Get(middleware.ContextKeyEmail)
	senderNameStr, _ := senderName.(string)

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		in, err := DecodeInbound(raw)
		if err != nil {
			continue
		}
		switch in.Type {
		case "send":
			out, ok := BuildChatMessage(in, roomID, uid, senderNameStr, time.Now())
			if !ok {
				continue
			}
			h.broadcastJSON(roomID, out)
		case "delete":
			out, ok := BuildChatDelete(in, roomID)
			if !ok {
				continue
			}
			h.broadcastJSON(roomID, out)
		}
	}
}

func (h *ChatWsHandler) join(roomID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[roomID][conn] = struct{}{}
}

func (h *ChatWsHandler) leave(roomID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if peers, ok := h.rooms[roomID]; ok {
		delete(peers, conn)
		if len(peers) == 0 {
			delete(h.rooms, roomID)
		}
	}
	conn.Close()
}

func (h *ChatWsHandler) broadcastJSON(roomID string, out ChatOutbound) {
	raw, err := EncodeOutbound(out)
	if err != nil {
		return
	}
	h.mu.RLock()
	peers := h.rooms[roomID]
	conns := make([]*websocket.Conn, 0, len(peers))
	for c := range peers {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	// 送信者にも自分の発言を返す（フロントは subscribe で受け取って描画する設計）。
	for _, c := range conns {
		_ = c.WriteMessage(websocket.TextMessage, raw)
	}
}
