package domain

import "time"

type AdminInvitation struct {
	ID          uint64 `gorm:"primaryKey" json:"id"`
	CompanyID   uint64 `gorm:"column:company_id;index" json:"companyId"`
	Email       string `gorm:"column:email" json:"email"`
	Role        string `gorm:"column:role" json:"role"`
	DisplayName string `gorm:"column:display_name" json:"displayName"`
	Status      string `gorm:"column:status" json:"status"`
	// Token はマジックリンク用の不透明 UUID（秘匿値なので json では返さない）。
	// 未設定値を NULL にして UNIQUE 制約に引っかけないため *string にしている。
	Token     *string   `gorm:"column:token;uniqueIndex;size:64" json:"-"`
	ExpiresAt time.Time `gorm:"column:expires_at" json:"expiresAt"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (AdminInvitation) TableName() string { return "invitations" }

const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusCanceled = "canceled"
)
