package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_ページ移動_自分自身の下には移せない(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	uc := usecase.NewMovePageUseCase(repo)

	self := kbPage
	_, err := uc.Execute(context.Background(), usecase.MovePageInput{
		WorkspaceID: kbWS, PageID: kbPage, NewParentID: &self,
	})
	require.ErrorIs(t, err, usecase.ErrPageCycle)
	repo.AssertNotCalled(t, "MovePage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_ページ移動_子孫の下には移せない(t *testing.T) {
	desc := "0198a000-0000-7000-8000-0000000000aa"
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	repo.On("FindPage", mock.Anything, kbWS, desc).Return(kbActivePage(desc, kbSpace, nil), nil)
	repo.On("HasDescendant", mock.Anything, kbWS, kbPage, desc).Return(true, nil)
	uc := usecase.NewMovePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.MovePageInput{
		WorkspaceID: kbWS, PageID: kbPage, NewParentID: &desc,
	})
	require.ErrorIs(t, err, usecase.ErrPageCycle)
	repo.AssertNotCalled(t, "MovePage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_ページ移動_アーカイブ済みページは移せない(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbArchivedPage(kbPage, kbSpace, nil), nil)
	uc := usecase.NewMovePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.MovePageInput{WorkspaceID: kbWS, PageID: kbPage})
	require.ErrorIs(t, err, usecase.ErrPageArchived)
}

func Test_ページ移動_アーカイブ済みの親の下には移せない(t *testing.T) {
	parent := "0198a000-0000-7000-8000-0000000000bb"
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	repo.On("FindPage", mock.Anything, kbWS, parent).Return(kbArchivedPage(parent, kbSpace, nil), nil)
	uc := usecase.NewMovePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.MovePageInput{
		WorkspaceID: kbWS, PageID: kbPage, NewParentID: &parent,
	})
	require.ErrorIs(t, err, usecase.ErrPageParentArchived)
}

func Test_ページ移動_指定スペースと親のスペースが食い違えば拒否(t *testing.T) {
	parent := "0198a000-0000-7000-8000-0000000000bb"
	otherSpace := "0198a000-0000-7000-8000-0000000000cc"
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	repo.On("FindPage", mock.Anything, kbWS, parent).Return(kbActivePage(parent, otherSpace, nil), nil)
	uc := usecase.NewMovePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.MovePageInput{
		WorkspaceID: kbWS, PageID: kbPage, NewParentID: &parent, NewSpaceID: kbSpace,
	})
	require.ErrorIs(t, err, usecase.ErrPageParentSpaceMismatch)
}

func Test_ページ移動_親の下の末尾へ移す(t *testing.T) {
	parent := "0198a000-0000-7000-8000-0000000000bb"
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil).Once()
	repo.On("FindPage", mock.Anything, kbWS, parent).Return(kbActivePage(parent, kbSpace, nil), nil)
	repo.On("HasDescendant", mock.Anything, kbWS, kbPage, parent).Return(false, nil)
	repo.On("LastActiveSiblingPosition", mock.Anything, kbWS, kbSpace, &parent).Return("a2", nil)
	repo.On("MovePage", mock.Anything, kbWS, kbPage, &parent, kbSpace, "a3").Return(nil)
	moved := kbActivePage(kbPage, kbSpace, &parent)
	moved.Position = "a3"
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(moved, nil)
	uc := usecase.NewMovePageUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.MovePageInput{
		WorkspaceID: kbWS, PageID: kbPage, NewParentID: &parent,
	})
	require.NoError(t, err)
	assert.Equal(t, "a3", got.Position)
	repo.AssertExpectations(t)
}

func Test_ページ移動_別スペースのルートへ移す(t *testing.T) {
	otherSpace := "0198a000-0000-7000-8000-0000000000cc"
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	repo.On("FindSpace", mock.Anything, kbWS, otherSpace).Return(&domain.Space{ID: otherSpace, WorkspaceID: kbWS}, nil)
	repo.On("LastActiveSiblingPosition", mock.Anything, kbWS, otherSpace, (*string)(nil)).Return("", nil)
	repo.On("MovePage", mock.Anything, kbWS, kbPage, (*string)(nil), otherSpace, "a0").Return(nil)
	uc := usecase.NewMovePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.MovePageInput{
		WorkspaceID: kbWS, PageID: kbPage, NewSpaceID: otherSpace,
	})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func Test_ページ移動_移動先スペースが無ければ失敗(t *testing.T) {
	otherSpace := "0198a000-0000-7000-8000-0000000000cc"
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	repo.On("FindSpace", mock.Anything, kbWS, otherSpace).Return(nil, repository.ErrSpaceNotFound)
	uc := usecase.NewMovePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.MovePageInput{
		WorkspaceID: kbWS, PageID: kbPage, NewSpaceID: otherSpace,
	})
	require.ErrorIs(t, err, repository.ErrSpaceNotFound)
}

func Test_ページアーカイブ_サブツリーごと隠す(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	repo.On("ArchivePageSubtree", mock.Anything, kbWS, kbPage).Return(nil)
	uc := usecase.NewArchivePageUseCase(repo)

	require.NoError(t, uc.Execute(context.Background(), usecase.ArchivePageInput{WorkspaceID: kbWS, PageID: kbPage}))
	repo.AssertExpectations(t)
}

func Test_ページアーカイブ_アーカイブ済みなら何もしない(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbArchivedPage(kbPage, kbSpace, nil), nil)
	uc := usecase.NewArchivePageUseCase(repo)

	require.NoError(t, uc.Execute(context.Background(), usecase.ArchivePageInput{WorkspaceID: kbWS, PageID: kbPage}))
	repo.AssertNotCalled(t, "ArchivePageSubtree", mock.Anything, mock.Anything, mock.Anything)
}

func Test_ページ復帰_一括分をarchived_atを境に戻す(t *testing.T) {
	page := kbArchivedPage(kbPage, kbSpace, nil)
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(page, nil).Once()
	repo.On("HasActiveSiblingPosition", mock.Anything, kbWS, kbSpace, (*string)(nil), page.Position, kbPage).Return(false, nil)
	repo.On("UnarchivePageSubtree", mock.Anything, kbWS, kbPage, *page.ArchivedAt, (*string)(nil)).Return(nil)
	restored := kbActivePage(kbPage, kbSpace, nil)
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(restored, nil)
	uc := usecase.NewUnarchivePageUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.UnarchivePageInput{WorkspaceID: kbWS, PageID: kbPage})
	require.NoError(t, err)
	assert.Nil(t, got.ArchivedAt)
	repo.AssertExpectations(t)
}

func Test_ページ復帰_positionが衝突したら末尾へ再採番(t *testing.T) {
	page := kbArchivedPage(kbPage, kbSpace, nil)
	page.Position = "a0"
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(page, nil).Once()
	repo.On("HasActiveSiblingPosition", mock.Anything, kbWS, kbSpace, (*string)(nil), "a0", kbPage).Return(true, nil)
	repo.On("LastActiveSiblingPosition", mock.Anything, kbWS, kbSpace, (*string)(nil)).Return("a7", nil)
	newPos := "a8"
	repo.On("UnarchivePageSubtree", mock.Anything, kbWS, kbPage, *page.ArchivedAt, &newPos).Return(nil)
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	uc := usecase.NewUnarchivePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.UnarchivePageInput{WorkspaceID: kbWS, PageID: kbPage})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func Test_ページ復帰_親がアーカイブ中なら拒否(t *testing.T) {
	parent := "0198a000-0000-7000-8000-0000000000bb"
	page := kbArchivedPage(kbPage, kbSpace, &parent)
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(page, nil)
	repo.On("FindPage", mock.Anything, kbWS, parent).Return(kbArchivedPage(parent, kbSpace, nil), nil)
	uc := usecase.NewUnarchivePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.UnarchivePageInput{WorkspaceID: kbWS, PageID: kbPage})
	require.ErrorIs(t, err, usecase.ErrPageParentArchived)
	repo.AssertNotCalled(t, "UnarchivePageSubtree", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_ページ復帰_現役ページなら何もしない(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	uc := usecase.NewUnarchivePageUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.UnarchivePageInput{WorkspaceID: kbWS, PageID: kbPage})
	require.NoError(t, err)
	assert.Nil(t, got.ArchivedAt)
	repo.AssertNotCalled(t, "UnarchivePageSubtree", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
