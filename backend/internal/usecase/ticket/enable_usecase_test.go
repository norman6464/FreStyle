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

// fakeTxManager は DoInTx をそのまま fn(ctx) の呼び出しに委譲する（テストではトランザクションの
// 有無を区別しない。usecase/kb/mocks_test.go の同名ヘルパーと同じ理由）。
type fakeTxManager struct{}

func (fakeTxManager) DoInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// Test_チケット有効化_既に有効なら409相当のエラー は「有効化済み」の正本判定
// （初期状態を持つ現役の状態が 1 つでもあるか）を検証する。
func Test_チケット有効化_既に有効なら拒否(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("HasActiveInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(true, nil)

	_, err := ticket.NewEnableTicketsForSpaceUseCase(repo, fakeTxManager{}).
		Execute(context.Background(), ticket.EnableTicketsForSpaceInput{WorkspaceID: tkWS, SpaceID: tkSpace})

	require.ErrorIs(t, err, repository.ErrTicketsAlreadyEnabled)
	repo.AssertNotCalled(t, "InsertTicketStatus")
	repo.AssertNotCalled(t, "InsertTicketType")
}

// 最小構成: To Do(todo・初期状態) / 進行中(in_progress) / 完了(done) の 3 状態と、
// 種別「タスク」(hierarchy_level=0・既定) 1 つだけを作る（2026-09-09 ユーザー判断:
// 状態・種別は有効化後にいつでも編集画面で変えられる前提なので、最小構成のみでよい）。
func Test_チケット有効化_最小構成を作る(t *testing.T) {
	repo := &mockTicketRepo{}
	repo.On("HasActiveInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(false, nil)

	var insertedStatuses []*domain.TicketStatus
	repo.On("InsertTicketStatus", mock.Anything, mock.AnythingOfType("*domain.TicketStatus")).
		Run(func(args mock.Arguments) {
			s := args.Get(1).(*domain.TicketStatus)
			cp := *s
			insertedStatuses = append(insertedStatuses, &cp)
		}).Return(nil)

	var insertedTypes []*domain.TicketType
	repo.On("InsertTicketType", mock.Anything, mock.AnythingOfType("*domain.TicketType")).
		Run(func(args mock.Arguments) {
			ty := args.Get(1).(*domain.TicketType)
			cp := *ty
			insertedTypes = append(insertedTypes, &cp)
		}).Return(nil)

	err := func() error {
		_, err := ticket.NewEnableTicketsForSpaceUseCase(repo, fakeTxManager{}).
			Execute(context.Background(), ticket.EnableTicketsForSpaceInput{WorkspaceID: tkWS, SpaceID: tkSpace})
		return err
	}()
	require.NoError(t, err)

	require.Len(t, insertedStatuses, 3, "To Do / 進行中 / 完了")
	byCategory := map[domain.TicketStatusCategory]*domain.TicketStatus{}
	for _, s := range insertedStatuses {
		require.Equal(t, tkWS, s.WorkspaceID)
		require.Equal(t, tkSpace, s.SpaceID)
		require.True(t, domain.ValidHexColor(s.Color), "色は正規化済みで保存する: %s", s.Color)
		require.NotEmpty(t, s.Position)
		byCategory[s.Category] = s
	}
	require.Contains(t, byCategory, domain.TicketStatusCategoryTodo)
	require.Contains(t, byCategory, domain.TicketStatusCategoryInProgress)
	require.Contains(t, byCategory, domain.TicketStatusCategoryDone)
	require.True(t, byCategory[domain.TicketStatusCategoryTodo].IsInitial, "初期状態は To Do")
	require.False(t, byCategory[domain.TicketStatusCategoryInProgress].IsInitial)
	require.False(t, byCategory[domain.TicketStatusCategoryDone].IsInitial)
	// position は fracindex のバイト順で To Do → 進行中 → 完了 の順に並ぶこと。
	require.Less(t,
		byCategory[domain.TicketStatusCategoryTodo].Position,
		byCategory[domain.TicketStatusCategoryInProgress].Position)
	require.Less(t,
		byCategory[domain.TicketStatusCategoryInProgress].Position,
		byCategory[domain.TicketStatusCategoryDone].Position)

	require.Len(t, insertedTypes, 1, "種別はタスク1つだけ")
	require.Equal(t, tkWS, insertedTypes[0].WorkspaceID)
	require.Equal(t, tkSpace, insertedTypes[0].SpaceID)
	require.Equal(t, 0, insertedTypes[0].HierarchyLevel)
	require.True(t, insertedTypes[0].IsDefault)
	require.True(t, domain.ValidHexColor(insertedTypes[0].Color))
	require.NotEmpty(t, insertedTypes[0].Position)
}

// sourceSpaceId を指定すると、そのスペースの現役の状態・種別をそのまま複製する
// （設計 Ⅵ「別スペースの構成を複製できる」）。複製元にも参照権限があることは
// handler/呼び出し側が別途確かめる前提（このユースケースは複製そのものだけを担う）。
func Test_チケット有効化_複製元を指定すると現役の構成を複製する(t *testing.T) {
	sourceSpace := "01a00000-0000-7000-8000-000000000099"
	repo := &mockTicketRepo{}
	repo.On("HasActiveInitialTicketStatus", mock.Anything, tkWS, tkSpace).Return(false, nil)
	repo.On("ListTicketStatuses", mock.Anything, tkWS, sourceSpace, false).Return([]domain.TicketStatus{
		{Name: "未対応", Category: domain.TicketStatusCategoryTodo, Color: "#5b6b7a", Position: "a0", IsInitial: true},
		{Name: "対応中", Category: domain.TicketStatusCategoryInProgress, Color: "#a0661a", Position: "a1"},
		{Name: "対応済み", Category: domain.TicketStatusCategoryDone, Color: "#2f6b47", Position: "a2"},
		{Name: "却下", Category: domain.TicketStatusCategoryDone, Color: "#9a3b2e", Position: "a3"},
	}, nil)
	repo.On("ListTicketTypes", mock.Anything, tkWS, sourceSpace, false).Return([]domain.TicketType{
		{Name: "バグ", HierarchyLevel: 0, Color: "#9a3b2e", Position: "a0", IsDefault: true},
		{Name: "要望", HierarchyLevel: 0, Color: "#2f4858", Position: "a1"},
	}, nil)

	var insertedStatuses []*domain.TicketStatus
	repo.On("InsertTicketStatus", mock.Anything, mock.AnythingOfType("*domain.TicketStatus")).
		Run(func(args mock.Arguments) {
			s := args.Get(1).(*domain.TicketStatus)
			cp := *s
			insertedStatuses = append(insertedStatuses, &cp)
		}).Return(nil)
	var insertedTypes []*domain.TicketType
	repo.On("InsertTicketType", mock.Anything, mock.AnythingOfType("*domain.TicketType")).
		Run(func(args mock.Arguments) {
			ty := args.Get(1).(*domain.TicketType)
			cp := *ty
			insertedTypes = append(insertedTypes, &cp)
		}).Return(nil)

	_, err := ticket.NewEnableTicketsForSpaceUseCase(repo, fakeTxManager{}).
		Execute(context.Background(), ticket.EnableTicketsForSpaceInput{
			WorkspaceID: tkWS, SpaceID: tkSpace, SourceSpaceID: &sourceSpace,
		})
	require.NoError(t, err)

	require.Len(t, insertedStatuses, 4)
	require.Len(t, insertedTypes, 2)
	names := map[string]bool{}
	for _, s := range insertedStatuses {
		names[s.Name] = true
		require.Equal(t, tkSpace, s.SpaceID, "複製先のスペースに作る（複製元ではない）")
	}
	require.True(t, names["却下"], "複製元の状態名をそのまま使う")
}
