package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListTicketStatusesUseCase はスペースの状態一覧を返す（管理画面・作成フォームの選択肢）。
// 畳み方の規則を持たない薄い層（ListTicketsUseCase と同じ形）。
type ListTicketStatusesUseCase struct {
	repo repository.TicketRepository
}

func NewListTicketStatusesUseCase(r repository.TicketRepository) *ListTicketStatusesUseCase {
	return &ListTicketStatusesUseCase{repo: r}
}

type ListTicketStatusesInput struct {
	WorkspaceID     string
	SpaceID         string
	IncludeArchived bool
}

func (u *ListTicketStatusesUseCase) Execute(ctx context.Context, in ListTicketStatusesInput) ([]domain.TicketStatus, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	return u.repo.ListTicketStatuses(ctx, in.WorkspaceID, in.SpaceID, in.IncludeArchived)
}

// ListTicketTypesUseCase はスペースの種別一覧を返す。
type ListTicketTypesUseCase struct {
	repo repository.TicketRepository
}

func NewListTicketTypesUseCase(r repository.TicketRepository) *ListTicketTypesUseCase {
	return &ListTicketTypesUseCase{repo: r}
}

type ListTicketTypesInput struct {
	WorkspaceID     string
	SpaceID         string
	IncludeArchived bool
}

func (u *ListTicketTypesUseCase) Execute(ctx context.Context, in ListTicketTypesInput) ([]domain.TicketType, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	return u.repo.ListTicketTypes(ctx, in.WorkspaceID, in.SpaceID, in.IncludeArchived)
}
