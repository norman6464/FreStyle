package domain

import "time"

// UserLessonProgress は trainee が章を完了したことの記録。
// 1 行 = その (user, chapter) が完了済み。未完了に戻すときは行を削除する。
// (user_id, chapter_id) は複合ユニーク（同じ章の二重記録を防ぐ）。
type UserLessonProgress struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID uint64 `gorm:"column:user_id;not null;uniqueIndex:ux_user_chapter_progress_user_chapter" json:"userId"`
	// TeachingMaterialID は章(course_chapters)の ID。DB 列は chapter_id(FRESTYLE-185 で改名)。
	// JSON キーは互換のため teachingMaterialId のまま。
	TeachingMaterialID uint64    `gorm:"column:chapter_id;not null;uniqueIndex:ux_user_chapter_progress_user_chapter" json:"teachingMaterialId"`
	CourseID           uint64    `gorm:"column:course_id;not null;index" json:"courseId"`
	CompletedAt        time.Time `gorm:"column:completed_at;not null" json:"completedAt"`
	CreatedAt          time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (UserLessonProgress) TableName() string { return "user_lesson_progress" }
