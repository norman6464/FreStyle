package ticket_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_チケット一覧_必須項目の検証(t *testing.T) {
	uc := ticket.NewListTicketsUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), ticket.ListTicketsInput{})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(context.Background(), ticket.ListTicketsInput{WorkspaceID: tkWS})
	require.Error(t, err, "spaceID 必須")
}

func Test_チケット一覧_絞り込みをそのままrepositoryへ渡す(t *testing.T) {
	repo := &mockTicketRepo{}
	statusID := "status-1"
	repo.On("ListTickets", mock.Anything, repository.ListTicketsInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, IncludeArchived: false, StatusID: &statusID,
	}).Return([]domain.Ticket{{ID: "t1"}, {ID: "t2"}}, nil)

	got, err := ticket.NewListTicketsUseCase(repo).Execute(context.Background(), ticket.ListTicketsInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, StatusID: &statusID,
	})
	require.NoError(t, err)
	assert.Len(t, got, 2)
}
