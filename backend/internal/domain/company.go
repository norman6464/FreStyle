package domain

import "time"

type Company struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	// AiChatEnabledForTrainees は自社 trainee に AI チャットを許可するか（既定 true）。
	// company_admin / super_admin が /company/settings で切り替える。
	AiChatEnabledForTrainees bool `json:"aiChatEnabledForTrainees"`
	// IsActive は会社アカウントの有効/無効。false（無効）にすると、その会社の全ユーザーが
	// ログイン/利用不可になる（middleware で弾く）。super_admin が会社一覧から切り替える。
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
