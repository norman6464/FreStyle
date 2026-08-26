package domain

import "time"

// CompanyExercise は CompanyAdmin が自社 trainee 向けに作る独自問題（自社内のみ閲覧可）。
// 提出履歴は ExerciseSubmission.ExerciseKind = "company" で参照する。
type CompanyExercise struct {
	ID             uint64     `json:"id"`
	CompanyID      uint64     `json:"companyId"`
	Language       string     `json:"language"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	StarterCode    string     `json:"starterCode"`
	HintText       string     `json:"hintText"`
	ExpectedOutput string     `json:"expectedOutput"`
	Difficulty     int16      `json:"difficulty"`
	IsPublished    bool       `json:"isPublished"`
	ChapterID      *uint64    `json:"chapterId,omitempty"`
	CreatedBy      uint64     `json:"createdBy"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}
