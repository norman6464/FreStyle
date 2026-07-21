package domain

import "time"

// TeachingMaterial はコースを構成する「章」。必ず 1 つの Course に所属する（course 1 : N chapter）。
// テーブルは course_chapters（FRESTYLE-184 で teaching_materials から改名）。
// 本文は raw Markdown 文字列。コース内の並び順は sort_order 列（同値時 ID 昇順）。
type TeachingMaterial struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID uint64 `gorm:"column:company_id;not null;index" json:"companyId"`
	// NOT NULL は migration 0004 で確定するため GORM tag では指定しない（既存行への ADD COLUMN 対策）。
	CourseID        uint64    `gorm:"column:course_id;index" json:"courseId"`
	CreatedByUserID uint64    `gorm:"column:created_by_user_id;not null" json:"createdByUserId"`
	Title           string    `gorm:"column:title;not null;default:''" json:"title"`
	Content         string    `gorm:"column:content;type:text;not null;default:''" json:"content"`
	OrderInCourse   int       `gorm:"column:sort_order;not null;default:100" json:"orderInCourse"`
	IsPublished     bool      `gorm:"column:is_published;not null;default:false" json:"isPublished"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (TeachingMaterial) TableName() string { return "course_chapters" }
