package domain

import "time"

// CompanyApplication は未登録の企業担当者がログイン前の公開フォームから出す「利用申請」。
// super_admin が一覧で確認し、問題なければ既存の招待フローで company_admin を招待する。
type CompanyApplication struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyName   string    `gorm:"size:200;not null" json:"companyName"`
	ApplicantName string    `gorm:"size:120;not null" json:"applicantName"`
	Email         string    `gorm:"size:255;not null;index" json:"email"`
	Message       string    `gorm:"type:text" json:"message"`
	Status        string    `gorm:"size:16;not null;default:'pending';index" json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (CompanyApplication) TableName() string { return "company_applications" }

const (
	CompanyApplicationStatusPending  = "pending"
	CompanyApplicationStatusApproved = "approved"
	CompanyApplicationStatusRejected = "rejected"
)

// NotificationTypeCompanyApplication は企業申請が届いたことを super_admin に知らせる通知の Type。
const NotificationTypeCompanyApplication = "company_application"
