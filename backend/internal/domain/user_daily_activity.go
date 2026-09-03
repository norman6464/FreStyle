package domain

import "time"

// UserDailyActivity はユーザーの1日分の学習活動をまとめたサマリーテーブル。
// PK = (user_id, activity_date)。書き込み時に upsert (+= delta) する。
// 実列の型は schema/core.sql が持つ（各 *_count は migration 0005 に合わせて integer）。
type UserDailyActivity struct {
	UserID        uint64    `json:"userId"`
	ActivityDate  time.Time `json:"activityDate"`
	ExerciseCount int       `json:"exerciseCount"`
	CorrectCount  int       `json:"correctCount"`
	LessonCount   int       `json:"lessonCount"`
	NoteCount     int       `json:"noteCount"`
}
