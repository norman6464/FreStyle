package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrLabelNotFound は対象ラベルが存在しない（または別ワークスペース / 別スペースのもの）
// ときに返す。
var ErrLabelNotFound = errors.New("label not found")

// ErrLabelNameTaken は現役の中で同名（空白・大文字小文字を区別しない）が既にあるときに返す
// （uq_labels_space_name の一意制約違反を repository が翻訳する。ErrTicketStatusNameTaken と
// 同じ役割）。
var ErrLabelNameTaken = errors.New("label name is already taken")

// LabelRepository はラベル（labels）とチケットへの付け外し（ticket_labels）の永続化を担う。
// TicketRepository とは別 interface（TicketCommentRepository と同じ判断 — 固有の CRUD を
// 持つ機能領域は fat interface へ足さず、専用の interface + persistence struct を新設する）。
type LabelRepository interface {
	CreateLabel(ctx context.Context, l *domain.Label) error
	FindLabel(ctx context.Context, workspaceID, labelID string) (*domain.Label, error)
	ListLabels(ctx context.Context, workspaceID, spaceID string) ([]domain.Label, error)
	UpdateLabel(ctx context.Context, l *domain.Label) error
	DeleteLabel(ctx context.Context, workspaceID, labelID string) error

	AddTicketLabel(ctx context.Context, workspaceID, ticketID, labelID string) error
	RemoveTicketLabel(ctx context.Context, workspaceID, ticketID, labelID string) error
	ListLabelsByTicket(ctx context.Context, workspaceID, ticketID string) ([]domain.Label, error)
	// ListLabelsByTicketIDs は一覧画面向けにチケット ID ごとのラベルをまとめて引く
	// （N+1 を避ける。ticket_id -> ラベル一覧 の対応表で返す。1 件も無いチケットは
	// 対応表に現れない — 呼び出し側は空スライスとみなす）。
	ListLabelsByTicketIDs(ctx context.Context, workspaceID string, ticketIDs []string) (map[string][]domain.Label, error)
}
