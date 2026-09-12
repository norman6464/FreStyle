package domain

import (
	"strings"
	"time"
)

// Profile は users とは別管理のプロフィール拡張情報。
//
// 一言ステータスは StatusEmoji / StatusText / StatusExpiresAt の 3 列（段 14）。
// 失効時刻を過ぎたステータスは、DB の値は残したまま応答からだけ空にする
// （ClearExpiredStatus 参照。UI 側で同じ内容を再設定する手間を増やさないため）。
type Profile struct {
	UserID          uint64     `json:"userId"`
	Bio             string     `json:"bio"`
	AvatarURL       string     `json:"avatarUrl"`
	StatusText      string     `json:"status"`
	StatusEmoji     string     `json:"statusEmoji"`
	StatusExpiresAt *time.Time `json:"statusExpiresAt,omitempty"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type ProfileView struct {
	UserID          uint64     `json:"userId"`
	Name            string     `json:"displayName"`
	Email           string     `json:"email"`
	Bio             string     `json:"bio"`
	AvatarURL       string     `json:"avatarUrl"`
	StatusText      string     `json:"status"`
	StatusEmoji     string     `json:"statusEmoji"`
	StatusExpiresAt *time.Time `json:"statusExpiresAt,omitempty"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// ClearExpiredStatus は失効時刻(expiresAt)を過ぎていれば絵文字・テキストを空文字にして
// 返す。DB の値そのものは変えない（呼び出し側は読み取り結果にだけ適用する）。
func ClearExpiredStatus(emoji, text string, expiresAt *time.Time, now time.Time) (string, string) {
	if expiresAt != nil && !now.Before(*expiresAt) {
		return "", ""
	}
	return emoji, text
}

// ComposeStatusDisplay は他ユーザー向けの一言ステータス欄（1 本の文字列）を組み立てる。
// 絵文字とテキストを結合し、失効していれば空文字を返す。
func ComposeStatusDisplay(emoji, text string, expiresAt *time.Time, now time.Time) string {
	emoji, text = ClearExpiredStatus(emoji, text, expiresAt, now)
	return strings.TrimSpace(emoji + " " + text)
}
