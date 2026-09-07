package domain

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"
)

// ErrInvalidCommentBody は comments.body に保存できない形（JSON 配列でない・空配列）の
// 本文が渡されたときに返す。usecase はこれをそのまま呼び出し元へ伝播させ、
// handler は 400 invalid_request へマップする。
var ErrInvalidCommentBody = errors.New("invalid comment body")

// ErrInvalidCommentAnchor は錨（block_id / anchor_from / anchor_to / quote）の組み合わせが
// 不正なときに返す（段 3・錨付きコメント）。usecase はこれをそのまま呼び出し元へ伝播させ、
// handler は 400 invalid_comment_anchor へマップする。
var ErrInvalidCommentAnchor = errors.New("invalid comment anchor")

// CommentAnchorMaxQuoteLen は quote に許す最大バイト数。バブルメニューから送られる引用文は
// 通常短いが、極端に長い選択をそのまま保存させないための上限。
const CommentAnchorMaxQuoteLen = 2000

// CommentThread はページ全体、またはページ内の特定ブロック・特定文字範囲（錨）に
// 付いたコメントのスレッド。
//
// BlockID / AnchorFrom / AnchorTo / Quote は 4 つとも nil（page-level）か、4 つとも
// 非 nil（錨付き）のどちらか — ValidateCommentAnchor がこの不変条件を守る。
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

// ValidateCommentAnchor は錨（block_id / anchor_from / anchor_to / quote）の組み合わせが
// 保存してよい形かを検証する。
//
//   - 4 つとも nil なら有効（page-level のコメント）
//   - 4 つとも非 nil なら、以下をすべて満たすときだけ有効:
//     blockID が空文字でない／anchorFrom が 0 以上／anchorFrom < anchorTo／
//     quote が空白のみでなく CommentAnchorMaxQuoteLen バイト以下
//   - それ以外（一部だけ非 nil の中途半端な組み合わせ）は無効
//
// blockID が実際にそのページに属するか（他ページ・他テナントのブロックでないか）は
// ここでは見ない。DB を引く必要があるため repository 層（BlockExistsInPage）の責務。
func ValidateCommentAnchor(blockID *string, anchorFrom, anchorTo *int, quote *string) error {
	present := 0
	for _, v := range []bool{blockID != nil, anchorFrom != nil, anchorTo != nil, quote != nil} {
		if v {
			present++
		}
	}
	if present == 0 {
		return nil
	}
	if present != 4 {
		return ErrInvalidCommentAnchor
	}
	if *blockID == "" {
		return ErrInvalidCommentAnchor
	}
	if *anchorFrom < 0 {
		return ErrInvalidCommentAnchor
	}
	if *anchorFrom >= *anchorTo {
		return ErrInvalidCommentAnchor
	}
	// comment_threads.anchor_from/anchor_to は DB 上 int（32bit）。ここで弾いておかないと
	// persistence 層で int32(v) への縮小キャストが符号ごと丸め込まれた値のまま保存されてしまう
	// （ORM 移行で踏んだ「縮小キャストの無言失敗」と同種の罠）。ここは「形」の検証なので domain の責務。
	if *anchorFrom < math.MinInt32 || *anchorFrom > math.MaxInt32 ||
		*anchorTo < math.MinInt32 || *anchorTo > math.MaxInt32 {
		return ErrInvalidCommentAnchor
	}
	if strings.TrimSpace(*quote) == "" {
		return ErrInvalidCommentAnchor
	}
	if len(*quote) > CommentAnchorMaxQuoteLen {
		return ErrInvalidCommentAnchor
	}
	return nil
}
