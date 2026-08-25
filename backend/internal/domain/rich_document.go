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
	// CompanyID は作成時に作成者の所属会社を写し取る列で、公開文書の閲覧範囲を同一会社内へ
	// 閉じるためのテナント境界として使う（判定は CanBeReadBy が持つ）。
	// 写すのは作成時点の所属だけで、その後の異動では更新されない。未所属の作成者
	// （運営管理者など）や、この列を足す前に作られた行では nil になる。
	// そのため CanBeReadBy は所有者判定を先に置き、締めるのは「公開文書を他社が読むこと」
	// だけにしている（詳細は CanBeReadBy のコメント）。
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

// CompanyRef は文書に焼き付いた所属会社を返す。company_id は NULL を取り得る（未所属の
// 作成者・列を足す前の行）ので、0 という番兵値へ潰さず CompanyRef として表す。
func (d *RichDocument) CompanyRef() CompanyRef {
	if d.CompanyID == nil {
		return NoCompany()
	}
	return CompanyRefOf(*d.CompanyID)
}

// CanBeReadBy は viewerID / viewerCompany の利用者が本文書を読めるかを返す。
// 所有者は常に読める。公開文書は同一会社の利用者だけが読める（会社をまたいだ閲覧は不可）。
// viewerID=0（未認証）は所有者になり得ず、未所属の閲覧者はどの会社とも一致しない。
func (d *RichDocument) CanBeReadBy(viewerID uint64, viewerCompany CompanyRef) bool {
	// 所有者判定は必ずテナント一致より先に置く。CompanyID は作成時の所属の写しで移管を追わず、
	// 未所属の作成者や列を足す前の行では NULL のまま残る。所有者にまで会社一致を要求すると、
	// それらの文書が作成者自身からも見えなくなる。
	if viewerID != 0 && d.OwnerID == viewerID {
		return true
	}
	if !d.IsPublic {
		return false
	}
	// 「公開」は同一会社の中での公開に閉じる。会社が分からない文書（company_id が NULL）は
	// どの会社とも一致しないので、所有者以外からは見えなくなる。会社を特定できないものを
	// 全社へ開くのではなく見えない側へ倒す（fail-closed）。企業をまたいでノートを見せないという
	// 要件に対して、NULL を「誰にでも見せる」側へ倒すのは矛盾するため。
	companyID, known := d.CompanyRef().CompanyID()
	return known && viewerCompany.Matches(companyID)
}
