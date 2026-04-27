package domain

import "time"

// Note は学習メモを表す。Spring Boot の entity.Note に相当。
type Note struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"column:user_id;index" json:"userId"`
	Title     string    `gorm:"column:title" json:"title"`
	Content   string    `gorm:"column:content" json:"content"`
	IsPublic  bool      `gorm:"column:is_public" json:"isPublic"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Note) TableName() string { return "notes" }
