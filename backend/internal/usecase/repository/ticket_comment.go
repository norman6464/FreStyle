package repository

import (
	"context"
	"errors"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// ErrTicketCommentNotFound は対象の発言が存在しない（削除済み・他テナント・他チケットを含む）
// ときに返す。
var ErrTicketCommentNotFound = errors.New("ticket comment not found")

// TicketCommentRepository はチケットへの発言・編集履歴・反応の永続化を担う（段 3）。
// TicketRepository とは別の fat interface にする（これ以上 TicketRepository を太らせない
// 判断。InsertTicketStatusTransition だけは状態変更と同じ操作の一部なので TicketRepository
// 側に残す）。usecase/ticket が両方を depend する。
//
// 認可（CanComment / CanView）はここでは判定しない — handler が
// CheckTicketPermissionUseCase を先に通す（既存の TicketRepository と同じ分担）。
type TicketCommentRepository interface {
	CreateTicketComment(ctx context.Context, c *domain.TicketComment) error
	FindTicketComment(ctx context.Context, workspaceID, ticketID, commentID string) (*domain.TicketComment, error)
	ListTicketComments(ctx context.Context, workspaceID, ticketID string) ([]domain.TicketComment, error)
	// UpdateTicketCommentBody は本文を書き換える（edited_at を今にする）。呼び出し側
	// （usecase）が同一トランザクションで先に InsertTicketCommentEdit を呼び、編集前の
	// 本文を退避してから使う。
	UpdateTicketCommentBody(ctx context.Context, workspaceID, ticketID, commentID, body string) (*domain.TicketComment, error)
	DeleteTicketComment(ctx context.Context, workspaceID, ticketID, commentID string) error

	InsertTicketCommentEdit(ctx context.Context, e *domain.TicketCommentEdit) error
	ListTicketCommentEdits(ctx context.Context, workspaceID, commentID string) ([]domain.TicketCommentEdit, error)

	// AddTicketCommentReaction / RemoveTicketCommentReaction は反応の付け外し。付け外しは
	// 冪等（既に付いている反応をもう一度付けても・付いていない反応を外そうとしてもエラーに
	// しない）。
	AddTicketCommentReaction(ctx context.Context, workspaceID, commentID string, userID uint64, emoji string) error
	RemoveTicketCommentReaction(ctx context.Context, workspaceID, commentID string, userID uint64, emoji string) error
	// ListTicketCommentReactions は複数発言分の反応を 1 回でまとめて取る（一覧表示の N+1 回避）。
	ListTicketCommentReactions(ctx context.Context, workspaceID string, commentIDs []string) ([]domain.TicketCommentReaction, error)
}
