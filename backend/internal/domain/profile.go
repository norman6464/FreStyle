package domain

import "time"

// Profile は users テーブルとは別管理のプロフィール拡張情報。
type Profile struct {
	UserID         uint64    `gorm:"primaryKey;column:user_id" json:"userId"`
	Bio            string    `gorm:"column:bio" json:"bio"`
	AvatarURL      string    `gorm:"column:avatar_url" json:"avatarUrl"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Profile) TableName() string { return "profiles" }
