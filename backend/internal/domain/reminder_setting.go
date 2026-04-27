package domain

import "time"

type ReminderSetting struct {
	UserID       uint64    `gorm:"primaryKey;column:user_id" json:"userId"`
	HourLocal    int       `gorm:"column:hour_local" json:"hourLocal"`
	MinuteLocal  int       `gorm:"column:minute_local" json:"minuteLocal"`
	WeekdaysMask int       `gorm:"column:weekdays_mask" json:"weekdaysMask"`
	IsActive     bool      `gorm:"column:is_active" json:"isActive"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (ReminderSetting) TableName() string { return "reminder_settings" }
