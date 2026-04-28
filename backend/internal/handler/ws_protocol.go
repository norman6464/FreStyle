package handler

import (
	"encoding/json"
	"time"
)

// ChatInbound はクライアントから WebSocket 経由で送られるメッセージの正規形。
// senderId / roomId / id / createdAt はクライアント任意値を信用しないためサーバが埋める。
// Spring Boot 時代の STOMP destination ベースから純 JSON ベースへ移行。
type ChatInbound struct {
	Type    string `json:"type"` // "send" | "delete"
	Content string `json:"content,omitempty"`
	// CreatedAtRef は delete 時にどのメッセージを論理削除するかを指す。
	// メッセージ id を sender 側で持っている場合のみ送信する。
	CreatedAtRef string `json:"createdAtRef,omitempty"`
}

// ChatOutbound はサーバが broadcast するメッセージの正規形。
type ChatOutbound struct {
	Type       string `json:"type"`             // "message" | "delete"
	ID         string `json:"id,omitempty"`     // server 生成
	RoomID     string `json:"roomId,omitempty"` // path 由来
	SenderID   uint64 `json:"senderId,omitempty"`
	SenderName string `json:"senderName,omitempty"`
	Content    string `json:"content,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"` // RFC3339
}

// BuildChatMessage は Inbound の "send" に対して Outbound の "message" を組み立てる。
// 純粋関数化することでテストしやすい状態にする。
func BuildChatMessage(in ChatInbound, roomID string, senderID uint64, senderName string, now time.Time) (ChatOutbound, bool) {
	if in.Type != "send" || in.Content == "" {
		return ChatOutbound{}, false
	}
	createdAt := now.UTC().Format(time.RFC3339Nano)
	return ChatOutbound{
		Type:       "message",
		ID:         createdAt + "-" + senderName, // 簡易識別子（永続化導入時に ULID へ移行）
		RoomID:     roomID,
		SenderID:   senderID,
		SenderName: senderName,
		Content:    in.Content,
		CreatedAt:  createdAt,
	}, true
}

// BuildChatDelete は Inbound の "delete" に対して Outbound の "delete" を組み立てる。
func BuildChatDelete(in ChatInbound, roomID string) (ChatOutbound, bool) {
	if in.Type != "delete" || in.CreatedAtRef == "" {
		return ChatOutbound{}, false
	}
	return ChatOutbound{
		Type:      "delete",
		RoomID:    roomID,
		CreatedAt: in.CreatedAtRef,
	}, true
}

// EncodeOutbound は Outbound を JSON バイト列に変換する。
func EncodeOutbound(out ChatOutbound) ([]byte, error) {
	return json.Marshal(out)
}

// DecodeInbound は WS frame の生バイト列を Inbound に変換する。
func DecodeInbound(raw []byte) (ChatInbound, error) {
	var in ChatInbound
	if err := json.Unmarshal(raw, &in); err != nil {
		return ChatInbound{}, err
	}
	return in, nil
}
