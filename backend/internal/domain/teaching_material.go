package domain

import "time"

// TeachingMaterial は company_admin が自社 trainee 向けに作成する教材。
//
// アクセス制御:
//   - same company の company_admin: 自社の教材を全件 list / 編集 / 削除可
//   - same company の trainee: 自社の `is_published=true` 教材のみ閲覧可
//   - super_admin: 横断的に閲覧可（運用上の監査）
//
// 本文 (`Content`) は raw Markdown 文字列で保存し、 フロントは Edit / Preview タブで
// 表示する（Note と同じ形式）。
type TeachingMaterial struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID       uint64    `gorm:"column:company_id;not null;index" json:"companyId"`
	CreatedByUserID uint64    `gorm:"column:created_by_user_id;not null" json:"createdByUserId"`
	Title           string    `gorm:"column:title;not null;default:''" json:"title"`
	Content         string    `gorm:"column:content;type:text;not null;default:''" json:"content"`
	IsPublished     bool      `gorm:"column:is_published;not null;default:false" json:"isPublished"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (TeachingMaterial) TableName() string { return "teaching_materials" }
