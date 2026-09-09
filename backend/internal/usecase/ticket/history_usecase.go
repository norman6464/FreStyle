package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListTicketHistoryUseCase はチケットの変更履歴を新しい順に返す。
type ListTicketHistoryUseCase struct {
	repo repository.TicketRepository
}

func NewListTicketHistoryUseCase(r repository.TicketRepository) *ListTicketHistoryUseCase {
	return &ListTicketHistoryUseCase{repo: r}
}

type ListTicketHistoryInput struct {
	WorkspaceID string
	TicketID    string
}

func (u *ListTicketHistoryUseCase) Execute(ctx context.Context, in ListTicketHistoryInput) ([]domain.TicketChangeGroup, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.TicketID == "" {
		return nil, errors.New("ticketID is required")
	}
	return u.repo.ListTicketChangeGroups(ctx, in.WorkspaceID, in.TicketID)
}
