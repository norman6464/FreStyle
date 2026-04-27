package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ChatWsHandler はユーザー間チャット用 WebSocket。
// Spring Boot の ChatWebSocketController に相当。
// Phase 28 ではルームごとの簡易ブロードキャスト機構を提供する。
// DynamoDB への永続化は Phase 28.1 で別 PR。
type ChatWsHandler struct {
	upgrader websocket.Upgrader

	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{}
}

func NewChatWsHandler() *ChatWsHandler {
	return &ChatWsHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		rooms: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (h *ChatWsHandler) Handle(c *gin.Context) {
	roomID := c.Param("roomId")
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Chat ws upgrade failed: %v", err)
		return
	}
	h.join(roomID, conn)
	defer h.leave(roomID, conn)

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		h.broadcast(roomID, mt, msg, conn)
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

func (h *ChatWsHandler) broadcast(roomID string, mt int, msg []byte, sender *websocket.Conn) {
	h.mu.RLock()
	peers := h.rooms[roomID]
	h.mu.RUnlock()
	for c := range peers {
		if c == sender {
			continue
		}
		_ = c.WriteMessage(mt, msg)
	}
}
