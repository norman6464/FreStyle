package domain

import "time"

// SessionNote は AI チャットセッション固有のメモ。
type SessionNote struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	SessionID uint64    `gorm:"column:session_id;index" json:"sessionId"`
	UserID    uint64    `gorm:"column:user_id;index" json:"userId"`
	Content   string    `gorm:"column:content" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (SessionNote) TableName() string { return "session_notes" }
