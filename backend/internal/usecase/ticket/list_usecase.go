package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListTicketsUseCase はスペース内のチケット一覧を返す（position 順）。担当は同じ
// 問い合わせの LEFT JOIN で一緒に返る（画面が一覧でも担当を出すため）。
// 絞り込み条件はそのまま repository へ渡す（畳み方の規則を持たない、薄い層）。
type ListTicketsUseCase struct {
	repo repository.TicketRepository
}

func NewListTicketsUseCase(r repository.TicketRepository) *ListTicketsUseCase {
	return &ListTicketsUseCase{repo: r}
}

type ListTicketsInput struct {
	WorkspaceID         string
	SpaceID             string
	IncludeArchived     bool
	StatusID            *string
	TypeID              *string
	AssigneePrincipalID *string
}

func (u *ListTicketsUseCase) Execute(ctx context.Context, in ListTicketsInput) ([]repository.TicketWithAssignee, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	return u.repo.ListTickets(ctx, repository.ListTicketsInput{
		WorkspaceID:         in.WorkspaceID,
		SpaceID:             in.SpaceID,
		IncludeArchived:     in.IncludeArchived,
		StatusID:            in.StatusID,
		TypeID:              in.TypeID,
		AssigneePrincipalID: in.AssigneePrincipalID,
	})
}
