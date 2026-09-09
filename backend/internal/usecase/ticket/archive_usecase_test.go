package ticket_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_チケットアーカイブ_必須項目の検証(t *testing.T) {
	uc := ticket.NewArchiveTicketUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), ticket.ArchiveTicketInput{TicketID: tkTicket, ActorUserID: 1})
	require.Error(t, err)
}

func Test_チケットアーカイブ_履歴を残す(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("ArchiveTicket", mock.Anything, tkWS, tkTicket).Return(nil)
	repo.On("InsertTicketChangeGroup", mock.Anything, mock.MatchedBy(func(g *domain.TicketChangeGroup) bool {
		return g.WorkspaceID == tkWS && g.TicketID == tkTicket && g.ActorUserID == 1 &&
			len(g.Items) == 1 && g.Items[0].Field == domain.TicketChangeFieldArchived
	})).Return(nil)
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).
		Return(&domain.Ticket{ID: tkTicket, WorkspaceID: tkWS}, nil)

	_, err := ticket.NewArchiveTicketUseCase(repo).Execute(context.Background(), ticket.ArchiveTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
	})
	require.NoError(t, err)
}

func Test_チケット復元_positionを末尾へ付け直す(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).
		Return(&domain.Ticket{ID: tkTicket, WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
	repo.On("LastActiveTicketPosition", mock.Anything, tkWS, tkSpace).Return("a0", nil)
	repo.On("RestoreTicket", mock.Anything, tkWS, tkTicket, mock.MatchedBy(func(pos string) bool {
		return pos > "a0"
	})).Return(nil)
	repo.On("InsertTicketChangeGroup", mock.Anything, mock.AnythingOfType("*domain.TicketChangeGroup")).Return(nil)

	_, err := ticket.NewRestoreTicketUseCase(repo).Execute(context.Background(), ticket.RestoreTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
	})
	require.NoError(t, err)
}

func Test_チケットアーカイブ_存在しなければそのまま伝える(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("ArchiveTicket", mock.Anything, tkWS, tkTicket).Return(repository.ErrTicketNotFound)

	_, err := ticket.NewArchiveTicketUseCase(repo).Execute(context.Background(), ticket.ArchiveTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
	})
	require.ErrorIs(t, err, repository.ErrTicketNotFound)
	repo.AssertNotCalled(t, "InsertTicketChangeGroup")
}
