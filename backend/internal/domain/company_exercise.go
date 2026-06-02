package domain

import "time"

// CompanyExercise は CompanyAdmin が自社 trainee 向けに作る独自問題（自社内のみ閲覧可）。
// 提出履歴は ExerciseSubmission.ExerciseKind = "company" で参照する。
type CompanyExercise struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID      uint64     `gorm:"column:company_id;not null;index" json:"companyId"`
	Language       string     `gorm:"size:32;not null;index" json:"language"`
	Title          string     `gorm:"size:200;not null" json:"title"`
	Description    string     `gorm:"type:text;not null" json:"description"`
	StarterCode    string     `gorm:"type:text;not null" json:"starterCode"`
	HintText       string     `gorm:"type:text" json:"hintText"`
	ExpectedOutput string     `gorm:"type:text" json:"expectedOutput"`
	Difficulty     int16      `gorm:"type:smallint;not null;default:1" json:"difficulty"`
	IsPublished    bool       `gorm:"not null;default:false" json:"isPublished"`
	ChapterID      *uint64    `gorm:"column:chapter_id" json:"chapterId,omitempty"`
	CreatedBy      uint64     `gorm:"column:created_by;not null" json:"createdBy"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

func (CompanyExercise) TableName() string { return "company_exercises" }
