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

const tkParentConst = "01a00000-0000-7000-8000-000000000010"

var tkParent = tkParentConst

func tkDefaultType() domain.TicketType {
	return domain.TicketType{ID: "01a00000-0000-7000-8000-000000000020", WorkspaceID: tkWS, SpaceID: tkSpace, Name: "タスク", HierarchyLevel: 0, IsDefault: true}
}

func tkInitialStatus() domain.TicketStatus {
	return domain.TicketStatus{ID: "01a00000-0000-7000-8000-000000000030", WorkspaceID: tkWS, SpaceID: tkSpace, Name: "To Do", Category: domain.TicketStatusCategoryTodo, IsInitial: true}
}

func Test_チケット作成_必須項目の検証(t *testing.T) {
	uc := ticket.NewCreateTicketUseCase(&mockTicketRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, ticket.CreateTicketInput{SpaceID: tkSpace, Title: "x", CreatedByUserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, ticket.CreateTicketInput{WorkspaceID: tkWS, Title: "x", CreatedByUserID: 1})
	require.Error(t, err, "spaceID 必須")
	_, err = uc.Execute(ctx, ticket.CreateTicketInput{WorkspaceID: tkWS, SpaceID: tkSpace, CreatedByUserID: 1})
	require.ErrorIs(t, err, domain.ErrInvalidTicketName, "title 必須")
	_, err = uc.Execute(ctx, ticket.CreateTicketInput{WorkspaceID: tkWS, SpaceID: tkSpace, Title: "  "})
	require.ErrorIs(t, err, domain.ErrInvalidTicketName, "空白だけの title は拒否")
	_, err = uc.Execute(ctx, ticket.CreateTicketInput{WorkspaceID: tkWS, SpaceID: tkSpace, Title: "x"})
	require.Error(t, err, "createdByUserID 必須")
}

// 開始日が期限より後なら、DB の CHECK（ck_tickets_dates_ordered）に到達させず
// ここで断る（素の Postgres エラーで 500 になるのを避けるため）。
func Test_チケット作成_開始日が期限より後なら拒否(t *testing.T) {
	uc := ticket.NewCreateTicketUseCase(&mockTicketRepo{})
	start, due := "2026-09-10", "2026-09-01"

	_, err := uc.Execute(context.Background(), ticket.CreateTicketInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, Title: "x", CreatedByUserID: 1,
		StartDate: &start, DueDate: &due,
	})
	require.ErrorIs(t, err, domain.ErrTicketDateRangeInverted)
}

// type/status を指定しなければ既定（GetDefaultTicketType / GetInitialTicketStatus）を解決する。
// position は末尾（LastActiveTicketPosition から fracindex.Between で採番）に置く。
func Test_チケット作成_既定の種別と状態を解決して作る(t *testing.T) {
	repo := &mockTicketRepo{}
	defaultType := tkDefaultType()
	initialStatus := tkInitialStatus()
	repo.On("GetDefaultTicketType", mock.Anything, tkWS, tkSpace).Return(&defaultType, nil)
	repo.On("GetInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(&initialStatus, nil)
	repo.On("LastActiveTicketPosition", mock.Anything, tkWS, tkSpace).Return("a0", nil)
	repo.On("LastActiveTicketRankPosition", mock.Anything, tkWS, tkSpace).Return("a0", nil)

	var captured repository.TicketCreateInput
	repo.On("CreateTicket", mock.Anything, mock.AnythingOfType("repository.TicketCreateInput")).
		Run(func(args mock.Arguments) { captured = args.Get(1).(repository.TicketCreateInput) }).
		Return(&domain.Ticket{ID: "01a00000-0000-7000-8000-000000000040", WorkspaceID: tkWS, SpaceID: tkSpace, Number: 1}, nil)
	repo.On("InsertTicketRank", mock.Anything, tkWS, "01a00000-0000-7000-8000-000000000040", mock.AnythingOfType("string")).Return(nil)
	repo.On("InsertTicketPathSelf", mock.Anything, tkWS, "01a00000-0000-7000-8000-000000000040").Return(nil)
	repo.On("ReplaceTicketPageLinks", mock.Anything, tkWS, mock.Anything, []string(nil)).Return(nil)
	repo.On("ReplaceTicketTicketLinks", mock.Anything, tkWS, mock.Anything, []string(nil)).Return(nil)

	got, err := ticket.NewCreateTicketUseCase(repo).Execute(context.Background(), ticket.CreateTicketInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, Title: "新しいチケット",
		Doc: `{"type":"doc","content":[]}`, CreatedByUserID: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, "01a00000-0000-7000-8000-000000000040", got.ID)
	assert.Equal(t, defaultType.ID, captured.TypeID)
	assert.Equal(t, initialStatus.ID, captured.StatusID)
	assert.Equal(t, domain.TicketPriorityDefault, captured.Priority)
	assert.Nil(t, captured.ParentID)
	assert.Greater(t, captured.Position, "a0", "既存の末尾より後ろに置く")
	assert.Greater(t, got.Position, "a0", "応答の position は ticket_ranks 由来の値に上書きされる")
}

func Test_チケット作成_既定の種別が無ければ拒否(t *testing.T) {
	// 有効化されていないスペースでは GetDefaultTicketType が ErrTicketTypeNotFound を返す。
	repo := &mockTicketRepo{}
	repo.On("GetDefaultTicketType", mock.Anything, tkWS, tkSpace).Return(nil, repository.ErrTicketTypeNotFound)

	_, err := ticket.NewCreateTicketUseCase(repo).Execute(context.Background(), ticket.CreateTicketInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, Title: "x", Doc: `{"type":"doc","content":[]}`, CreatedByUserID: 1,
	})
	require.ErrorIs(t, err, repository.ErrTicketTypeNotFound)
}

// 親を指定するときは、同じスペースに実在し、種別の階層規則（Ⅳ-D）を満たすことを検証する。
func Test_チケット作成_親を指定する場合の階層規則(t *testing.T) {
	t.Run("親のレベルより深い子は拒否", func(t *testing.T) {
		repo := &mockTicketRepo{}
		defaultType := tkDefaultType() // level 0
		initialStatus := tkInitialStatus()
		parentType := domain.TicketType{ID: "type-sub", HierarchyLevel: -1}
		parent := domain.Ticket{ID: tkParent, WorkspaceID: tkWS, SpaceID: tkSpace, TypeID: "type-sub", ArchivedAt: nil}
		repo.On("GetDefaultTicketType", mock.Anything, tkWS, tkSpace).Return(&defaultType, nil)
		repo.On("GetInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(&initialStatus, nil)
		repo.On("FindTicket", mock.Anything, tkWS, tkParent).Return(&parent, nil)
		repo.On("FindTicketType", mock.Anything, tkWS, tkSpace, "type-sub").Return(&parentType, nil)
		repo.On("ListTicketParentChain", mock.Anything, tkWS, tkParent).Return([]domain.Ticket{}, nil)

		_, err := ticket.NewCreateTicketUseCase(repo).Execute(context.Background(), ticket.CreateTicketInput{
			WorkspaceID: tkWS, SpaceID: tkSpace, Title: "x", Doc: `{"type":"doc","content":[]}`,
			CreatedByUserID: 1, ParentID: &tkParent,
		})
		require.ErrorIs(t, err, domain.ErrTicketHierarchyRejected)
	})

	t.Run("深さ4は拒否", func(t *testing.T) {
		repo := &mockTicketRepo{}
		defaultType := tkDefaultType()
		initialStatus := tkInitialStatus()
		parentType := domain.TicketType{ID: "type-task", HierarchyLevel: 0}
		parent := domain.Ticket{ID: tkParent, WorkspaceID: tkWS, SpaceID: tkSpace, TypeID: "type-task"}
		repo.On("GetDefaultTicketType", mock.Anything, tkWS, tkSpace).Return(&defaultType, nil)
		repo.On("GetInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(&initialStatus, nil)
		repo.On("FindTicket", mock.Anything, tkWS, tkParent).Return(&parent, nil)
		repo.On("FindTicketType", mock.Anything, tkWS, tkSpace, "type-task").Return(&parentType, nil)
		// ListTicketParentChain は「自分を含まない」祖先列（root から順）。親が root(depth1)
		// から数えて既に 2 段の祖先を持つ = 親自身は depth 3。その子は depth 4 になり、
		// 最大 3 段を超える。
		repo.On("ListTicketParentChain", mock.Anything, tkWS, tkParent).Return([]domain.Ticket{
			{ID: "root"}, {ID: "a"},
		}, nil)

		_, err := ticket.NewCreateTicketUseCase(repo).Execute(context.Background(), ticket.CreateTicketInput{
			WorkspaceID: tkWS, SpaceID: tkSpace, Title: "x", Doc: `{"type":"doc","content":[]}`,
			CreatedByUserID: 1, ParentID: &tkParent,
		})
		require.ErrorIs(t, err, domain.ErrTicketHierarchyRejected)
	})

	t.Run("親が実在し規則を満たせば閉包表の自己参照と祖先集合の両方を張る", func(t *testing.T) {
		repo := &mockTicketRepo{}
		defaultType := tkDefaultType() // level 0
		initialStatus := tkInitialStatus()
		parentType := domain.TicketType{ID: "type-task", HierarchyLevel: 0}
		parent := domain.Ticket{ID: tkParent, WorkspaceID: tkWS, SpaceID: tkSpace, TypeID: "type-task"}
		repo.On("GetDefaultTicketType", mock.Anything, tkWS, tkSpace).Return(&defaultType, nil)
		repo.On("GetInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(&initialStatus, nil)
		repo.On("FindTicket", mock.Anything, tkWS, tkParent).Return(&parent, nil)
		repo.On("FindTicketType", mock.Anything, tkWS, tkSpace, "type-task").Return(&parentType, nil)
		repo.On("ListTicketParentChain", mock.Anything, tkWS, tkParent).Return([]domain.Ticket{}, nil)
		repo.On("LastActiveTicketPosition", mock.Anything, tkWS, tkSpace).Return("a0", nil)
		repo.On("LastActiveTicketRankPosition", mock.Anything, tkWS, tkSpace).Return("a0", nil)
		newID := "01a00000-0000-7000-8000-000000000050"
		repo.On("CreateTicket", mock.Anything, mock.AnythingOfType("repository.TicketCreateInput")).
			Return(&domain.Ticket{ID: newID, WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
		repo.On("InsertTicketRank", mock.Anything, tkWS, newID, mock.AnythingOfType("string")).Return(nil)
		repo.On("InsertTicketPathSelf", mock.Anything, tkWS, newID).Return(nil)
		repo.On("InsertTicketPathAncestors", mock.Anything, tkWS, newID, tkParent).Return(nil)
		repo.On("ReplaceTicketPageLinks", mock.Anything, tkWS, newID, []string(nil)).Return(nil)
		repo.On("ReplaceTicketTicketLinks", mock.Anything, tkWS, newID, []string(nil)).Return(nil)

		_, err := ticket.NewCreateTicketUseCase(repo).Execute(context.Background(), ticket.CreateTicketInput{
			WorkspaceID: tkWS, SpaceID: tkSpace, Title: "x", Doc: `{"type":"doc","content":[]}`,
			CreatedByUserID: 1, ParentID: &tkParent,
		})
		require.NoError(t, err)
		repo.AssertCalled(t, "InsertTicketPathSelf", mock.Anything, tkWS, newID)
		repo.AssertCalled(t, "InsertTicketPathAncestors", mock.Anything, tkWS, newID, tkParent)
	})

	t.Run("別スペースの親は404相当", func(t *testing.T) {
		repo := &mockTicketRepo{}
		defaultType := tkDefaultType()
		initialStatus := tkInitialStatus()
		otherSpaceParent := domain.Ticket{ID: tkParent, WorkspaceID: tkWS, SpaceID: "other-space", TypeID: "type-task"}
		repo.On("GetDefaultTicketType", mock.Anything, tkWS, tkSpace).Return(&defaultType, nil)
		repo.On("GetInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(&initialStatus, nil)
		repo.On("FindTicket", mock.Anything, tkWS, tkParent).Return(&otherSpaceParent, nil)

		_, err := ticket.NewCreateTicketUseCase(repo).Execute(context.Background(), ticket.CreateTicketInput{
			WorkspaceID: tkWS, SpaceID: tkSpace, Title: "x", Doc: `{"type":"doc","content":[]}`,
			CreatedByUserID: 1, ParentID: &tkParent,
		})
		require.ErrorIs(t, err, repository.ErrTicketNotFound)
	})
}

// 本文保存時と同じく、作成時も pageRef/ticketRef の title を剥がし、plain_text を作り、
// 派生表（ticket_page_links / ticket_ticket_links）を張る。
func Test_チケット作成_本文から参照を張る(t *testing.T) {
	repo := &mockTicketRepo{}
	defaultType := tkDefaultType()
	initialStatus := tkInitialStatus()
	repo.On("GetDefaultTicketType", mock.Anything, tkWS, tkSpace).Return(&defaultType, nil)
	repo.On("GetInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(&initialStatus, nil)
	repo.On("LastActiveTicketPosition", mock.Anything, tkWS, tkSpace).Return("", nil)
	repo.On("LastActiveTicketRankPosition", mock.Anything, tkWS, tkSpace).Return("", nil)

	pageID := "01a00000-0000-7000-8000-0000000000e1"
	var captured repository.TicketCreateInput
	repo.On("CreateTicket", mock.Anything, mock.AnythingOfType("repository.TicketCreateInput")).
		Run(func(args mock.Arguments) { captured = args.Get(1).(repository.TicketCreateInput) }).
		Return(&domain.Ticket{ID: "ticket-1", WorkspaceID: tkWS, SpaceID: tkSpace}, nil)
	repo.On("InsertTicketRank", mock.Anything, tkWS, "ticket-1", mock.AnythingOfType("string")).Return(nil)
	repo.On("InsertTicketPathSelf", mock.Anything, tkWS, "ticket-1").Return(nil)
	repo.On("ReplaceTicketPageLinks", mock.Anything, tkWS, "ticket-1", []string{pageID}).Return(nil)
	repo.On("ReplaceTicketTicketLinks", mock.Anything, tkWS, "ticket-1", []string(nil)).Return(nil)

	doc := `{"type":"doc","content":[{"type":"paragraph","content":[
		{"type":"text","text":"参照先はこちら: "},
		{"type":"pageRef","attrs":{"pageId":"` + pageID + `","title":"隠したい題名"}}
	]}]}`
	_, err := ticket.NewCreateTicketUseCase(repo).Execute(context.Background(), ticket.CreateTicketInput{
		WorkspaceID: tkWS, SpaceID: tkSpace, Title: "x", Doc: doc, CreatedByUserID: 1,
	})
	require.NoError(t, err)

	assert.NotContains(t, string(captured.Doc), "隠したい題名", "保存前に title を剥がす")
	assert.Contains(t, string(captured.Doc), pageID)
	assert.NotEmpty(t, captured.PlainText)
	assert.NotContains(t, captured.PlainText, "隠したい題名")
	repo.AssertCalled(t, "ReplaceTicketPageLinks", mock.Anything, tkWS, "ticket-1", []string{pageID})
}
