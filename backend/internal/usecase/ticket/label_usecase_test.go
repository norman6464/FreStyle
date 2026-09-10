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

func Test_ラベル作成_不正な値は拒否(t *testing.T) {
	uc := ticket.NewCreateLabelUseCase(&mockLabelRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, ticket.CreateLabelInput{WorkspaceID: tkWS, SpaceID: tkSpace, Name: "", Color: "#2f6b47"})
	require.ErrorIs(t, err, domain.ErrInvalidLabelName, "空の名前は拒否")
	_, err = uc.Execute(ctx, ticket.CreateLabelInput{WorkspaceID: tkWS, SpaceID: tkSpace, Name: "  ", Color: "#2f6b47"})
	require.ErrorIs(t, err, domain.ErrInvalidLabelName, "空白だけの名前も拒否")
	_, err = uc.Execute(ctx, ticket.CreateLabelInput{WorkspaceID: tkWS, SpaceID: tkSpace, Name: "OK", Color: "url(evil)"})
	require.ErrorIs(t, err, domain.ErrInvalidLabelColor, "不正な色は拒否")
}

func Test_ラベル作成_名前をトリムし色を正規化してから保存する(t *testing.T) {
	repo := &mockLabelRepo{}
	var captured *domain.Label
	repo.On("CreateLabel", mock.Anything, mock.AnythingOfType("*domain.Label")).
		Run(func(args mock.Arguments) { captured = args.Get(1).(*domain.Label) }).Return(nil)

	_, err := ticket.NewCreateLabelUseCase(repo).Execute(context.Background(), ticket.CreateLabelInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, Name: "  緊急  ", Color: "#FF0000",
	})
	require.NoError(t, err)
	require.Equal(t, "緊急", captured.Name)
	require.Equal(t, "#ff0000", captured.Color)
}

func Test_ラベル更新_同名なら重複エラーをそのまま伝える(t *testing.T) {
	repo := &mockLabelRepo{}
	repo.On("FindLabel", mock.Anything, tkWS, "label-1").
		Return(&domain.Label{ID: "label-1", WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
	repo.On("UpdateLabel", mock.Anything, mock.AnythingOfType("*domain.Label")).
		Return(repository.ErrLabelNameTaken)

	_, err := ticket.NewUpdateLabelUseCase(repo).Execute(context.Background(), ticket.UpdateLabelInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, LabelID: "label-1", Name: "重複", Color: "#2f6b47",
	})
	require.ErrorIs(t, err, repository.ErrLabelNameTaken)
}

func Test_ラベル更新_違うスペースのラベルは拒否(t *testing.T) {
	repo := &mockLabelRepo{}
	repo.On("FindLabel", mock.Anything, tkWS, "label-1").
		Return(&domain.Label{ID: "label-1", WorkspaceID: tkWS, SpaceID: "other-space"}, nil)

	_, err := ticket.NewUpdateLabelUseCase(repo).Execute(context.Background(), ticket.UpdateLabelInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, LabelID: "label-1", Name: "改名", Color: "#2f6b47",
	})
	require.ErrorIs(t, err, repository.ErrLabelNotFound, "権限を確かめた相手と所属が違えば「無い」と同じ扱い")
	repo.AssertNotCalled(t, "UpdateLabel")
}

func Test_ラベル更新_同じスペースなら書き換える(t *testing.T) {
	repo := &mockLabelRepo{}
	repo.On("FindLabel", mock.Anything, tkWS, "label-1").
		Return(&domain.Label{ID: "label-1", WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
	var captured *domain.Label
	repo.On("UpdateLabel", mock.Anything, mock.AnythingOfType("*domain.Label")).
		Run(func(args mock.Arguments) { captured = args.Get(1).(*domain.Label) }).Return(nil)

	_, err := ticket.NewUpdateLabelUseCase(repo).Execute(context.Background(), ticket.UpdateLabelInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, LabelID: "label-1", Name: "改名", Color: "#2f6b47",
	})
	require.NoError(t, err)
	require.Equal(t, tkSpace, captured.SpaceID, "SQL 側でも絞れるようスペースを渡す")
}

func Test_チケットへのラベル付与_違うスペースのラベルは拒否(t *testing.T) {
	tickets := &mockTicketRepo{}
	tickets.On("FindTicket", mock.Anything, tkWS, tkTicket).
		Return(&domain.Ticket{ID: tkTicket, WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
	labels := &mockLabelRepo{}
	labels.On("FindLabel", mock.Anything, tkWS, "label-1").
		Return(&domain.Label{ID: "label-1", WorkspaceID: tkWS, SpaceID: "other-space"}, nil)

	err := ticket.NewAddTicketLabelUseCase(labels, tickets).Execute(context.Background(), ticket.AddTicketLabelInput{
		WorkspaceID: tkWS, TicketID: tkTicket, LabelID: "label-1",
	})
	require.ErrorIs(t, err, repository.ErrLabelNotFound, "違うスペースのラベルは「無い」と同じ扱い")
	labels.AssertNotCalled(t, "AddTicketLabel")
}

func Test_チケットへのラベル付与_同じスペースなら付ける(t *testing.T) {
	tickets := &mockTicketRepo{}
	tickets.On("FindTicket", mock.Anything, tkWS, tkTicket).
		Return(&domain.Ticket{ID: tkTicket, WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
	labels := &mockLabelRepo{}
	labels.On("FindLabel", mock.Anything, tkWS, "label-1").
		Return(&domain.Label{ID: "label-1", WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
	labels.On("AddTicketLabel", mock.Anything, tkWS, tkTicket, "label-1").Return(nil)

	err := ticket.NewAddTicketLabelUseCase(labels, tickets).Execute(context.Background(), ticket.AddTicketLabelInput{
		WorkspaceID: tkWS, TicketID: tkTicket, LabelID: "label-1",
	})
	require.NoError(t, err)
	labels.AssertExpectations(t)
}

func Test_チケットからのラベル除去_付いていなくても冪等に成功する(t *testing.T) {
	repo := &mockLabelRepo{}
	repo.On("RemoveTicketLabel", mock.Anything, tkWS, tkTicket, "label-1").Return(nil)

	err := ticket.NewRemoveTicketLabelUseCase(repo).Execute(context.Background(), ticket.RemoveTicketLabelInput{
		WorkspaceID: tkWS, TicketID: tkTicket, LabelID: "label-1",
	})
	require.NoError(t, err)
}

func Test_ラベル一覧_スペース単位で返す(t *testing.T) {
	repo := &mockLabelRepo{}
	want := []domain.Label{{ID: "label-1", Name: "緊急"}}
	repo.On("ListLabels", mock.Anything, tkWS, tkSpace).Return(want, nil)

	got, err := ticket.NewListLabelsUseCase(repo).Execute(context.Background(), tkWS, tkSpace)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func Test_チケット単位のラベル一覧(t *testing.T) {
	repo := &mockLabelRepo{}
	want := []domain.Label{{ID: "label-1", Name: "緊急"}}
	repo.On("ListLabelsByTicket", mock.Anything, tkWS, tkTicket).Return(want, nil)

	got, err := ticket.NewListLabelsForTicketUseCase(repo).Execute(context.Background(), tkWS, tkTicket)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func Test_チケットID群のラベル一括取得(t *testing.T) {
	repo := &mockLabelRepo{}
	want := map[string][]domain.Label{tkTicket: {{ID: "label-1", Name: "緊急"}}}
	repo.On("ListLabelsByTicketIDs", mock.Anything, tkWS, []string{tkTicket}).Return(want, nil)

	got, err := ticket.NewListLabelsByTicketIDsUseCase(repo).Execute(context.Background(), tkWS, []string{tkTicket})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func Test_ラベル削除(t *testing.T) {
	repo := &mockLabelRepo{}
	repo.On("FindLabel", mock.Anything, tkWS, "label-1").
		Return(&domain.Label{ID: "label-1", WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
	repo.On("DeleteLabel", mock.Anything, tkWS, tkSpace, "label-1").Return(nil)

	err := ticket.NewDeleteLabelUseCase(repo).Execute(context.Background(), tkWS, tkSpace, "label-1")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func Test_ラベル削除_違うスペースのラベルは拒否(t *testing.T) {
	repo := &mockLabelRepo{}
	repo.On("FindLabel", mock.Anything, tkWS, "label-1").
		Return(&domain.Label{ID: "label-1", WorkspaceID: tkWS, SpaceID: "other-space"}, nil)

	err := ticket.NewDeleteLabelUseCase(repo).Execute(context.Background(), tkWS, tkSpace, "label-1")
	require.ErrorIs(t, err, repository.ErrLabelNotFound)
	repo.AssertNotCalled(t, "DeleteLabel")
}
