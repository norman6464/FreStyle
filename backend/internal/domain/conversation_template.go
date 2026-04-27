package domain

import "time"

type ConversationTemplate struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"column:title" json:"title"`
	Body      string    `gorm:"column:body" json:"body"`
	Category  string    `gorm:"column:category" json:"category"`
	IsActive  bool      `gorm:"column:is_active" json:"isActive"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (ConversationTemplate) TableName() string { return "conversation_templates" }
