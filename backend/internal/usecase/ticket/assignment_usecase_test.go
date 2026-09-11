package ticket_test

import (
	"context"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/norman6464/frestyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const tkPrincipal = "01a00000-0000-7000-8000-000000000050"

func Test_チケット担当設定_必須項目の検証(t *testing.T) {
	uc := ticket.NewAssignTicketUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), ticket.AssignTicketInput{TicketID: tkTicket, AssigneePrincipalID: tkPrincipal, AssignedByUserID: 1})
	require.Error(t, err)
	_, err = uc.Execute(context.Background(), ticket.AssignTicketInput{WorkspaceID: tkWS, AssigneePrincipalID: tkPrincipal, AssignedByUserID: 1})
	require.Error(t, err)
	_, err = uc.Execute(context.Background(), ticket.AssignTicketInput{WorkspaceID: tkWS, TicketID: tkTicket, AssignedByUserID: 1})
	require.Error(t, err)
}

// principals（ワークスペース所属の正本）への複合 FK が実際の境界を守る（設計 Ⅳ-G）。
// ここでは repository が返す ErrTicketAssigneeNotFound をそのまま usecase が伝えることだけ確認する。
func Test_チケット担当設定_別ワークスペースの主体なら404相当(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("FindTicketAssignment", mock.Anything, tkWS, tkTicket).Return(nil, nil)
	repo.On("UpsertTicketAssignment", mock.Anything, mock.AnythingOfType("*domain.TicketAssignment")).
		Return(repository.ErrTicketAssigneeNotFound)

	_, err := ticket.NewAssignTicketUseCase(repo).Execute(context.Background(), ticket.AssignTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, AssigneePrincipalID: tkPrincipal, AssignedByUserID: 1,
	})
	require.ErrorIs(t, err, repository.ErrTicketAssigneeNotFound)
}

func Test_チケット担当設定_履歴を残す(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("FindTicketAssignment", mock.Anything, tkWS, tkTicket).Return(nil, nil)
	repo.On("UpsertTicketAssignment", mock.Anything, mock.AnythingOfType("*domain.TicketAssignment")).Return(nil)
	repo.On("InsertTicketChangeGroup", mock.Anything, mock.MatchedBy(func(g *domain.TicketChangeGroup) bool {
		return len(g.Items) == 1 && g.Items[0].Field == domain.TicketChangeFieldAssignee &&
			g.Items[0].OldValue == nil && *g.Items[0].NewValue == tkPrincipal
	})).Return(nil)

	_, err := ticket.NewAssignTicketUseCase(repo).Execute(context.Background(), ticket.AssignTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, AssigneePrincipalID: tkPrincipal, AssignedByUserID: 1,
	})
	require.NoError(t, err)
}

func Test_チケット担当解除_履歴を残す(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("FindTicketAssignment", mock.Anything, tkWS, tkTicket).
		Return(&domain.TicketAssignment{WorkspaceID: tkWS, TicketID: tkTicket, AssigneePrincipalID: tkPrincipal}, nil)
	repo.On("DeleteTicketAssignment", mock.Anything, tkWS, tkTicket).Return(nil)
	repo.On("InsertTicketChangeGroup", mock.Anything, mock.MatchedBy(func(g *domain.TicketChangeGroup) bool {
		return len(g.Items) == 1 && g.Items[0].Field == domain.TicketChangeFieldAssignee &&
			*g.Items[0].OldValue == tkPrincipal && g.Items[0].NewValue == nil
	})).Return(nil)

	err := ticket.NewUnassignTicketUseCase(repo).Execute(context.Background(), ticket.UnassignTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
	})
	require.NoError(t, err)
}

func Test_チケット担当解除_既に無ければ何もしない(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("FindTicketAssignment", mock.Anything, tkWS, tkTicket).Return(nil, nil)

	err := ticket.NewUnassignTicketUseCase(repo).Execute(context.Background(), ticket.UnassignTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
	})
	require.NoError(t, err)
	repo.AssertNotCalled(t, "DeleteTicketAssignment")
	repo.AssertNotCalled(t, "InsertTicketChangeGroup")
}
