package domain

import "time"

// Course は教材を束ねる「コース（プロジェクト）」。
//
// 階層: Company 1 ── * Course 1 ── * TeachingMaterial
//
// アクセス制御:
//   - same company の company_admin: 自社のコースを全件 list / 編集 / 削除可
//   - same company の trainee: 自社の `is_published=true` コースのみ閲覧可
//   - super_admin: 横断的に閲覧可（運用上の監査）
//
// 並び順は `SortOrder` で明示的に指定する（同値時は ID 昇順）。
type Course struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID       uint64    `gorm:"column:company_id;not null;index" json:"companyId"`
	CreatedByUserID uint64    `gorm:"column:created_by_user_id;not null" json:"createdByUserId"`
	Title           string    `gorm:"column:title;not null;default:''" json:"title"`
	Description     string    `gorm:"column:description;type:text;not null;default:''" json:"description"`
	SortOrder       int       `gorm:"column:sort_order;not null;default:100" json:"sortOrder"`
	IsPublished     bool      `gorm:"column:is_published;not null;default:false" json:"isPublished"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Course) TableName() string { return "courses" }
