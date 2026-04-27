package domain

import "time"

type Notification struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"column:user_id;index" json:"userId"`
	Type      string    `gorm:"column:type" json:"type"`
	Title     string    `gorm:"column:title" json:"title"`
	Body      string    `gorm:"column:body" json:"body"`
	IsRead    bool      `gorm:"column:is_read" json:"isRead"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (Notification) TableName() string { return "notifications" }
