package domain

// PagePath はページの祖先関係を平らに持つ派生テーブル（closure table）。
// PK = (page_id, ancestor_id)。自分自身も depth=0 の行として持つ。
//
// pages.parent_id の連鎖だけでも木は表せるが、パンくず・サブツリー一括取得・移動時の
// 循環検出を再帰クエリなしの 1 回の JOIN で済ませるためにこの索引を別に持つ。
// あくまで pages から導ける派生データなので、正本は pages.parent_id 側。
type PagePath struct {
	// WorkspaceID はテナント境界。page_id / ancestor_id との複合 FK に使い、
	// 「別ワークスペースの 2 ページを組にした行」を DB が弾けるようにする
	// （単独 FK を 2 本張るだけでは、どちらの FK も通ってしまい防げない）。
	WorkspaceID string `gorm:"column:workspace_id;type:uuid;not null;index" json:"workspaceId"`
	// PageID は子孫側のページ。
	PageID string `gorm:"column:page_id;type:uuid;primaryKey" json:"pageId"`
	// AncestorID は祖先側のページ（page_id 自身を含む）。
	AncestorID string `gorm:"column:ancestor_id;type:uuid;primaryKey;index" json:"ancestorId"`
	// Depth は祖先までの距離。自分自身が 0、親が 1。
	Depth int `gorm:"column:depth;type:integer;not null" json:"depth"`
}

// TableName は GORM のテーブル名を固定する。
func (PagePath) TableName() string { return "page_paths" }
