package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetTicketUseCase はチケット 1 件を取得する。
//
// 本文（Doc）は保存時に pageRef / ticketRef の title を剥がしてあるので、そのまま返す
// （title=null のまま）。参照先の現在の題名を読み手ごとに解決して埋め込む処理は
// 段 1 の対象外（画面側は id だけを見て遅延解決する想定。ページの
// rewritePageRefTitles に相当する処理は、実際に画面が必要とする段で追加する）。
type GetTicketUseCase struct {
	repo repository.TicketRepository
}

func NewGetTicketUseCase(r repository.TicketRepository) *GetTicketUseCase {
	return &GetTicketUseCase{repo: r}
}

type GetTicketInput struct {
	WorkspaceID string
	TicketID    string
}

func (u *GetTicketUseCase) Execute(ctx context.Context, in GetTicketInput) (*domain.Ticket, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.TicketID == "" {
		return nil, errors.New("ticketID is required")
	}
	return u.repo.FindTicket(ctx, in.WorkspaceID, in.TicketID)
}
