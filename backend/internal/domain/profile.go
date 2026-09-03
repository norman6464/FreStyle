package domain

import "time"

// Profile は users とは別管理のプロフィール拡張情報。
type Profile struct {
	UserID        uint64    `json:"userId"`
	Bio           string    `json:"bio"`
	AvatarURL     string    `json:"avatarUrl"`
	StatusMessage string    `json:"status"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ProfileView struct {
	UserID        uint64    `json:"userId"`
	Name          string    `json:"displayName"`
	Email         string    `json:"email"`
	Bio           string    `json:"bio"`
	AvatarURL     string    `json:"avatarUrl"`
	StatusMessage string    `json:"status"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
