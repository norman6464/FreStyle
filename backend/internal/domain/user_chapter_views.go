package domain

import "time"

// UserChapterView はユーザーが章（教材）を開いた記録。
// PK = (user_id, chapter_id)。upsert により last_viewed_at と view_count を更新する。
// 実列の型は schema/core.sql が持つ（view_count は migration 0005 に合わせて integer）。
// フィールド個別のコメントは swaggo が API description に取り込むためここに書く。
type UserChapterView struct {
	UserID uint64 `json:"userId"`
	// TeachingMaterialID は章(course_chapters)の ID。DB 列は chapter_id(FRESTYLE-185 で改名)。
	// JSON キーは互換のため teachingMaterialId のまま。
	TeachingMaterialID uint64    `json:"teachingMaterialId"`
	CourseID           uint64    `json:"courseId"`
	FirstViewedAt      time.Time `json:"firstViewedAt"`
	LastViewedAt       time.Time `json:"lastViewedAt"`
	ViewCount          int       `json:"viewCount"`
}
