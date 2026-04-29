package domain

import "time"

type WeeklyChallenge struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	WeekStart   time.Time `gorm:"column:week_start;type:date" json:"weekStart"`
	Title       string    `gorm:"column:title" json:"title"`
	Description string    `gorm:"column:description" json:"description"`
	IsActive    bool      `gorm:"column:is_active" json:"isActive"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (WeeklyChallenge) TableName() string { return "weekly_challenges" }

type WeeklyChallengeProgress struct {
	UserID      uint64    `gorm:"primaryKey;column:user_id" json:"userId"`
	ChallengeID uint64    `gorm:"primaryKey;column:challenge_id" json:"challengeId"`
	Completed   bool      `gorm:"column:completed" json:"completed"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (WeeklyChallengeProgress) TableName() string { return "weekly_challenge_progress" }
