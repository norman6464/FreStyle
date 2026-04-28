package domain

import "time"

// Profile は users テーブルとは別管理のプロフィール拡張情報。
//
// Status はユーザーが任意で公開できるショートメッセージ（例: 「学習中」「取り込み中」）。
// users.display_name は別テーブル管理だが、フロントは「プロフィール表示」で両者を
// 合わせて使うため、handler 層で join して返す前提。
type Profile struct {
	UserID    uint64    `gorm:"primaryKey;column:user_id" json:"userId"`
	Bio       string    `gorm:"column:bio" json:"bio"`
	AvatarURL string    `gorm:"column:avatar_url" json:"avatarUrl"`
	Status    string    `gorm:"column:status;default:''" json:"status"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Profile) TableName() string { return "profiles" }

// ProfileView は handler 層で users.display_name と Profile を合成した DTO。
// フロントの「プロフィール表示」が期待する単一オブジェクトの形。
type ProfileView struct {
	UserID      uint64    `json:"userId"`
	DisplayName string    `json:"displayName"`
	Bio         string    `json:"bio"`
	AvatarURL   string    `json:"avatarUrl"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
