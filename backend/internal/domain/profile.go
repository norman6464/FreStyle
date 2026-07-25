package domain

import "time"

// Profile は users とは別管理のプロフィール拡張情報。
type Profile struct {
	UserID    uint64 `gorm:"primaryKey;column:user_id" json:"userId"`
	Bio       string `gorm:"column:bio" json:"bio"`
	AvatarURL string `gorm:"column:avatar_url" json:"avatarUrl"`
	// StatusMessage はプロフィールの一言ステータス(自由文)。状態機械の status 列とは別物。
	// JSON キーは互換のため status のまま(API 契約は不変)。
	StatusMessage string    `gorm:"column:status_message;default:''" json:"status"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Profile) TableName() string { return "profiles" }

// ProfileView は users.name と Profile を合成した表示用 DTO。
// Name の JSON キーはフロント (entities/user の Profile 型) に合わせて displayName
// (name だと取得・更新の両方向でフロントと食い違い氏名が消える: FRESTYLE-198)。
type ProfileView struct {
	UserID        uint64    `json:"userId"`
	Name          string    `json:"displayName"`
	Email         string    `json:"email"`
	Bio           string    `json:"bio"`
	AvatarURL     string    `json:"avatarUrl"`
	StatusMessage string    `json:"status"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
