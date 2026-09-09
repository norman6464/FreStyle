package ticket_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func tkBaseTicket() *domain.Ticket {
	return &domain.Ticket{
		ID: tkTicket, WorkspaceID: tkWS, SpaceID: tkSpace,
		TypeID: "type-task", Title: "旧タイトル", Doc: []byte(`{"type":"doc","content":[]}`),
		Priority: domain.TicketPriorityDefault, ParentID: nil,
	}
}

func Test_チケット更新_必須項目の検証(t *testing.T) {
	uc := ticket.NewUpdateTicketUseCase(&mockTicketRepo{})
	_, err := uc.Execute(context.Background(), ticket.UpdateTicketInput{TicketID: tkTicket, ActorUserID: 1})
	require.Error(t, err)
	_, err = uc.Execute(context.Background(), ticket.UpdateTicketInput{WorkspaceID: tkWS, ActorUserID: 1})
	require.Error(t, err)
}

// PUT 相当なので TypeID / Title は毎回必須。空のまま repo まで進むと ErrTicketNotFound に
// 化けてしまう（kbParseID("") が弾く）ため、ここで明確な入力エラーとして断る。
func Test_チケット更新_TypeIDとTitleは毎回必須(t *testing.T) {
	repo := &mockTicketRepo{}

	_, err := ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: "x", Doc: `{"type":"doc","content":[]}`, Priority: domain.TicketPriorityDefault,
	})
	require.Error(t, err, "typeID 必須")
	repo.AssertNotCalled(t, "FindTicket")

	_, err = ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: "  ", Doc: `{"type":"doc","content":[]}`, TypeID: "type-task", Priority: domain.TicketPriorityDefault,
	})
	require.ErrorIs(t, err, domain.ErrInvalidTicketName, "空白だけの title は拒否")
	repo.AssertNotCalled(t, "FindTicket")
}

// 開始日が期限より後なら、FindTicket すら呼ばずに断る。
func Test_チケット更新_開始日が期限より後なら拒否(t *testing.T) {
	repo := &mockTicketRepo{}
	start, due := "2026-09-10", "2026-09-01"

	_, err := ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: "x", Doc: `{"type":"doc","content":[]}`, TypeID: "type-task", Priority: domain.TicketPriorityDefault,
		StartDate: &start, DueDate: &due,
	})
	require.ErrorIs(t, err, domain.ErrTicketDateRangeInverted)
	repo.AssertNotCalled(t, "FindTicket")
}

// タイトルが変わったときだけ履歴に 'title' 項目を書く。
func Test_チケット更新_変わった項目だけ履歴に残す(t *testing.T) {
	repo := &mockTicketRepo{}
	before := tkBaseTicket()
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).Return(before, nil)
	repo.On("UpdateTicket", mock.Anything, tkWS, tkTicket, mock.AnythingOfType("repository.TicketUpdateFields")).
		Return(&domain.Ticket{ID: tkTicket, WorkspaceID: tkWS, Title: "新タイトル"}, nil)
	repo.On("ReplaceTicketPageLinks", mock.Anything, tkWS, tkTicket, []string(nil)).Return(nil)
	repo.On("ReplaceTicketTicketLinks", mock.Anything, tkWS, tkTicket, []string(nil)).Return(nil)
	repo.On("InsertTicketChangeGroup", mock.Anything, mock.MatchedBy(func(g *domain.TicketChangeGroup) bool {
		return len(g.Items) == 1 && g.Items[0].Field == domain.TicketChangeFieldTitle
	})).Return(nil)

	_, err := ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: "新タイトル", Doc: `{"type":"doc","content":[]}`, TypeID: "type-task", Priority: domain.TicketPriorityDefault,
	})
	require.NoError(t, err)
}

// 何も変わっていなければ履歴を残さない。
func Test_チケット更新_変化が無ければ履歴を残さない(t *testing.T) {
	repo := &mockTicketRepo{}
	before := tkBaseTicket()
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).Return(before, nil)
	repo.On("UpdateTicket", mock.Anything, tkWS, tkTicket, mock.AnythingOfType("repository.TicketUpdateFields")).
		Return(before, nil)
	repo.On("ReplaceTicketPageLinks", mock.Anything, tkWS, tkTicket, []string(nil)).Return(nil)
	repo.On("ReplaceTicketTicketLinks", mock.Anything, tkWS, tkTicket, []string(nil)).Return(nil)

	_, err := ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: before.Title, Doc: string(before.Doc), TypeID: before.TypeID, Priority: before.Priority,
	})
	require.NoError(t, err)
	repo.AssertNotCalled(t, "InsertTicketChangeGroup")
}

// 種別を段 -1（小作業）へ変えるとき、現役の子を持っていれば拒否する
// （-1 はどんな子も持てないため。設計 Ⅳ-D）。
func Test_チケット更新_子を持つチケットを小作業へ変えると拒否(t *testing.T) {
	repo := &mockTicketRepo{}
	before := tkBaseTicket() // type-task = level 0（下の FindTicketType で定義）
	subType := domain.TicketType{ID: "type-sub", HierarchyLevel: -1}
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).Return(before, nil)
	repo.On("FindTicketType", mock.Anything, tkWS, tkSpace, "type-sub").Return(&subType, nil)
	repo.On("CountActiveTicketChildren", mock.Anything, tkWS, tkTicket).Return(int64(2), nil)

	_, err := ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: before.Title, Doc: string(before.Doc), TypeID: "type-sub", Priority: before.Priority,
	})
	require.ErrorIs(t, err, domain.ErrTicketHierarchyRejected)
	repo.AssertNotCalled(t, "UpdateTicket")
}

// 種別を変えるとき、既存の親との階層規則も再検証する。
func Test_チケット更新_種別変更で親との階層規則を再検証する(t *testing.T) {
	repo := &mockTicketRepo{}
	before := tkBaseTicket()
	before.ParentID = &tkParent
	bundleType := domain.TicketType{ID: "type-bundle", HierarchyLevel: 1}
	parentType := domain.TicketType{ID: "type-parent", HierarchyLevel: 0}
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).Return(before, nil)
	repo.On("FindTicketType", mock.Anything, tkWS, tkSpace, "type-bundle").Return(&bundleType, nil)
	repo.On("CountActiveTicketChildren", mock.Anything, tkWS, tkTicket).Return(int64(0), nil)
	repo.On("FindTicket", mock.Anything, tkWS, tkParent).
		Return(&domain.Ticket{ID: tkParent, WorkspaceID: tkWS, SpaceID: tkSpace, TypeID: "type-parent"}, nil)
	repo.On("FindTicketType", mock.Anything, tkWS, tkSpace, "type-parent").Return(&parentType, nil)

	_, err := ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: before.Title, Doc: string(before.Doc), TypeID: "type-bundle", Priority: before.Priority,
	})
	require.ErrorIs(t, err, domain.ErrTicketHierarchyRejected, "段1(束ね)の親を持つチケット自身を段1にはできない")
}

// pageRef の title（クライアントが持っていた表示用の写し）が保存済みと違うだけでは
// 本文が「変わった」ことにしない。in.Doc（生の値）と before.Doc（既に剥がされた保存値）を
// 単純比較すると誤検知する — 剥がした後どうしで比較する。
func Test_チケット更新_参照タイトルの違いだけでは本文変更と見なさない(t *testing.T) {
	repo := &mockTicketRepo{}
	pageID := "01a00000-0000-7000-8000-0000000000e1"
	before := tkBaseTicket()
	before.Doc = []byte(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"pageRef","attrs":{"pageId":"` + pageID + `","title":null}}]}]}`)
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).Return(before, nil)
	repo.On("UpdateTicket", mock.Anything, tkWS, tkTicket, mock.AnythingOfType("repository.TicketUpdateFields")).
		Return(before, nil)
	repo.On("ReplaceTicketPageLinks", mock.Anything, tkWS, tkTicket, []string{pageID}).Return(nil)
	repo.On("ReplaceTicketTicketLinks", mock.Anything, tkWS, tkTicket, []string(nil)).Return(nil)

	// クライアントは自分が見えている表示用 title（"今の題名"）を含めて送ってくる。
	doc := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"pageRef","attrs":{"pageId":"` + pageID + `","title":"今の題名"}}]}]}`
	_, err := ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: before.Title, Doc: doc, TypeID: before.TypeID, Priority: before.Priority,
	})
	require.NoError(t, err)
	repo.AssertNotCalled(t, "InsertTicketChangeGroup")
}

// 本文を変えると、保存前に title を剥がし、plain_text を作り直し、派生表を張り替える。
func Test_チケット更新_本文の参照を張り替える(t *testing.T) {
	repo := &mockTicketRepo{}
	before := tkBaseTicket()
	repo.On("FindTicket", mock.Anything, tkWS, tkTicket).Return(before, nil)
	repo.On("UpdateTicket", mock.Anything, tkWS, tkTicket, mock.AnythingOfType("repository.TicketUpdateFields")).
		Return(&domain.Ticket{ID: tkTicket, WorkspaceID: tkWS}, nil)

	pageID := "01a00000-0000-7000-8000-0000000000e1"
	doc := `{"type":"doc","content":[{"type":"paragraph","content":[
		{"type":"text","text":"更新後の本文"},
		{"type":"pageRef","attrs":{"pageId":"` + pageID + `"}}
	]}]}`
	repo.On("ReplaceTicketPageLinks", mock.Anything, tkWS, tkTicket, []string{pageID}).Return(nil)
	repo.On("ReplaceTicketTicketLinks", mock.Anything, tkWS, tkTicket, []string(nil)).Return(nil)
	repo.On("InsertTicketChangeGroup", mock.Anything, mock.AnythingOfType("*domain.TicketChangeGroup")).Return(nil)

	_, err := ticket.NewUpdateTicketUseCase(repo).Execute(context.Background(), ticket.UpdateTicketInput{
		WorkspaceID: tkWS, TicketID: tkTicket, ActorUserID: 1,
		Title: before.Title, Doc: doc, TypeID: before.TypeID, Priority: before.Priority,
	})
	require.NoError(t, err)
	repo.AssertCalled(t, "ReplaceTicketPageLinks", mock.Anything, tkWS, tkTicket, []string{pageID})
}
