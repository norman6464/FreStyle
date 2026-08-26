package domain

import "time"

// MasterExercise は運営が用意した練習問題のマスタ。Language 列で言語を表現し言語非依存に扱う。
type MasterExercise struct {
	ID             uint64 `json:"id"`
	Slug           string `json:"slug"`
	Language       string `json:"language"`
	SortOrder      int    `json:"orderIndex"`
	Category       string `json:"category"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	StarterCode    string `json:"starterCode"`
	HintText       string `json:"hintText"`
	ExpectedOutput string `json:"expectedOutput"`
	// Mode は採点モード。execute は実行して stdout 比較、qa は提出文字列と ExpectedOutput を trim 比較。
	Mode string `json:"mode"`
	// Explanation は qa モードで正解後に表示する markdown 解説。
	Explanation string    `json:"explanation"`
	Difficulty  int16     `json:"difficulty"`
	IsPublished bool      `json:"isPublished"`
	ChapterID   *uint64   `json:"chapterId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// 対応言語の定数。追加時は usecase / フロント側の許容セットも揃える。
const (
	ExerciseLanguagePhp  = "php"
	ExerciseLanguageSql  = "sql"
	ExerciseLanguageGo   = "go"
	ExerciseLanguageJs   = "javascript"
	ExerciseLanguageBash = "bash"
	ExerciseLanguageGit  = "git"
	ExerciseLanguageJava = "java"
	ExerciseLanguageHTML = "html"
	ExerciseLanguageRuby = "ruby"
	ExerciseLanguageC    = "c"
	ExerciseLanguageCpp  = "cpp"
)

// 採点モードの定数。
const (
	ExerciseModeExecute = "execute"
	ExerciseModeQA      = "qa"
	// ExerciseModePreview はブラウザ描画のライブプレビュー演習(HTML/CSS 等)。
	// サーバー実行せず、学習者が見本と見比べて完了を宣言する(視覚的セルフチェック)。
	ExerciseModePreview = "preview"
)
