package domain

import "time"

// UserLessonProgress は trainee が教材（レッスン）を完了したことの記録。
// 1 行 = その (user, teaching_material) が完了済み。未完了に戻すときは行を削除する。
// (user_id, teaching_material_id) は複合ユニーク（同じ教材の二重記録を防ぐ）。
type UserLessonProgress struct {
	ID                 uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID             uint64 `gorm:"column:user_id;not null;uniqueIndex:ux_user_lesson" json:"userId"`
	TeachingMaterialID uint64 `gorm:"column:teaching_material_id;not null;uniqueIndex:ux_user_lesson" json:"teachingMaterialId"`
	// ChapterID は teaching_material_id の後継(参照先 course_chapters に合わせた FK 名)。
	// Expand フェーズでは dual-write で両列を同値に保つ。Contract で旧列を削除する。
	ChapterID   uint64    `gorm:"column:chapter_id" json:"-"`
	CourseID    uint64    `gorm:"column:course_id;not null;index" json:"courseId"`
	CompletedAt time.Time `gorm:"column:completed_at;not null" json:"completedAt"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (UserLessonProgress) TableName() string { return "user_lesson_progress" }
