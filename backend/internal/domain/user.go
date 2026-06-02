package domain

import "time"

// User はアプリケーション利用者のドメインモデル。
type User struct {
	ID          uint64  `gorm:"primaryKey" json:"id"`
	CognitoSub  string  `gorm:"column:cognito_sub;uniqueIndex" json:"cognitoSub"`
	Email       string  `gorm:"column:email" json:"email"`
	DisplayName string  `gorm:"column:display_name" json:"displayName"`
	CompanyID   *uint64 `gorm:"column:company_id" json:"companyId,omitempty"`
	Role        string  `gorm:"column:role" json:"role"`
	// OnboardedAt は Welcome 完了日時。NULL なら Welcome を表示する。一度入ったら変えない。
	OnboardedAt *time.Time `gorm:"column:onboarded_at" json:"onboardedAt,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"column:deleted_at" json:"deletedAt,omitempty"`
}

func (User) TableName() string { return "users" }

const (
	RoleSuperAdmin   = "super_admin"
	RoleCompanyAdmin = "company_admin"
	RoleTrainee      = "trainee"
)
