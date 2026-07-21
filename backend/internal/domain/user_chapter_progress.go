package domain

import "time"

// UserChapterProgress は trainee が章を完了したことの記録（user × chapter に 1 行）。
// user_lesson_progress の後継テーブル（FRESTYLE-186）。移行期は両テーブルへ dual-write し、
// Contract フェーズで読み書きを本テーブルへ寄せて旧テーブルを削除する。
//
// AutoMigrate に登録することで、SQL migration を流さない CI / ローカル環境でも
// このテーブルが作られる（migration 0017 と同じ構造・同じ index 名になるようタグを揃える）。
type UserChapterProgress struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"column:user_id;not null;uniqueIndex:ux_user_chapter_progress" json:"userId"`
	ChapterID   uint64    `gorm:"column:chapter_id;not null;uniqueIndex:ux_user_chapter_progress" json:"chapterId"`
	CourseID    uint64    `gorm:"column:course_id;not null;index" json:"courseId"`
	CompletedAt time.Time `gorm:"column:completed_at;not null" json:"completedAt"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (UserChapterProgress) TableName() string { return "user_chapter_progress" }
