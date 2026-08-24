package domain

import "time"

// PageSnapshot はページのブロック行を組み直した ProseMirror ドキュメント（読み取り用のキャッシュ）。
// PK = page_id（1 ページ 1 行）。
//
// 表示のたびにブロック行を木に組み直すと 1 ページで数百行の取得と再帰的な組み立てが要るため、
// 編集のたびに 1 つの jsonb へ焼き直して読み出しを 1 行の取得に落とす。
// 正本はあくまで blocks 側で、この行は失っても blocks から再生成できる派生データ。
type PageSnapshot struct {
	// PageID は対象ページ。
	PageID string `gorm:"column:page_id;type:uuid;primaryKey" json:"pageId"`
	// Doc は tiptap の getJSON() 相当（type='doc' の ProseMirror ドキュメント）。
	// API へは handler の response 型で json.RawMessage に変換して出す。
	Doc string `gorm:"column:doc;type:jsonb;not null" json:"-"`
	// BuiltAt は焼き直した時刻。ブロックの更新時刻より古ければ作り直す判断に使う。
	BuiltAt time.Time `gorm:"column:built_at;not null" json:"builtAt"`
}

// TableName は GORM のテーブル名を固定する。
func (PageSnapshot) TableName() string { return "page_snapshots" }
