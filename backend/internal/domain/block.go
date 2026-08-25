package domain

import "time"

// BlockType は blocks.type に入るノード名。値は ProseMirror（tiptap）のノード名そのもの。
//
// frontend の createSchemaExtensions()（shared/ui/RichTextEditor/schemaExtensions.ts）が
// 組み立てるスキーマと 1 対 1 に対応する。片方を増やしたらもう片方も足すこと
// （スキーマにないノード名を保存すると、読み出したドキュメントがエディタで開けなくなる）。
type BlockType string

// ブロック行として保存するノード名の一覧。
//
// doc は「ページそのもの」なのでブロック行にはしない。text / hardBreak などのインラインノードは
// 行にせず、親ブロックの inline（jsonb）に ProseMirror の content 配列として持たせる
// （文字単位で行を作ると 1 段落の編集が大量の行更新になるため、行の粒度はブロックで止める）。
const (
	BlockTypeParagraph      BlockType = "paragraph"
	BlockTypeHeading        BlockType = "heading"
	BlockTypeCodeBlock      BlockType = "codeBlock"
	BlockTypeBlockquote     BlockType = "blockquote"
	BlockTypeBulletList     BlockType = "bulletList"
	BlockTypeOrderedList    BlockType = "orderedList"
	BlockTypeListItem       BlockType = "listItem"
	BlockTypeTaskList       BlockType = "taskList"
	BlockTypeTaskItem       BlockType = "taskItem"
	BlockTypeTable          BlockType = "table"
	BlockTypeTableRow       BlockType = "tableRow"
	BlockTypeTableHeader    BlockType = "tableHeader"
	BlockTypeTableCell      BlockType = "tableCell"
	BlockTypeImage          BlockType = "image"
	BlockTypeHorizontalRule BlockType = "horizontalRule"
)

// ValidBlockTypes は保存を許すノード名の一覧（登録順は表示順とは無関係）。
var ValidBlockTypes = []BlockType{
	BlockTypeParagraph,
	BlockTypeHeading,
	BlockTypeCodeBlock,
	BlockTypeBlockquote,
	BlockTypeBulletList,
	BlockTypeOrderedList,
	BlockTypeListItem,
	BlockTypeTaskList,
	BlockTypeTaskItem,
	BlockTypeTable,
	BlockTypeTableRow,
	BlockTypeTableHeader,
	BlockTypeTableCell,
	BlockTypeImage,
	BlockTypeHorizontalRule,
}

// Valid は既知のノード名かを返す（保存前の検証に使う）。
func (t BlockType) Valid() bool {
	for _, v := range ValidBlockTypes {
		if v == t {
			return true
		}
	}
	return false
}

// Block はページ本文を構成する 1 ブロック（段落・見出し・リスト項目・表のセル …）。
//
// ページ全体を 1 つの jsonb に持つ rich_documents と違い、ブロックを行に分解して持つ。
// 部分更新・ブロック単位のリンク / コメント・全文検索の単位を DB 側で扱えるようにするため。
// 入れ子（リストや表）は ParentID の自己参照で表し、兄弟の並びは Position（分数インデックス）で持つ。
//
// Workspace と同じくナレッジ基盤の型なので GORM を通さない（段 1-b で repository が付くまで参照元は無い）。
type Block struct {
	ID string `json:"id"`
	// WorkspaceID はテナント境界。page / 親ブロックとの複合 FK に使う。
	WorkspaceID string `json:"workspaceId"`
	// PageID は所属ページ。(workspace_id, page_id) の複合 FK で pages を参照する。
	PageID string `json:"pageId"`
	// ParentID は親ブロック。NULL はページ直下（トップレベル）を意味する。
	ParentID *string `json:"parentId,omitempty"`
	// Position は兄弟内の並び順を表す分数インデックス（fracindex.Between で採番する）。
	Position string `json:"position"`
	// Type は ProseMirror のノード名。
	Type BlockType `json:"type"`
	// Attrs は ProseMirror の attrs（見出しの level、コードブロックの language など）を jsonb で持つ。
	// 属性が無いノードでも空オブジェクト {} を入れる（NULL と {} の二通りを作らない）。
	// API へは handler の response 型で json.RawMessage に変換して出す。
	Attrs string `json:"-"`
	// Inline は葉ノードのインライン内容（text ノードとマークの配列）。
	// リストや表のような容器ノードは子をブロック行として持つため NULL にする。
	Inline    *string   `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
