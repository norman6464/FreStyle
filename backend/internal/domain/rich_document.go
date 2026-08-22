package domain

import "time"

// DocumentKind は rich_documents の用途区分。
type DocumentKind string

const (
	// DocumentKindNote は学習メモ用途。
	DocumentKindNote DocumentKind = "note"
	// DocumentKindCourseChapter は教材（章）用途。
	DocumentKindCourseChapter DocumentKind = "course-chapter"
)

// Valid は既知の kind かを返す（作成リクエストの検証に使う）。
func (k DocumentKind) Valid() bool {
	switch k {
	case DocumentKindNote, DocumentKindCourseChapter:
		return true
	}
	return false
}

// RichDocument は tiptap の JSON を正本として保持するリッチテキスト文書。
// doc は ProseMirror ドキュメント（tiptap の getJSON() 結果）を jsonb でそのまま持つ。
type RichDocument struct {
	// ID は推測不能な UUID（Notion 風の URL）。作成時に repository が採番する。
	ID string `gorm:"type:uuid;primaryKey" json:"id"`
	// OwnerID は作成者。users.id への FK（制約は ApplyRichDocumentConstraints が張る）。
	OwnerID uint64 `gorm:"column:owner_id;not null;index" json:"ownerId"`
	// CompanyID は会社スコープ（テナント境界）。未所属なら nil。
	CompanyID *uint64 `gorm:"column:company_id" json:"companyId,omitempty"`
	// Kind は用途区分（note / course-chapter …）。
	Kind DocumentKind `gorm:"column:kind;not null" json:"kind"`
	// Title は一覧・検索用のタイトル。
	Title string `gorm:"column:title;not null" json:"title"`
	// IsPublic は公開可否（既定 false）。
	IsPublic bool `gorm:"column:is_public;not null;default:false" json:"isPublic"`
	// SchemaVersion はエディタ拡張の版。読込時アップキャストの目印。
	SchemaVersion int `gorm:"column:schema_version;not null;default:1" json:"schemaVersion"`
	// Doc は tiptap JSON を jsonb で保持する正本。API へは response 型で json.RawMessage に変換して出す。
	Doc string `gorm:"column:doc;type:jsonb;not null" json:"-"`
	// Revision は楽観ロック用の版番号。更新のたびに +1 する。
	Revision int `gorm:"column:revision;not null;default:1" json:"revision"`
	// PlainText / 保存履歴は初期スコープ外（後追い）。
	CreatedAt time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deletedAt,omitempty"`
}

// TableName は GORM のテーブル名を固定する。
func (RichDocument) TableName() string { return "rich_documents" }

// CanBeReadBy は viewerID が本文書を読めるかを返す。所有者、または公開文書は読める。
// viewerID=0（未認証）は非公開を読めない。
func (d *RichDocument) CanBeReadBy(viewerID uint64) bool {
	if d.IsPublic {
		return true
	}
	return viewerID != 0 && d.OwnerID == viewerID
}
