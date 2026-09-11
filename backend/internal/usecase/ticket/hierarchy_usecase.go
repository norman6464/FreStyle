package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// ListTicketAncestorsUseCase はチケットのパンくず（根から順の祖先列）を返す
// （ticket_paths の閉包表を読む。段 5）。
//
// ページ側の ListViewableAncestorsUseCase と違い、祖先ごとの可視判定は行わない。
// チケットの親は常に同一スペース限定（fk_tickets_parent）なので、このチケット自体が
// 見えている（handler が requireTicketPermission を先に通している）なら、同じスペースの
// 祖先もすべて見える（設計 Ⅳ-H: チケットの実効権限はスペース単位）。
type ListTicketAncestorsUseCase struct {
	repo repository.TicketRepository
}

func NewListTicketAncestorsUseCase(r repository.TicketRepository) *ListTicketAncestorsUseCase {
	return &ListTicketAncestorsUseCase{repo: r}
}

func (u *ListTicketAncestorsUseCase) Execute(ctx context.Context, workspaceID, ticketID string) ([]domain.Ticket, error) {
	if workspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if ticketID == "" {
		return nil, errors.New("ticketID is required")
	}
	return u.repo.ListTicketAncestors(ctx, workspaceID, ticketID)
}

// ListTicketsReferencingPageUseCase はページ詳細の逆参照一覧が使う（そのページを本文中の
// pageRef で参照しているチケット一覧。段 5）。
//
// チケットには pages のような個票の権限が無い（設計 Ⅳ-H）ため、ここではスペース単位の
// 可視判定を行わない — 候補チケットをそのまま返し、どのスペースを見せてよいかの判定は
// handler 側（複数スペースにまたがりうるので、候補の空間集合を CheckSpacePermissionUseCase で
// 確かめる）に委ねる。usecase/ticket は usecase/kb を import できない（サブパッケージ同士は
// import しない規約）ため、この分担は handler 層でのみ組める。
type ListTicketsReferencingPageUseCase struct {
	repo repository.TicketRepository
}

func NewListTicketsReferencingPageUseCase(r repository.TicketRepository) *ListTicketsReferencingPageUseCase {
	return &ListTicketsReferencingPageUseCase{repo: r}
}

func (u *ListTicketsReferencingPageUseCase) Execute(ctx context.Context, workspaceID, pageID string) ([]domain.Ticket, error) {
	if workspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if pageID == "" {
		return nil, errors.New("pageID is required")
	}
	return u.repo.ListTicketsReferencingPage(ctx, workspaceID, pageID)
}
