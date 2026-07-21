package domain

import "time"

// UserChapterView はユーザーが章（教材）を開いた記録。
// PK = (user_id, teaching_material_id)。upsert により last_viewed_at と view_count を更新する。
// GORM タグの type:integer は migration 0005 の実テーブル定義に合わせるための明示
// (省略すると AutoMigrate が bigint への ALTER を発行してしまう)。
// フィールド個別のコメントは swaggo が API description に取り込むためここに書く。
type UserChapterView struct {
	UserID             uint64 `gorm:"column:user_id;primaryKey"              json:"userId"`
	TeachingMaterialID uint64 `gorm:"column:teaching_material_id;primaryKey" json:"teachingMaterialId"`
	// ChapterID は teaching_material_id の後継(参照先 course_chapters に合わせた FK 名)。
	// Expand フェーズでは dual-write で両列を同値に保つ。Contract で旧列を削除する。
	ChapterID     uint64    `gorm:"column:chapter_id" json:"-"`
	CourseID      uint64    `gorm:"column:course_id"                       json:"courseId"`
	FirstViewedAt time.Time `gorm:"column:first_viewed_at"                 json:"firstViewedAt"`
	LastViewedAt  time.Time `gorm:"column:last_viewed_at"                  json:"lastViewedAt"`
	ViewCount     int       `gorm:"column:view_count;default:1;type:integer" json:"viewCount"`
}

func (UserChapterView) TableName() string { return "user_chapter_views" }
