package domain

import "time"

// MasterExercise は運営が用意した練習問題のマスタ。Language 列で言語を表現し言語非依存に扱う。
type MasterExercise struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug           string `gorm:"size:64;not null;uniqueIndex" json:"slug"`
	Language       string `gorm:"size:32;not null;index" json:"language"`
	SortOrder      int    `gorm:"column:sort_order;not null;default:0" json:"orderIndex"`
	Category       string `gorm:"size:64;not null" json:"category"`
	Title          string `gorm:"size:200;not null" json:"title"`
	Description    string `gorm:"type:text;not null" json:"description"`
	StarterCode    string `gorm:"type:text;not null" json:"starterCode"`
	HintText       string `gorm:"type:text" json:"hintText"`
	ExpectedOutput string `gorm:"type:text" json:"expectedOutput"`
	// Mode は採点モード。execute は実行して stdout 比較、qa は提出文字列と ExpectedOutput を trim 比較。
	Mode string `gorm:"size:16;not null;default:'execute'" json:"mode"`
	// Explanation は qa モードで正解後に表示する markdown 解説。
	Explanation string    `gorm:"type:text;not null;default:''" json:"explanation"`
	Difficulty  int16     `gorm:"type:smallint;not null;default:1" json:"difficulty"`
	IsPublished bool      `gorm:"not null;default:true" json:"isPublished"`
	ChapterID   *uint64   `gorm:"column:chapter_id" json:"chapterId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (MasterExercise) TableName() string { return "master_exercises" }

// 対応言語の定数。追加時は usecase / フロント側の許容セットも揃える。
const (
	ExerciseLanguagePhp  = "php"
	ExerciseLanguageSql  = "sql"
	ExerciseLanguageGo   = "go"
	ExerciseLanguageJs   = "javascript"
	ExerciseLanguageBash = "bash"
	ExerciseLanguageGit  = "git"
	ExerciseLanguageJava = "java"
)

// 採点モードの定数。
const (
	ExerciseModeExecute = "execute"
	ExerciseModeQA      = "qa"
)
