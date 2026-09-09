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
	}).Return([]repository.TicketWithAssignee{
		{Ticket: domain.Ticket{ID: "t1"}},
		{Ticket: domain.Ticket{ID: "t2"}, AssigneePrincipalID: &statusID},
	}, nil)

	got, err := ticket.NewListTicketsUseCase(repo).Execute(context.Background(), ticket.ListTicketsInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, StatusID: &statusID,
	})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "t1", got[0].Ticket.ID)
	assert.Nil(t, got[0].AssigneePrincipalID, "担当が居なければ nil のまま運ぶ")
	require.NotNil(t, got[1].AssigneePrincipalID)
}
