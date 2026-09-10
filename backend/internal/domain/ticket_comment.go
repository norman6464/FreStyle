package domain

import (
	"errors"
	"time"
)

// TicketStatusTransition は状態が変わった記録 1 件（段 3・設計 Ⅵ）。
//
// 汎用の TicketChangeItem（field="status"）とは役割が違う — あちらは人が読む変更履歴
// （旧値・新値を ID と表示文字列の両方で持つ）、こちらは「いつどの状態にいたか」を
// 集計で引くための専用ログ（当時の表示名は持たず、常に FK を辿る）。
// ChangeTicketStatusUseCase が状態変更のたびに両方へ書く。
type TicketStatusTransition struct {
	ID              string    `json:"id"`
	TicketID        string    `json:"ticketId"`
	FromStatusID    string    `json:"fromStatusId"`
	ToStatusID      string    `json:"toStatusId"`
	ChangedByUserID uint64    `json:"changedByUserId"`
	ChangedAt       time.Time `json:"changedAt"`
}

// TicketComment はチケットへの発言 1 件（段 3）。ParentCommentID が nil ならトップレベル、
// 非 nil なら返信（同じチケット内の別の発言を指す）。
//
// Body は ProseMirror インラインノードの配列（JSON 文字列）。domain.ValidateCommentBody
// （comment.go・ノート側と共通）で検証する。
type TicketComment struct {
	ID              string  `json:"id"`
	WorkspaceID     string  `json:"-"`
	TicketID        string  `json:"ticketId"`
	ParentCommentID *string `json:"parentCommentId,omitempty"`
	AuthorUserID    uint64  `json:"-"`
	Body            string  `json:"-"`
	// EditedAt は編集済みなら値を持つ（未編集は nil）。以前の本文そのものは
	// TicketCommentEdit に別途積む（この構造体は最新の本文だけを運ぶ）。
	EditedAt  *time.Time `json:"editedAt,omitempty"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// TicketCommentEdit は発言 1 件の編集前スナップショット（段 3）。
type TicketCommentEdit struct {
	ID           string    `json:"id"`
	WorkspaceID  string    `json:"-"`
	CommentID    string    `json:"commentId"`
	EditorUserID uint64    `json:"editorUserId"`
	PreviousBody string    `json:"-"`
	EditedAt     time.Time `json:"editedAt"`
}

// TicketCommentReaction は発言への絵文字反応 1 件（段 3）。同じ人が同じ発言へ同じ絵文字で
// 反応するのは 1 回だけ（DB の複合主キー (comment_id, user_id, emoji) が守る）。
type TicketCommentReaction struct {
	CommentID string    `json:"commentId"`
	UserID    uint64    `json:"userId"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"createdAt"`
}

// TicketCommentReactionMaxEmojiBytes は emoji 列に許す最大バイト数
// （ck_ticket_comment_reactions_emoji_not_empty の DB 側 CHECK と同じ値）。
const TicketCommentReactionMaxEmojiBytes = 32

// ErrInvalidTicketCommentReactionEmoji は emoji が空・上限超過のときに返す。
var ErrInvalidTicketCommentReactionEmoji = errors.New("invalid ticket comment reaction emoji")

// ValidTicketCommentReactionEmoji は emoji が保存してよい形かを返す
// （空でない・DB の上限バイト数以内）。1 grapheme かどうかは見ない
// （kb ページアイコンの絵文字検証と同じ判断 — 見た目はピッカーが担う）。
func ValidTicketCommentReactionEmoji(emoji string) bool {
	return emoji != "" && len(emoji) <= TicketCommentReactionMaxEmojiBytes
}
