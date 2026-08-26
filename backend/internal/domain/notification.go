package domain

import "time"

type Notification struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"userId"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}
