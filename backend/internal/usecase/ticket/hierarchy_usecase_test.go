package ticket_test

import (
	"context"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_チケット祖先列_閉包表をそのまま返す(t *testing.T) {
	repo := &mockTicketRepo{}
	want := []domain.Ticket{{ID: "root"}, {ID: "child"}}
	repo.On("ListTicketAncestors", mock.Anything, tkWS, tkTicket).Return(want, nil)

	got, err := ticket.NewListTicketAncestorsUseCase(repo).Execute(context.Background(), tkWS, tkTicket)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func Test_チケット祖先列_必須項目の検証(t *testing.T) {
	uc := ticket.NewListTicketAncestorsUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), "", tkTicket)
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(context.Background(), tkWS, "")
	require.Error(t, err, "ticketID 必須")
}

func Test_ページの逆参照一覧_リポジトリの結果をそのまま返す(t *testing.T) {
	repo := &mockTicketRepo{}
	want := []domain.Ticket{{ID: "ticket-1"}}
	repo.On("ListTicketsReferencingPage", mock.Anything, tkWS, "page-1").Return(want, nil)

	got, err := ticket.NewListTicketsReferencingPageUseCase(repo).Execute(context.Background(), tkWS, "page-1")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func Test_ページの逆参照一覧_必須項目の検証(t *testing.T) {
	uc := ticket.NewListTicketsReferencingPageUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), "", "page-1")
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(context.Background(), tkWS, "")
	require.Error(t, err, "pageID 必須")
}
