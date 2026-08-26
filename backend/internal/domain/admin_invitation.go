package domain

import "time"

type AdminInvitation struct {
	ID        uint64   `json:"id"`
	CompanyID uint64   `json:"companyId"`
	Email     string   `json:"email"`
	Role      RoleName `json:"role"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	// Token はマジックリンク用の不透明 UUID（秘匿値なので json では返さない）。
	// 未設定値を NULL にして UNIQUE 制約に引っかけないため *string にしている。
	Token     *string   `json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusCanceled = "canceled"
)
