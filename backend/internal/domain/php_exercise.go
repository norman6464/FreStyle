package domain

import "time"

// PhpExercise は PHP 学習教材の演習問題エンティティ。
// 難易度・カテゴリ別に整理し、スターターコードとヒントを持つ。
type PhpExercise struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	OrderIndex   int       `gorm:"not null;default:0"`
	Category     string    `gorm:"size:64;not null"`
	Title        string    `gorm:"size:200;not null"`
	Description  string    `gorm:"type:text;not null"`
	StarterCode  string    `gorm:"type:text;not null"`
	HintText     string    `gorm:"type:text"`
	ExpectedOutput string  `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
