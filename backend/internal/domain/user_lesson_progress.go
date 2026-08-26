package domain

import "time"

// UserLessonProgress は trainee が章を完了したことの記録。
// 1 行 = その (user, chapter) が完了済み。未完了に戻すときは行を削除する。
// (user_id, chapter_id) は複合ユニーク（同じ章の二重記録を防ぐ）。
type UserLessonProgress struct {
	ID     uint64 `json:"id"`
	UserID uint64 `json:"userId"`
	// TeachingMaterialID は章(course_chapters)の ID。DB 列は chapter_id(FRESTYLE-185 で改名)。
	// JSON キーは互換のため teachingMaterialId のまま。
	TeachingMaterialID uint64    `json:"teachingMaterialId"`
	CourseID           uint64    `json:"courseId"`
	CompletedAt        time.Time `json:"completedAt"`
	CreatedAt          time.Time `json:"createdAt"`
}
