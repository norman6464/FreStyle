package domain

import "time"

type AdminInvitation struct {
	ID          uint64 `gorm:"primaryKey" json:"id"`
	CompanyID   uint64 `gorm:"column:company_id;index" json:"companyId"`
	Email       string `gorm:"column:email" json:"email"`
	Role        string `gorm:"column:role" json:"role"`
	DisplayName string `gorm:"column:display_name" json:"displayName"`
	Status      string `gorm:"column:status" json:"status"`
	// Token はマジックリンク用の不透明 UUID。SES で送信するメールに ?token=<UUID> として埋め込み、
	// 受諾画面で照合する。漏洩時の被害局所化のため一招待につき 1 値・受諾後は無効化（status=accepted で参照不可に）。
	// json では返さない（管理 UI でも露出させない方針。受諾フローでのみ使う秘匿値）。
	//
	// *string にしている理由: GORM AutoMigrate で UNIQUE INDEX を張るとき、既存 pending 行は
	// token がまだ無い（後続 PR-B で backfill 予定）。MariaDB の UNIQUE 制約は空文字 '' 同士は
	// 重複扱いになるが NULL 同士は重複扱いにならないため、未設定値は NULL にする必要がある。
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
