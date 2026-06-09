package domain

import "time"

type Company struct {
	ID   uint64 `gorm:"primaryKey" json:"id"`
	Name string `gorm:"column:name;not null" json:"name"`
	// AiChatEnabledForTrainees は自社 trainee に AI チャットを許可するか（既定 true）。
	// company_admin / super_admin が /company/settings で切り替える。AutoMigrate が列を追加する。
	AiChatEnabledForTrainees bool      `gorm:"column:ai_chat_enabled_for_trainees;not null;default:true" json:"aiChatEnabledForTrainees"`
	CreatedAt                time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt                time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Company) TableName() string { return "companies" }
