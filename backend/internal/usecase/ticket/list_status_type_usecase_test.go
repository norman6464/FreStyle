package ticket_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_状態一覧_必須項目の検証(t *testing.T) {
	uc := ticket.NewListTicketStatusesUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), ticket.ListTicketStatusesInput{})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(context.Background(), ticket.ListTicketStatusesInput{WorkspaceID: tkWS})
	require.Error(t, err, "spaceID 必須")
}

func Test_状態一覧_そのままrepositoryへ渡す(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("ListTicketStatuses", mock.Anything, tkWS, tkSpace, true).
		Return([]domain.TicketStatus{{ID: "s1"}, {ID: "s2"}}, nil)

	got, err := ticket.NewListTicketStatusesUseCase(repo).Execute(context.Background(), ticket.ListTicketStatusesInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, IncludeArchived: true,
	})
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func Test_種別一覧_必須項目の検証(t *testing.T) {
	uc := ticket.NewListTicketTypesUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), ticket.ListTicketTypesInput{})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(context.Background(), ticket.ListTicketTypesInput{WorkspaceID: tkWS})
	require.Error(t, err, "spaceID 必須")
}

func Test_種別一覧_そのままrepositoryへ渡す(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("ListTicketTypes", mock.Anything, tkWS, tkSpace, false).
		Return([]domain.TicketType{{ID: "t1"}}, nil)

	got, err := ticket.NewListTicketTypesUseCase(repo).Execute(context.Background(), ticket.ListTicketTypesInput{
		WorkspaceID: tkWS, SpaceID: tkSpace,
	})
	require.NoError(t, err)
	assert.Len(t, got, 1)
}
