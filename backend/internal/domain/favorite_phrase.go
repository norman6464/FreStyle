package domain

import "time"

type FavoritePhrase struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"column:user_id;index" json:"userId"`
	Phrase    string    `gorm:"column:phrase" json:"phrase"`
	Note      string    `gorm:"column:note" json:"note"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (FavoritePhrase) TableName() string { return "favorite_phrases" }
