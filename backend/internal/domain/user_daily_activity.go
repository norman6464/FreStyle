package domain

import "time"

// UserDailyActivity はユーザーの1日分の学習活動をまとめたサマリーテーブル。
// PK = (user_id, activity_date)。書き込み時に upsert (+= delta) する。
type UserDailyActivity struct {
	UserID        uint64    `gorm:"column:user_id;primaryKey"    json:"userId"`
	ActivityDate  time.Time `gorm:"column:activity_date;type:date;primaryKey" json:"activityDate"`
	ExerciseCount int       `gorm:"column:exercise_count;default:0" json:"exerciseCount"`
	CorrectCount  int       `gorm:"column:correct_count;default:0"  json:"correctCount"`
	LessonCount   int       `gorm:"column:lesson_count;default:0"   json:"lessonCount"`
	AiChatCount   int       `gorm:"column:ai_chat_count;default:0"  json:"aiChatCount"`
	NoteCount     int       `gorm:"column:note_count;default:0"     json:"noteCount"`
}

func (UserDailyActivity) TableName() string { return "user_daily_activities" }
