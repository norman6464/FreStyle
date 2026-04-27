package domain

import "time"

type DailyGoal struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"column:user_id;index" json:"userId"`
	Date       time.Time `gorm:"column:goal_date;type:date" json:"date"`
	TargetMin  int       `gorm:"column:target_minutes" json:"targetMinutes"`
	ActualMin  int       `gorm:"column:actual_minutes" json:"actualMinutes"`
	IsAchieved bool      `gorm:"column:is_achieved" json:"isAchieved"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (DailyGoal) TableName() string { return "daily_goals" }
