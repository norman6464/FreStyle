package usecase_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	kbWS    = "0198a000-0000-7000-8000-000000000001"
	kbSpace = "0198a000-0000-7000-8000-000000000002"
	kbPage  = "0198a000-0000-7000-8000-000000000003"
)

func kbActivePage(id, spaceID string, parentID *string) *domain.Page {
	return &domain.Page{
		ID: id, WorkspaceID: kbWS, SpaceID: spaceID, ParentID: parentID,
		Position: "a0", Title: "ページ", CreatedByUserID: 1,
	}
}

func kbArchivedPage(id, spaceID string, parentID *string) *domain.Page {
	p := kbActivePage(id, spaceID, parentID)
	at := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	p.ArchivedAt = &at
	return p
}

func Test_ページ作成_必須項目の検証(t *testing.T) {
	uc := usecase.NewCreatePageUseCase(&mockKnowledgeBaseRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.CreatePageInput{SpaceID: kbSpace, CreatedByUserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.CreatePageInput{WorkspaceID: kbWS, CreatedByUserID: 1})
	require.Error(t, err, "spaceID 必須")
	_, err = uc.Execute(ctx, usecase.CreatePageInput{WorkspaceID: kbWS, SpaceID: kbSpace})
	require.Error(t, err, "createdByUserID 必須")
	_, err = uc.Execute(ctx, usecase.CreatePageInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, CreatedByUserID: 1,
		Title: strings.Repeat("あ", 201),
	})
	require.Error(t, err, "title は 200 文字まで")
}

func Test_ページ作成_スペースが無ければ失敗(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(nil, repository.ErrSpaceNotFound)
	uc := usecase.NewCreatePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreatePageInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, CreatedByUserID: 1,
	})
	require.ErrorIs(t, err, repository.ErrSpaceNotFound)
}

func Test_ページ作成_親が別スペースなら拒否(t *testing.T) {
	otherSpace := "0198a000-0000-7000-8000-00000000000f"
	parentID := kbPage
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	repo.On("FindPage", mock.Anything, kbWS, parentID).Return(kbActivePage(parentID, otherSpace, nil), nil)
	uc := usecase.NewCreatePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreatePageInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, ParentID: &parentID, CreatedByUserID: 1,
	})
	require.ErrorIs(t, err, usecase.ErrPageParentSpaceMismatch)
}

func Test_ページ作成_アーカイブ済みの親は拒否(t *testing.T) {
	parentID := kbPage
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	repo.On("FindPage", mock.Anything, kbWS, parentID).Return(kbArchivedPage(parentID, kbSpace, nil), nil)
	uc := usecase.NewCreatePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreatePageInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, ParentID: &parentID, CreatedByUserID: 1,
	})
	require.ErrorIs(t, err, usecase.ErrPageParentArchived)
}

func Test_ページ作成_末尾のpositionを採番して保存(t *testing.T) {
	parentID := kbPage
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	repo.On("FindPage", mock.Anything, kbWS, parentID).Return(kbActivePage(parentID, kbSpace, nil), nil)
	repo.On("LastActiveSiblingPosition", mock.Anything, kbWS, kbSpace, &parentID).Return("a5", nil)
	var created *domain.Page
	repo.On("CreatePage", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { created = args.Get(1).(*domain.Page) }).Return(nil)
	uc := usecase.NewCreatePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreatePageInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, ParentID: &parentID, Title: "新ページ", CreatedByUserID: 7,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "a6", created.Position, "末尾 a5 の次 = fracindex.Between(\"a5\", \"\")")
	assert.Equal(t, kbSpace, created.SpaceID)
	assert.Equal(t, &parentID, created.ParentID)
	assert.Equal(t, uint64(7), created.CreatedByUserID)
}

func Test_ページ作成_最初の1件はルートに採番(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	repo.On("LastActiveSiblingPosition", mock.Anything, kbWS, kbSpace, (*string)(nil)).Return("", nil)
	var created *domain.Page
	repo.On("CreatePage", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { created = args.Get(1).(*domain.Page) }).Return(nil)
	uc := usecase.NewCreatePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreatePageInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, CreatedByUserID: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "a0", created.Position, "兄弟なし = fracindex の最初のキー")
	assert.Nil(t, created.ParentID)
}

func Test_ページ取得_snapshotがあればそれを返す(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	repo.On("GetPageSnapshot", mock.Anything, kbWS, kbPage).
		Return(&domain.PageSnapshot{PageID: kbPage, Doc: `{"type":"doc","content":[]}`}, nil)
	uc := usecase.NewGetPageUseCase(repo)

	out, err := uc.Execute(context.Background(), usecase.GetPageInput{WorkspaceID: kbWS, PageID: kbPage})
	require.NoError(t, err)
	assert.Equal(t, `{"type":"doc","content":[]}`, out.Doc)
	assert.Equal(t, kbPage, out.Page.ID)
	repo.AssertNotCalled(t, "ListBlocksByPage", mock.Anything, mock.Anything, mock.Anything)
}

func Test_ページ取得_snapshotが無ければブロックから組み立てる(t *testing.T) {
	inline := `[{"type":"text","text":"本文"}]`
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	repo.On("GetPageSnapshot", mock.Anything, kbWS, kbPage).Return(nil, repository.ErrPageSnapshotNotFound)
	repo.On("ListBlocksByPage", mock.Anything, kbWS, kbPage).Return([]domain.Block{
		{ID: "b1", PageID: kbPage, Type: domain.BlockTypeParagraph, Position: "a0", Attrs: "{}", Inline: &inline},
	}, nil)
	uc := usecase.NewGetPageUseCase(repo)

	out, err := uc.Execute(context.Background(), usecase.GetPageInput{WorkspaceID: kbWS, PageID: kbPage})
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"本文"}]}]}`, out.Doc)
}

func Test_ページ取得_無いページはそのまま失敗(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(nil, repository.ErrPageNotFound)
	uc := usecase.NewGetPageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.GetPageInput{WorkspaceID: kbWS, PageID: kbPage})
	require.ErrorIs(t, err, repository.ErrPageNotFound)
}

func Test_ページツリー_親子と兄弟順を組み立てる(t *testing.T) {
	root1 := kbActivePage("p1", kbSpace, nil)
	root1.Position = "a0"
	root2 := kbActivePage("p2", kbSpace, nil)
	root2.Position = "a1"
	p1 := "p1"
	child1 := kbActivePage("p3", kbSpace, &p1)
	child1.Position = "a0"
	child2 := kbActivePage("p4", kbSpace, &p1)
	child2.Position = "a1"

	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	// ListActivePagesBySpace は position 順で返る（クエリの ORDER BY）。
	repo.On("ListActivePagesBySpace", mock.Anything, kbWS, kbSpace).
		Return([]domain.Page{*root1, *child1, *root2, *child2}, nil)
	uc := usecase.NewGetPageTreeUseCase(repo)

	tree, err := uc.Execute(context.Background(), usecase.GetPageTreeInput{WorkspaceID: kbWS, SpaceID: kbSpace})
	require.NoError(t, err)
	require.Len(t, tree, 2)
	assert.Equal(t, "p1", tree[0].Page.ID)
	assert.Equal(t, "p2", tree[1].Page.ID)
	require.Len(t, tree[0].Children, 2)
	assert.Equal(t, "p3", tree[0].Children[0].Page.ID)
	assert.Equal(t, "p4", tree[0].Children[1].Page.ID)
	assert.Empty(t, tree[1].Children)
}

func Test_ページツリー_スペースが無ければ失敗(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(nil, repository.ErrSpaceNotFound)
	uc := usecase.NewGetPageTreeUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.GetPageTreeInput{WorkspaceID: kbWS, SpaceID: kbSpace})
	require.ErrorIs(t, err, repository.ErrSpaceNotFound)
}

func Test_ページツリー_親が一覧に無い行はルート扱いで隠さない(t *testing.T) {
	missing := "not-in-list"
	orphan := kbActivePage("p9", kbSpace, &missing)
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	repo.On("ListActivePagesBySpace", mock.Anything, kbWS, kbSpace).Return([]domain.Page{*orphan}, nil)
	uc := usecase.NewGetPageTreeUseCase(repo)

	tree, err := uc.Execute(context.Background(), usecase.GetPageTreeInput{WorkspaceID: kbWS, SpaceID: kbSpace})
	require.NoError(t, err)
	require.Len(t, tree, 1)
	assert.Equal(t, "p9", tree[0].Page.ID)
}

func Test_ページ改名_アーカイブ済みは拒否(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbArchivedPage(kbPage, kbSpace, nil), nil)
	uc := usecase.NewRenamePageUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.RenamePageInput{WorkspaceID: kbWS, PageID: kbPage, Title: "x"})
	require.ErrorIs(t, err, usecase.ErrPageArchived)
}

func Test_ページ改名_タイトルを更新する(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	updated := kbActivePage(kbPage, kbSpace, nil)
	updated.Title = "改名後"
	repo.On("UpdatePageTitle", mock.Anything, kbWS, kbPage, "改名後").Return(updated, nil)
	uc := usecase.NewRenamePageUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.RenamePageInput{WorkspaceID: kbWS, PageID: kbPage, Title: "改名後"})
	require.NoError(t, err)
	assert.Equal(t, "改名後", got.Title)
}

func Test_ページ改名_長すぎるタイトルは拒否(t *testing.T) {
	uc := usecase.NewRenamePageUseCase(&mockKnowledgeBaseRepo{})
	_, err := uc.Execute(context.Background(), usecase.RenamePageInput{
		WorkspaceID: kbWS, PageID: kbPage, Title: strings.Repeat("あ", 201),
	})
	require.Error(t, err)
}
