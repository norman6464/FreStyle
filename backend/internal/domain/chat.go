package domain

import "time"

// ChatRoom は 1 対 1 / グループのチャットルーム。
type ChatRoom struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	IsGroup   bool      `gorm:"column:is_group" json:"isGroup"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (ChatRoom) TableName() string { return "chat_rooms" }

// ChatRoomMember はルームに所属するユーザー。
type ChatRoomMember struct {
	ID       uint64 `gorm:"primaryKey" json:"id"`
	RoomID   uint64 `gorm:"column:room_id;index" json:"roomId"`
	UserID   uint64 `gorm:"column:user_id;index" json:"userId"`
	JoinedAt time.Time `gorm:"column:joined_at" json:"joinedAt"`
}

func (ChatRoomMember) TableName() string { return "chat_room_members" }

// ChatMessage は DynamoDB に保存されるユーザー間メッセージ。
type ChatMessage struct {
	RoomID    uint64    `json:"roomId"`
	MessageID string    `json:"messageId"`
	SenderID  uint64    `json:"senderId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
