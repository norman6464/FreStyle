package domain

import "time"

// CompanyApplication は未登録の企業担当者がログイン前の公開フォームから出す「利用申請」。
// super_admin が一覧で確認し、問題なければ既存の招待フローで company_admin を招待する。
type CompanyApplication struct {
	ID            uint64    `json:"id"`
	CompanyName   string    `json:"companyName"`
	ApplicantName string    `json:"applicantName"`
	Email         string    `json:"email"`
	Message       string    `json:"message"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

const (
	CompanyApplicationStatusPending  = "pending"
	CompanyApplicationStatusApproved = "approved"
	CompanyApplicationStatusRejected = "rejected"
)

// NotificationTypeCompanyApplication は企業申請が届いたことを super_admin に知らせる通知の Type。
const NotificationTypeCompanyApplication = "company_application"
