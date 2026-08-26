package domain

import "time"

// CourseCategory は「色＝学習領域」の連想を支える定義済み分類（FRESTYLE-67）。
// 自由入力にすると色の一貫性が崩れるため選択式とし、表示色は frontend の
// カラーマップが持つ。backend は値の正当性のみ保証する。
const (
	CourseCategoryDevBasics    = "dev-basics"   // 開発基礎
	CourseCategoryBackend      = "backend"      // バックエンド開発
	CourseCategoryArchitecture = "architecture" // 設計・アーキテクチャ
	CourseCategoryDatabase     = "database"     // データベース
	CourseCategoryInfra        = "infra"        // インフラ・クラウド
	CourseCategorySecurity     = "security"     // セキュリティ
	CourseCategoryProduct      = "product"      // プロダクト・仕様
	CourseCategoryDesign       = "design"       // 設計・デザインパターン
)

// ValidCourseCategories は選択可能なカテゴリの一覧（未分類 = 空文字は含まない）。
var ValidCourseCategories = []string{
	CourseCategoryDevBasics,
	CourseCategoryBackend,
	CourseCategoryArchitecture,
	CourseCategoryDatabase,
	CourseCategoryInfra,
	CourseCategorySecurity,
	CourseCategoryProduct,
	CourseCategoryDesign,
}

// IsValidCourseCategory は c が未分類("")または定義済みカテゴリかを返す。
func IsValidCourseCategory(c string) bool {
	if c == "" {
		return true
	}
	for _, v := range ValidCourseCategories {
		if v == c {
			return true
		}
	}
	return false
}

// Course は教材を束ねるコース。階層は Company 1 ── * Course 1 ── * TeachingMaterial。
// trainee は自社の is_published=true のみ閲覧可。並び順は SortOrder（同値時 ID 昇順）。
type Course struct {
	ID              uint64 `json:"id"`
	CompanyID       uint64 `json:"companyId"`
	CreatedByUserID uint64 `json:"createdByUserId"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	// Language は主に扱う言語・技術（例: "go" / "docker" / "terraform"。空 = 言語が主題でない）。
	// 演習の language と同じ自由文字列方式で、表示色は frontend のカラーマップが持つ。
	Language    string    `json:"language"`
	SortOrder   int       `json:"sortOrder"`
	IsPublished bool      `json:"isPublished"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
