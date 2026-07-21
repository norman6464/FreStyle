package domain

import "time"

// Profile は users とは別管理のプロフィール拡張情報。
type Profile struct {
	UserID    uint64 `gorm:"primaryKey;column:user_id" json:"userId"`
	Bio       string `gorm:"column:bio" json:"bio"`
	AvatarURL string `gorm:"column:avatar_url" json:"avatarUrl"`
	// Status は旧列。StatusMessage(新列・一言ステータス)へ移行中で、Expand フェーズでは
	// dual-write により両者を同値に保つ。読みは当面 Status を使い、Contract で削除する。
	Status        string    `gorm:"column:status;default:''" json:"status"`
	StatusMessage string    `gorm:"column:status_message;default:''" json:"-"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Profile) TableName() string { return "profiles" }

// ProfileView は users.name と Profile を合成した表示用 DTO。
type ProfileView struct {
	UserID    uint64    `json:"userId"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio"`
	AvatarURL string    `json:"avatarUrl"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}
