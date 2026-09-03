package domain

import "time"

const (
	CourseCategoryDevBasics    = "dev-basics"
	CourseCategoryBackend      = "backend"
	CourseCategoryArchitecture = "architecture"
	CourseCategoryDatabase     = "database"
	CourseCategoryInfra        = "infra"
	CourseCategorySecurity     = "security"
	CourseCategoryProduct      = "product"
	CourseCategoryDesign       = "design"
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
	ID              uint64    `json:"id"`
	CreatedByUserID uint64    `json:"createdByUserId"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	Language        string    `json:"language"`
	SortOrder       int       `json:"sortOrder"`
	WorkspaceID     *string   `json:"workspaceId,omitempty"`
	IsPublished     bool      `json:"isPublished"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (c Course) WorkspaceRef() WorkspaceRef {
	if c.WorkspaceID == nil {
		return NoWorkspace()
	}
	return WorkspaceRefOf(*c.WorkspaceID)
}
