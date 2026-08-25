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
	// CompanyID は作成時に作成者の所属会社を写し取るだけの列で、テナント境界としては機能していない。
	// 読み出し（FindByID / ListByOwner）も可視性判定（CanBeReadBy）も owner_id と is_public だけで
	// 決めており、この値を絞り込み条件に使っている箇所は無い（API 応答に出るだけで、フロントも読まない）。
	// いま絞り込みに使い始めると、これまで見えていた文書が会社をまたいだ瞬間に見えなくなる。
	//
	// テナント統合での扱いは「workspace_id へ置き換えず、列ごと捨てる」。写しているのは
	// 作成時点の作成者の所属で、その後の異動を追わないため既に事実として古い。指す先の
	// companies そのものが畳まれて消える以上、置き換えは「無かった境界を新しく作る」ことになり、
	// 上のとおり見えなくなる文書が出る。文書にテナント境界が要るなら、ナレッジ基盤の
	// 権限モデルの上で改めて設計する（この列を作り直すのではなく）。
	// 消す順序は「書き込みと API 応答を外す → 列を落とす」。逆にすると、ローリングデプロイ中の
	// 旧タスクが存在しない列へ INSERT して落ちる。未所属なら nil。
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
// CompanyID は意図的に見ない（現状テナント境界として機能していない。フィールドのコメント参照）。
func (d *RichDocument) CanBeReadBy(viewerID uint64) bool {
	if d.IsPublic {
		return true
	}
	return viewerID != 0 && d.OwnerID == viewerID
}
