package domain

import "time"

// UserDailyActivity はユーザーの1日分の学習活動をまとめたサマリーテーブル。
// PK = (user_id, activity_date)。書き込み時に upsert (+= delta) する。
// GORM タグの type:integer は migration 0005 の実テーブル定義に合わせるための明示
// (省略すると AutoMigrate が bigint への ALTER を発行してしまう)。
type UserDailyActivity struct {
	UserID        uint64    `gorm:"column:user_id;primaryKey"    json:"userId"`
	ActivityDate  time.Time `gorm:"column:activity_date;type:date;primaryKey" json:"activityDate"`
	ExerciseCount int       `gorm:"column:exercise_count;default:0;type:integer" json:"exerciseCount"`
	CorrectCount  int       `gorm:"column:correct_count;default:0;type:integer"  json:"correctCount"`
	LessonCount   int       `gorm:"column:lesson_count;default:0;type:integer"   json:"lessonCount"`
	// ChapterCount は lesson_count の後継(章を数えるカウンタ)。Expand では dual-write で同値に保つ。
	ChapterCount int `gorm:"column:chapter_count;default:0;type:integer" json:"-"`
	AiChatCount  int `gorm:"column:ai_chat_count;default:0;type:integer"  json:"aiChatCount"`
	NoteCount    int `gorm:"column:note_count;default:0;type:integer"     json:"noteCount"`
}

func (UserDailyActivity) TableName() string { return "user_daily_activities" }
