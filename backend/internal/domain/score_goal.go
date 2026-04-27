package domain

import "time"

// ScoreGoal はユーザーごとの目標スコア。
type ScoreGoal struct {
	UserID      uint64    `gorm:"primaryKey;column:user_id" json:"userId"`
	TargetScore float64   `gorm:"column:target_score" json:"targetScore"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (ScoreGoal) TableName() string { return "score_goals" }
