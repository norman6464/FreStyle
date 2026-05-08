package domain

import "time"

// MasterExercise は運営（FreStyle 側）が用意した練習問題のマスタエンティティ。
//
// 旧 `php_exercises` を「言語非依存の汎用テーブル」に拡張したもの。
// `language` 列で 'php' / 'sql' / 'go' / 'js' などを表現し、テーブル単位での分割は行わない。
//
// CompanyAdmin が自社向けに作る独自問題は別テーブル `CompanyExercise` で扱う
// （提出履歴は両者を `ExerciseSubmission.ExerciseKind` で polymorphic に参照）。
//
// `ChapterID` は将来 PR-H1 で chapters テーブルが入ったとき、章末演習として
// 紐付けるための任意 FK。本 PR では NULL のまま運用する。
type MasterExercise struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug           string    `gorm:"size:64;not null;uniqueIndex" json:"slug"`
	Language       string    `gorm:"size:32;not null;index" json:"language"`
	OrderIndex     int       `gorm:"not null;default:0" json:"orderIndex"`
	Category       string    `gorm:"size:64;not null" json:"category"`
	Title          string    `gorm:"size:200;not null" json:"title"`
	Description    string    `gorm:"type:text;not null" json:"description"`
	StarterCode    string    `gorm:"type:text;not null" json:"starterCode"`
	HintText       string    `gorm:"type:text" json:"hintText"`
	ExpectedOutput string    `gorm:"type:text" json:"expectedOutput"`
	Difficulty     int16     `gorm:"type:smallint;not null;default:1" json:"difficulty"`
	IsPublished    bool      `gorm:"not null;default:true" json:"isPublished"`
	ChapterID      *uint64   `gorm:"column:chapter_id" json:"chapterId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (MasterExercise) TableName() string { return "master_exercises" }

// 演習問題の対応言語の文字列定数。新規追加時は usecase / フロント側の許容セットも揃える。
const (
	ExerciseLanguagePhp = "php"
	ExerciseLanguageSql = "sql"
	ExerciseLanguageGo  = "go"
	ExerciseLanguageJs  = "javascript"
)
