package domain

import "time"

// TeachingMaterial はコースを構成する「章」。必ず 1 つの Course に所属する（course 1 : N chapter）。
// テーブルは course_chapters（FRESTYLE-184 で teaching_materials から改名）。
//
// 本文はリッチテキスト（tiptap の ProseMirror JSON）を doc(jsonb) に保持する。
// content(text・raw Markdown) は移行期間の互換用で、全章の一括変換完了後に撤去する。
// コース内の並び順は sort_order 列（同値時 ID 昇順）。
type TeachingMaterial struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID uint64 `gorm:"column:company_id;not null;index" json:"companyId"`
	// NOT NULL は migration 0004 で確定するため GORM tag では指定しない（既存行への ADD COLUMN 対策）。
	CourseID        uint64 `gorm:"column:course_id;index" json:"courseId"`
	CreatedByUserID uint64 `gorm:"column:created_by_user_id;not null" json:"createdByUserId"`
	Title           string `gorm:"column:title;not null;default:''" json:"title"`
	Content         string `gorm:"column:content;type:text;not null;default:''" json:"content"`
	// Doc はリッチテキスト本文（tiptap JSON）。未移行の章は NULL。
	// JSON 出力は handler 側で json.RawMessage として制御するため json:"-"。
	Doc *string `gorm:"column:doc;type:jsonb" json:"-"`
	// Revision は doc 更新の楽観ロック用。doc を更新するたびに +1（不一致は 409）。
	Revision int `gorm:"column:revision;not null;default:1" json:"revision"`
	// SchemaVersion は doc のエディタスキーマ版。読込時アップキャストの目印（現行 1）。
	SchemaVersion int       `gorm:"column:schema_version;not null;default:1" json:"schemaVersion"`
	OrderInCourse int       `gorm:"column:sort_order;not null;default:100" json:"orderInCourse"`
	IsPublished   bool      `gorm:"column:is_published;not null;default:false" json:"isPublished"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (TeachingMaterial) TableName() string { return "course_chapters" }
