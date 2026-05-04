package domain

import "time"

// PhpExercise は PHP 学習教材の演習問題エンティティ。
// 難易度・カテゴリ別に整理し、スターターコードとヒントを持つ。
type PhpExercise struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderIndex     int       `gorm:"not null;default:0" json:"orderIndex"`
	Category       string    `gorm:"size:64;not null" json:"category"`
	Title          string    `gorm:"size:200;not null" json:"title"`
	Description    string    `gorm:"type:text;not null" json:"description"`
	StarterCode    string    `gorm:"type:text;not null" json:"starterCode"`
	HintText       string    `gorm:"type:text" json:"hintText"`
	ExpectedOutput string    `gorm:"type:text" json:"expectedOutput"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
