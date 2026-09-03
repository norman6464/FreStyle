package domain

import "time"

// UserChapterView はユーザーが章（教材）を開いた記録。
// PK = (user_id, chapter_id)。upsert により last_viewed_at と view_count を更新する。
// 実列の型は schema/core.sql が持つ（view_count は migration 0005 に合わせて integer）。
type UserChapterView struct {
	UserID             uint64    `json:"userId"`
	TeachingMaterialID uint64    `json:"teachingMaterialId"`
	CourseID           uint64    `json:"courseId"`
	FirstViewedAt      time.Time `json:"firstViewedAt"`
	LastViewedAt       time.Time `json:"lastViewedAt"`
	ViewCount          int       `json:"viewCount"`
}
