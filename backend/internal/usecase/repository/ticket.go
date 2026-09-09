package repository

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrTicketNotFound は対象チケットが存在しない（または別ワークスペース / 別スペースのもの）
// ときに返す。テナント越えのアクセスは「無い」と同じ扱いにする（存在の有無を漏らさない）。
var ErrTicketNotFound = errors.New("ticket not found")

// ErrTicketStatusNotFound は対象の状態が存在しないときに返す。
var ErrTicketStatusNotFound = errors.New("ticket status not found")

// ErrTicketTypeNotFound は対象の種別が存在しないときに返す。
var ErrTicketTypeNotFound = errors.New("ticket type not found")

// ErrTicketsAlreadyEnabled は既に有効化済みのスペースへ再度有効化しようとしたときに返す
// （「有効化済み」の正本は「初期状態を持つ現役の状態が 1 つある」。設計 Ⅵ）。
var ErrTicketsAlreadyEnabled = errors.New("tickets already enabled for this space")

// ErrTicketStatusNameTaken / ErrTicketTypeNameTaken は現役の中で同名（大文字小文字を
// 区別しない）が既にあるときに返す（uq_ticket_statuses_space_name /
// uq_ticket_types_space_name の一意制約違反を repository が翻訳する）。
var (
	ErrTicketStatusNameTaken = errors.New("ticket status name is already taken")
	ErrTicketTypeNameTaken   = errors.New("ticket type name is already taken")
)

// ErrTicketAssigneeNotFound は担当に指定した principal が同じワークスペースに
// 実在しない（または kind='user' でない）ときに返す。
var ErrTicketAssigneeNotFound = errors.New("ticket assignee principal not found")

// TicketCreateInput は CreateTicket に渡す入力。
// Number は渡さない（採番 CTE が決める。呼び出し側が直接指定する経路は持たない —
// 設計 Ⅳ-B の「tickets への INSERT はこのクエリ 1 本だけ」というレビュー項目）。
// TicketCreateInput に ID は含めない。採番は repository（CreateTicket 実装）が
// InsertTicketStatus / InsertTicketType と同じ流儀で行う（呼び出し側は入れ物を渡さない）。
type TicketCreateInput struct {
	WorkspaceID     string
	SpaceID         string
	TypeID          string
	StatusID        string
	ParentID        *string
	Title           string
	Doc             []byte
	PlainText       string
	Priority        domain.TicketPriority
	StartDate       *string
	DueDate         *string
	Position        string
	CreatedByUserID uint64
}

// TicketUpdateFields は UpdateTicket が書き換える列（部分更新）。
// StatusID / ClosedAt / Resolution はここに含めない（状態変更は別の専用メソッドを通す。
// closed_at / resolution を category から必ず導く責務を 1 か所に閉じるため）。
type TicketUpdateFields struct {
	TypeID    string
	ParentID  *string
	Title     string
	Doc       []byte
	PlainText string
	Priority  domain.TicketPriority
	StartDate *string
	DueDate   *string
}

// TicketWithAssignee はチケット 1 件と、その担当（principals への参照）の組。
//
// 担当は別表（ticket_assignments）なので domain.Ticket には持たせない（あの型は
// tickets の 1 行を表す）。画面は一覧でも詳細でも担当を出すため、SQL 側の LEFT JOIN で
// 一緒に取り、この型で運ぶ。担当が居なければ AssigneePrincipalID は nil。
type TicketWithAssignee struct {
	Ticket              domain.Ticket
	AssigneePrincipalID *string
}

// ListTicketsInput は一覧の絞り込み条件。ゼロ値は「絞らない」を意味する。
type ListTicketsInput struct {
	WorkspaceID         string
	SpaceID             string
	IncludeArchived     bool
	StatusID            *string
	TypeID              *string
	AssigneePrincipalID *string
}

// TicketRepository はチケット（段 1: 骨格）の永続化を担う。
//
// 1 boundary = 1 fat interface（KnowledgeBaseRepository と同じ方針）。
//
// 権限判定はここに含めない。チケットの実効権限はページを介さない「スペース単位」の
// 判定（設計 Ⅳ-H）で、既存の KnowledgeBasePermissionRepository.SpacePermissionFactsForUser
// がそのまま使える（domain.ScopePermission / ScopeFacts / ResolveScopePermission は
// 段 0 から流用）。usecase/ticket が両方の repository を depend し、
// 「チケット→スペース」の解決だけをこちらの FindTicket に任せる。同じ判定ロジックを
// SQL に二重化しない。
type TicketRepository interface {
	// --- 状態・種別（管理画面） ---

	// HasActiveInitialTicketStatus は「有効化済み」の判定に使う
	// （初期状態を持つ現役の状態が 1 つでもあるか）。
	HasActiveInitialTicketStatus(ctx context.Context, workspaceID, spaceID string) (bool, error)
	InsertTicketStatus(ctx context.Context, s *domain.TicketStatus) error
	FindTicketStatus(ctx context.Context, workspaceID, spaceID, statusID string) (*domain.TicketStatus, error)
	ListTicketStatuses(ctx context.Context, workspaceID, spaceID string, includeArchived bool) ([]domain.TicketStatus, error)
	// GetInitialTicketStatus は新規チケット作成時の既定状態を解決する。
	GetInitialTicketStatus(ctx context.Context, workspaceID, spaceID string) (*domain.TicketStatus, error)
	UpdateTicketStatus(ctx context.Context, s *domain.TicketStatus) error
	SetTicketStatusInitial(ctx context.Context, workspaceID, spaceID, statusID string) error
	ArchiveTicketStatus(ctx context.Context, workspaceID, spaceID, statusID string) error
	RestoreTicketStatus(ctx context.Context, workspaceID, spaceID, statusID, position string) error
	CountActiveTicketsByStatus(ctx context.Context, workspaceID, spaceID, statusID string) (int64, error)
	LastActiveTicketStatusPosition(ctx context.Context, workspaceID, spaceID string) (string, error)

	InsertTicketType(ctx context.Context, t *domain.TicketType) error
	FindTicketType(ctx context.Context, workspaceID, spaceID, typeID string) (*domain.TicketType, error)
	ListTicketTypes(ctx context.Context, workspaceID, spaceID string, includeArchived bool) ([]domain.TicketType, error)
	// GetDefaultTicketType は新規チケット作成時の既定種別を解決する。
	GetDefaultTicketType(ctx context.Context, workspaceID, spaceID string) (*domain.TicketType, error)
	UpdateTicketType(ctx context.Context, t *domain.TicketType) error
	SetTicketTypeDefault(ctx context.Context, workspaceID, spaceID, typeID string) error
	ArchiveTicketType(ctx context.Context, workspaceID, spaceID, typeID string) error
	RestoreTicketType(ctx context.Context, workspaceID, spaceID, typeID, position string) error
	CountActiveTicketsByType(ctx context.Context, workspaceID, spaceID, typeID string) (int64, error)
	LastActiveTicketTypePosition(ctx context.Context, workspaceID, spaceID string) (string, error)

	// --- チケット本体 ---

	// CreateTicket は採番 CTE を含む 1 文で番号を払い出し、tickets へ 1 行作る。
	CreateTicket(ctx context.Context, in TicketCreateInput) (*domain.Ticket, error)
	// FindTicket はチケット 1 件を返す（担当は付かない）。権限判定・親子の検証など、
	// 「その行が在るか・どのスペースか」だけが要る内部用途に使う。
	// 画面へ返す取得は FindTicketWithAssignee を使う。
	FindTicket(ctx context.Context, workspaceID, ticketID string) (*domain.Ticket, error)
	// FindTicketWithAssignee は詳細画面向けにチケット 1 件と担当を 1 回の問い合わせで返す。
	FindTicketWithAssignee(ctx context.Context, workspaceID, ticketID string) (*TicketWithAssignee, error)
	// ResolveTicketIDByKey は spaceKey（小文字）+ number から ticket_id を引く
	// （domain.ParseTicketKey で分解した結果を渡す）。
	ResolveTicketIDByKey(ctx context.Context, workspaceID, spaceKey string, number int64) (string, error)
	ListTickets(ctx context.Context, in ListTicketsInput) ([]TicketWithAssignee, error)
	ListTicketChildren(ctx context.Context, workspaceID, spaceID, parentID string) ([]domain.Ticket, error)
	UpdateTicket(ctx context.Context, workspaceID, ticketID string, fields TicketUpdateFields) (*domain.Ticket, error)
	// ChangeTicketStatus は closedAt / resolution を usecase 側で
	// domain.ResolveTicketClosedFields から導出した値をそのまま書く（ここでは判断しない）。
	ChangeTicketStatus(
		ctx context.Context, workspaceID, ticketID, statusID string,
		closedAt *time.Time, resolution *domain.TicketResolution,
	) (*domain.Ticket, error)
	MoveTicket(ctx context.Context, workspaceID, ticketID, position string) error
	ArchiveTicket(ctx context.Context, workspaceID, ticketID string) error
	RestoreTicket(ctx context.Context, workspaceID, ticketID, position string) error
	CountActiveTicketChildren(ctx context.Context, workspaceID, ticketID string) (int64, error)
	// ListTicketParentChain は親を根まで辿った列（自分を含まない、根に近い順）を返す。
	// 深さの検査・周期の検出・レベル整合性の検査に使う（最大 3 段なので閉包表は持たない）。
	ListTicketParentChain(ctx context.Context, workspaceID, ticketID string) ([]domain.Ticket, error)
	LastActiveTicketPosition(ctx context.Context, workspaceID, spaceID string) (string, error)
	// HasActiveTicketPosition は move の before/after 指定チケットが、指定スペースの
	// 現役の兄弟であることを検証する。
	FindActiveTicketPosition(ctx context.Context, workspaceID, spaceID, ticketID string) (string, bool, error)

	// --- 担当 ---

	UpsertTicketAssignment(ctx context.Context, a *domain.TicketAssignment) error
	DeleteTicketAssignment(ctx context.Context, workspaceID, ticketID string) error
	FindTicketAssignment(ctx context.Context, workspaceID, ticketID string) (*domain.TicketAssignment, error)
	ListTicketsAssignedToPrincipal(ctx context.Context, workspaceID, principalID string) ([]domain.Ticket, error)

	// --- 変更履歴 ---

	// InsertTicketChangeGroup はグループと項目群を同一トランザクションで書く
	// （呼び出し側は usecase から TxManager.DoInTx で境界を引く）。
	InsertTicketChangeGroup(ctx context.Context, g *domain.TicketChangeGroup) error
	ListTicketChangeGroups(ctx context.Context, workspaceID, ticketID string) ([]domain.TicketChangeGroup, error)

	// --- 派生表（本文からの参照） ---

	// ReplaceTicketPageLinks / ReplaceTicketTicketLinks は本文保存のたびに張り替える。
	// targetIDs は実在確認前の候補で、ワークスペース内に実在しない ID は黙って除外する
	// （リンク切れ 1 本のために保存全体を失敗させない。ページ側と同じ方針）。
	ReplaceTicketPageLinks(ctx context.Context, workspaceID, sourceTicketID string, targetPageIDs []string) error
	ReplaceTicketTicketLinks(ctx context.Context, workspaceID, sourceTicketID string, targetTicketIDs []string) error
	ListTicketPageLinks(ctx context.Context, workspaceID, sourceTicketID string) ([]domain.TicketPageLink, error)
	ListPagesReferencingTicket(ctx context.Context, workspaceID, targetTicketID string) ([]domain.TicketPageLink, error)
	ListTicketTicketLinks(ctx context.Context, workspaceID, sourceTicketID string) ([]domain.TicketTicketLink, error)
	ListTicketsReferencingTicket(ctx context.Context, workspaceID, targetTicketID string) ([]domain.TicketTicketLink, error)
}
