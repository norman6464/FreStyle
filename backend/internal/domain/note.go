package domain

import "time"

// Note は学習メモ。IsPinned と IsPublic は独立した属性。
type Note struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"userId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsPublic  bool      `json:"isPublic"`
	IsPinned  bool      `json:"isPinned"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
