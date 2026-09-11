package ticket_test

import (
	"context"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_チケット履歴_必須項目の検証(t *testing.T) {
	uc := ticket.NewListTicketHistoryUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), ticket.ListTicketHistoryInput{TicketID: tkTicket})
	require.Error(t, err)
}

func Test_チケット履歴_そのまま返す(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("ListTicketChangeGroups", mock.Anything, tkWS, tkTicket).
		Return([]domain.TicketChangeGroup{{ID: "g1"}, {ID: "g2"}}, nil)

	got, err := ticket.NewListTicketHistoryUseCase(repo).Execute(context.Background(), ticket.ListTicketHistoryInput{
		WorkspaceID: tkWS, TicketID: tkTicket,
	})
	require.NoError(t, err)
	assert.Len(t, got, 2)
}
