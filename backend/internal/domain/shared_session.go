package domain

import "time"

// SharedSession はコミュニティに共有された AI 会話セッションのメタデータ。
type SharedSession struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	OwnerID     uint64    `gorm:"column:owner_id;index" json:"ownerId"`
	SessionID   uint64    `gorm:"column:session_id;index" json:"sessionId"`
	Title       string    `gorm:"column:title" json:"title"`
	Description string    `gorm:"column:description" json:"description"`
	IsPublic    bool      `gorm:"column:is_public" json:"isPublic"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (SharedSession) TableName() string { return "shared_sessions" }
