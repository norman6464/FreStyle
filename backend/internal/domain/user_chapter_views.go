package domain

import "time"

// UserChapterView はユーザーが章（教材）を開いた記録。
// PK = (user_id, chapter_id)。upsert により last_viewed_at と view_count を更新する。
// GORM タグの type:integer は migration 0005 の実テーブル定義に合わせるための明示
// (省略すると AutoMigrate が bigint への ALTER を発行してしまう)。
// フィールド個別のコメントは swaggo が API description に取り込むためここに書く。
type UserChapterView struct {
	UserID uint64 `gorm:"column:user_id;primaryKey"              json:"userId"`
	// TeachingMaterialID は章(course_chapters)の ID。DB 列は chapter_id(FRESTYLE-185 で改名)。
	// JSON キーは互換のため teachingMaterialId のまま。
	TeachingMaterialID uint64    `gorm:"column:chapter_id;primaryKey"           json:"teachingMaterialId"`
	CourseID           uint64    `gorm:"column:course_id"                       json:"courseId"`
	FirstViewedAt      time.Time `gorm:"column:first_viewed_at"                 json:"firstViewedAt"`
	LastViewedAt       time.Time `gorm:"column:last_viewed_at"                  json:"lastViewedAt"`
	ViewCount          int       `gorm:"column:view_count;default:1;type:integer" json:"viewCount"`
}

func (UserChapterView) TableName() string { return "user_chapter_views" }
