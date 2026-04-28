package domain

import "time"

// Note は学習メモを表す。フロント NotesPage / SessionNote 機能のドメインモデル。
//
// IsPinned はユーザーがピン留めしたかどうか（ピン留めは並び順で先頭に固定する用途）。
// IsPublic は他ユーザーに公開するかどうか。両者は独立した属性。
type Note struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"column:user_id;index" json:"userId"`
	Title     string    `gorm:"column:title" json:"title"`
	Content   string    `gorm:"column:content" json:"content"`
	IsPublic  bool      `gorm:"column:is_public" json:"isPublic"`
	IsPinned  bool      `gorm:"column:is_pinned;default:false" json:"isPinned"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Note) TableName() string { return "notes" }
