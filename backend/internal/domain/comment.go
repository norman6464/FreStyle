package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrInvalidCommentBody は comments.body に保存できない形（JSON 配列でない・空配列）の
// 本文が渡されたときに返す。usecase はこれをそのまま呼び出し元へ伝播させ、
// handler は 400 invalid_request へマップする。
var ErrInvalidCommentBody = errors.New("invalid comment body")

// CommentThread はページ（または将来ブロック）に付いたコメントのスレッド。
//
// BlockID / AnchorFrom / AnchorTo / Quote は段 3（錨付きコメント）のための列で、
// このPR（段 2・ページ全体へのコメント）の書き込み経路は「本文」だけを受け取るため
// 常に nil のまま作られる。
type CommentThread struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"-"`
	PageID      string  `json:"-"`
	BlockID     *string `json:"blockId,omitempty"`
	AnchorFrom  *int    `json:"anchorFrom,omitempty"`
	AnchorTo    *int    `json:"anchorTo,omitempty"`
	Quote       *string `json:"quote,omitempty"`
	// ResolvedAt / ResolvedByUserID は両方あるか両方無いか（DB の CHECK と同じ不変条件）。
	ResolvedAt       *time.Time `json:"resolvedAt,omitempty"`
	ResolvedByUserID *uint64    `json:"-"`
	CreatedByUserID  uint64     `json:"-"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// Resolved はこのスレッドが解決済みかを返す。
func (t CommentThread) Resolved() bool { return t.ResolvedAt != nil }

// Comment はスレッドに付いた 1 件の発言（スレッドを開いた最初の発言も返信も同じ形）。
type Comment struct {
	ID           string `json:"id"`
	ThreadID     string `json:"-"`
	AuthorUserID uint64 `json:"-"`
	// Body は ProseMirror インラインノードの配列（JSON 文字列）。blocks.inline と同じ形。
	// API へは handler の response 型で json.RawMessage に変換して出す。
	Body      string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// commentInlineNode は comments.body の要素 1 つの最小限の形。ProseMirror インライン
// ノードは type を必ず持ち、text ノードは非空の text を持つ、という 2 点だけを見る
// （marks の中身までは検証しない。過検証で将来のノード種別を締め出さないため）。
type commentInlineNode struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ValidateCommentBody は comments.body に保存してよい形かを検証する。
//
//   - JSON 配列であること。空配列は「本文の無いコメント」なので拒否する
//   - 各要素は object で、type を持つこと（null・{} はここで弾かれる。json.Unmarshal は
//     JSON の null を非ポインタの struct へ当てても無効化しない = ゼロ値のままなので、
//     type=="" のチェックで null 要素も {} 要素も同じ理由で弾ける）
//   - type=="text" の要素は、空白のみでない text を持つこと（"" や空白だけの text は
//     見た目上「本文の無い発言」になり、KbCommentItem が text を持たないノードを
//     無視するため、保存後に空のコメントとして残ってしまう — CodeRabbit 指摘）
//
// これ以上（marks の中身・type の許可リスト等）は検証しない。blocks.inline の
// parseBlockNode も同じ水準（JSON 配列であること）までしか見ておらず、それに揃える。
func ValidateCommentBody(raw string) error {
	var items []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return ErrInvalidCommentBody
	}
	if len(items) == 0 {
		return ErrInvalidCommentBody
	}
	for _, item := range items {
		var node commentInlineNode
		if err := json.Unmarshal(item, &node); err != nil {
			return ErrInvalidCommentBody
		}
		if node.Type == "" {
			return ErrInvalidCommentBody
		}
		if node.Type == "text" && strings.TrimSpace(node.Text) == "" {
			return ErrInvalidCommentBody
		}
	}
	return nil
}
