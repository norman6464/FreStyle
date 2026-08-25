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

func Test_本文書き換え_docを行に分解して全入れ替えする(t *testing.T) {
	doc := `{"type":"doc","content":[
		{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"見出し"}]},
		{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"項目"}]}]}]}
	]}`
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	var gotRows []repository.BlockWrite
	var gotSnapshot string
	repo.On("ReplacePageBlocks", mock.Anything, kbWS, kbPage, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			gotRows = args.Get(3).([]repository.BlockWrite)
			gotSnapshot = args.String(4)
		}).Return(nil)
	repo.On("GetPageSnapshot", mock.Anything, kbWS, kbPage).
		Return(&domain.PageSnapshot{PageID: kbPage, Doc: doc}, nil)
	uc := usecase.NewReplacePageBlocksUseCase(repo)

	out, err := uc.Execute(context.Background(), usecase.ReplacePageBlocksInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: doc,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, gotRows, 4, "heading / bulletList / listItem / paragraph の 4 行")
	assert.Equal(t, domain.BlockTypeHeading, gotRows[0].Type)
	assert.Equal(t, domain.BlockTypeBulletList, gotRows[1].Type)
	assert.Equal(t, 1, gotRows[2].ParentIndex, "listItem の親は bulletList")
	assert.JSONEq(t, doc, gotSnapshot, "snapshot は行から再生成した正規形の doc")
}

func Test_本文書き換え_不正なdocは保存せず失敗(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	uc := usecase.NewReplacePageBlocksUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.ReplacePageBlocksInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: `{"type":"doc","content":[{"type":"iframe"}]}`,
	})
	require.ErrorIs(t, err, usecase.ErrPageDocUnknownNodeType)
	repo.AssertNotCalled(t, "ReplacePageBlocks", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_本文書き換え_アーカイブ済みページは拒否(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbArchivedPage(kbPage, kbSpace, nil), nil)
	uc := usecase.NewReplacePageBlocksUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.ReplacePageBlocksInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: `{"type":"doc","content":[]}`,
	})
	require.ErrorIs(t, err, usecase.ErrPageArchived)
}

func Test_本文書き換え_無いページはそのまま失敗(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("FindPage", mock.Anything, kbWS, kbPage).Return(nil, repository.ErrPageNotFound)
	uc := usecase.NewReplacePageBlocksUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.ReplacePageBlocksInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: `{"type":"doc","content":[]}`,
	})
	require.ErrorIs(t, err, repository.ErrPageNotFound)
}
