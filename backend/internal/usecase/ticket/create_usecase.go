package ticket

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/pkg/fracindex"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// TicketMaxDepth はチケットの親子関係の最大の深さ（設計 Ⅳ-D）。ルート自身が 1。
// 最大 3 段なので閉包表は持たず、ListTicketParentChain で親を辿って数える。
const TicketMaxDepth = 3

// CreateTicketUseCase はスペース直下または親チケットの下に新しいチケットを作る。
type CreateTicketUseCase struct {
	repo repository.TicketRepository
}

func NewCreateTicketUseCase(r repository.TicketRepository) *CreateTicketUseCase {
	return &CreateTicketUseCase{repo: r}
}

type CreateTicketInput struct {
	WorkspaceID string
	SpaceID     string
	// TypeID / StatusID が空なら既定（GetDefaultTicketType / GetInitialTicketStatus）を解決する。
	TypeID   string
	StatusID string
	// ParentID が nil ならトップレベル。
	ParentID        *string
	Title           string
	Doc             string
	Priority        domain.TicketPriority
	StartDate       *string
	DueDate         *string
	CreatedByUserID uint64
}

func (u *CreateTicketUseCase) Execute(ctx context.Context, in CreateTicketInput) (*domain.Ticket, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, domain.ErrInvalidTicketName
	}
	if in.CreatedByUserID == 0 {
		return nil, errors.New("createdByUserID is required")
	}
	if !domain.ValidTicketDateOrder(in.StartDate, in.DueDate) {
		return nil, domain.ErrTicketDateRangeInverted
	}

	typ, err := u.resolveType(ctx, in.WorkspaceID, in.SpaceID, in.TypeID)
	if err != nil {
		return nil, err
	}
	status, err := u.resolveStatus(ctx, in.WorkspaceID, in.SpaceID, in.StatusID)
	if err != nil {
		return nil, err
	}
	if err := u.validateParent(ctx, in.WorkspaceID, in.SpaceID, in.ParentID, typ.HierarchyLevel); err != nil {
		return nil, err
	}

	// tickets.position（NOT NULL を満たすためだけの列。以後読み手が居ない）と ticket_ranks.position
	// （並び順の正本）は独立に計算する — 以後の Move は ticket_ranks だけを更新するため（設計 Ⅳ-F）。
	last, err := u.repo.LastActiveTicketPosition(ctx, in.WorkspaceID, in.SpaceID)
	if err != nil {
		return nil, err
	}
	pos, err := fracindex.Between(last, "")
	if err != nil {
		return nil, err
	}
	lastRank, err := u.repo.LastActiveTicketRankPosition(ctx, in.WorkspaceID, in.SpaceID)
	if err != nil {
		return nil, err
	}
	rankPos, err := fracindex.Between(lastRank, "")
	if err != nil {
		return nil, err
	}

	stripped, err := StripDocRefTitles([]byte(in.Doc))
	if err != nil {
		return nil, fmt.Errorf("invalid doc: %w", err)
	}
	plainText := BuildPlainText(stripped)
	pageIDs, ticketIDs, err := ExtractDocRefs(stripped)
	if err != nil {
		return nil, fmt.Errorf("invalid doc: %w", err)
	}

	priority := in.Priority
	if priority == 0 {
		priority = domain.TicketPriorityDefault
	}

	created, err := u.repo.CreateTicket(ctx, repository.TicketCreateInput{
		WorkspaceID:     in.WorkspaceID,
		SpaceID:         in.SpaceID,
		TypeID:          typ.ID,
		StatusID:        status.ID,
		ParentID:        in.ParentID,
		Title:           in.Title,
		Doc:             stripped,
		PlainText:       plainText,
		Priority:        priority,
		StartDate:       in.StartDate,
		DueDate:         in.DueDate,
		Position:        pos,
		CreatedByUserID: in.CreatedByUserID,
	})
	if err != nil {
		return nil, err
	}

	// ticket_ranks へ並び順の正本を作る（設計 Ⅳ-F）。created.Position を rankPos で上書きするのは
	// GetTicket が返す position（ticket_ranks 由来）と作成直後の応答を一致させるため。
	if err := u.repo.InsertTicketRank(ctx, in.WorkspaceID, created.ID, rankPos); err != nil {
		return nil, err
	}
	created.Position = rankPos

	// ticket_paths（閉包表・段 5）: 自己参照行（depth=0）は常に張り、親があれば祖先集合を +1 して
	// 引き継ぐ（page_paths の CreatePage と同じ考え方）。
	if err := u.repo.InsertTicketPathSelf(ctx, in.WorkspaceID, created.ID); err != nil {
		return nil, err
	}
	if in.ParentID != nil {
		if err := u.repo.InsertTicketPathAncestors(ctx, in.WorkspaceID, created.ID, *in.ParentID); err != nil {
			return nil, err
		}
	}

	// 派生表（本文からの参照）は作成直後に張る。空スライスでも Replace を呼び「参照 0 件」を明示し、
	// UpdateTicket と同じ経路に揃える。
	if err := u.repo.ReplaceTicketPageLinks(ctx, in.WorkspaceID, created.ID, pageIDs); err != nil {
		return nil, err
	}
	if err := u.repo.ReplaceTicketTicketLinks(ctx, in.WorkspaceID, created.ID, ticketIDs); err != nil {
		return nil, err
	}
	created.Doc = stripped
	created.PlainText = plainText
	return created, nil
}

func (u *CreateTicketUseCase) resolveType(ctx context.Context, workspaceID, spaceID, typeID string) (*domain.TicketType, error) {
	if typeID == "" {
		return u.repo.GetDefaultTicketType(ctx, workspaceID, spaceID)
	}
	return u.repo.FindTicketType(ctx, workspaceID, spaceID, typeID)
}

func (u *CreateTicketUseCase) resolveStatus(ctx context.Context, workspaceID, spaceID, statusID string) (*domain.TicketStatus, error) {
	if statusID == "" {
		return u.repo.GetInitialTicketStatus(ctx, workspaceID, spaceID)
	}
	return u.repo.FindTicketStatus(ctx, workspaceID, spaceID, statusID)
}

// validateParent は親チケットが指定されたときだけ、実在・同一スペース・階層規則
// （設計 Ⅳ-D: 子の段 <= 親の段、同段の親子は 0 だけ、-1 は親になれない）・深さ最大 3 を検証する。
func (u *CreateTicketUseCase) validateParent(
	ctx context.Context, workspaceID, spaceID string, parentID *string, childLevel int,
) error {
	if parentID == nil {
		return nil
	}
	parent, err := u.repo.FindTicket(ctx, workspaceID, *parentID)
	if err != nil {
		return err
	}
	if parent.SpaceID != spaceID {
		// スペースをまたぐ親子は作れない。存在しない ID と同じ応答に畳み、どちらだったかを
		// 漏らさない（ページの木と同じ方針）。
		return repository.ErrTicketNotFound
	}
	parentType, err := u.repo.FindTicketType(ctx, workspaceID, spaceID, parent.TypeID)
	if err != nil {
		return err
	}
	if err := domain.ValidateTicketParentChild(childLevel, parentType.HierarchyLevel); err != nil {
		return err
	}
	chain, err := u.repo.ListTicketParentChain(ctx, workspaceID, *parentID)
	if err != nil {
		return err
	}
	// chain は親自身を含まない祖先列。親の深さ = len(chain)+1、新しい子の深さは
	// さらに +1。これが TicketMaxDepth を超えるなら拒否。
	childDepth := len(chain) + 2
	if childDepth > TicketMaxDepth {
		return domain.ErrTicketHierarchyRejected
	}
	return nil
}
