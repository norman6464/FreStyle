package domain

import (
	"encoding/json"
	"errors"
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

// ValidateCommentBody は comments.body に保存してよい形かを検証する
// （JSON 配列であること。空配列は「本文の無いコメント」なので拒否する）。
//
// 中身（各ノードの type 等）までは検証しない — blocks.inline の parseBlockNode 相当の
// 検証は無いが、これは既存の本文（blocks.inline）自体が個々のインラインノードの型を
// 厳密検証していないのと同じ水準（parseBlockNode は「JSON 配列であること」しか見ていない）。
func ValidateCommentBody(raw string) error {
	var items []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return ErrInvalidCommentBody
	}
	if len(items) == 0 {
		return ErrInvalidCommentBody
	}
	return nil
}
