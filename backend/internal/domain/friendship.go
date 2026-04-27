package domain

import "time"

type Friendship struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	RequesterID uint64    `gorm:"column:requester_id;index" json:"requesterId"`
	AddresseeID uint64    `gorm:"column:addressee_id;index" json:"addresseeId"`
	Status      string    `gorm:"column:status" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (Friendship) TableName() string { return "friendships" }

const (
	FriendshipStatusPending  = "pending"
	FriendshipStatusAccepted = "accepted"
	FriendshipStatusRejected = "rejected"
)
