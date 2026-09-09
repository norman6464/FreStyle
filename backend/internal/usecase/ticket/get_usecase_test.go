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

func Test_チケット取得_必須項目の検証(t *testing.T) {
	uc := ticket.NewGetTicketUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), ticket.GetTicketInput{TicketID: tkTicket})
	require.Error(t, err)
	_, err = uc.Execute(context.Background(), ticket.GetTicketInput{WorkspaceID: tkWS})
	require.Error(t, err)
}

func Test_チケット取得_存在しなければそのまま伝える(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).Return(nil, repository.ErrTicketNotFound)

	_, err := ticket.NewGetTicketUseCase(repo).Execute(context.Background(), ticket.GetTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket,
	})
	require.ErrorIs(t, err, repository.ErrTicketNotFound)
}

func Test_チケット取得_そのまま返す(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).
		Return(&domain.Ticket{
			ID: tkTicket, WorkspaceID: tkWS, SpaceID: tkSpace,
			Doc: []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"本文"}]}]}`),
		}, nil)

	got, err := ticket.NewGetTicketUseCase(repo).Execute(context.Background(), ticket.GetTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket,
	})
	require.NoError(t, err)
	assert.Equal(t, tkTicket, got.ID)
	assert.Contains(t, string(got.Doc), "本文")
}
