package domain

import "time"

// PracticeScenario は AI 練習モードのロールプレイ用シナリオ。
// Spring Boot の entity.PracticeScenario に相当。
type PracticeScenario struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Title         string    `gorm:"column:title" json:"title"`
	Description   string    `gorm:"column:description" json:"description"`
	Category      string    `gorm:"column:category" json:"category"`
	DifficultyLv  int       `gorm:"column:difficulty_level" json:"difficultyLevel"`
	SystemPrompt  string    `gorm:"column:system_prompt" json:"-"`
	IsActive      bool      `gorm:"column:is_active" json:"isActive"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (PracticeScenario) TableName() string { return "practice_scenarios" }
