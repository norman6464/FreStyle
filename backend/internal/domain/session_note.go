package domain

import "time"

// SessionNote は AI チャットセッション固有のメモ。
type SessionNote struct {
	ID        uint64    `json:"id"`
	SessionID uint64    `json:"sessionId"`
	UserID    uint64    `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
