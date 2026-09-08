package kb_test

import (
	"context"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/kb"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// FRESTYLE-435（段5・テンプレート）の kb.CreateTemplateFromPageUseCase /
// kb.ListPageTemplatesUseCase / kb.DeletePageTemplateUseCase /
// kb.CreatePageFromTemplateUseCase の単体テスト。page_usecase_external_test.go /
// page_version_usecase_external_test.go と同じ流儀（testify/mock、kbWS/kbSpace/kbPage/
// kbEditorUserID/kbActivePage/fakeTxManager/inTx を共有）。

const kbTemplateDoc = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"雛形の本文"}]}]}`

// Test_雛形として保存_snapshotがあればそれをdocとして使う は GetPageUseCase.Execute と同じ
// フォールバック手順（snapshot 優先）を CreateTemplateFromPageUseCase も踏むことを固定する。
func Test_雛形として保存_snapshotがあればそれをdocとして使う(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	kbRepo.On("GetPageSnapshot", mock.Anything, kbWS, kbPage).
		Return(&domain.PageSnapshot{PageID: kbPage, Doc: kbTemplateDoc}, nil)
	templates := &mockPageTemplateRepo{}
	var created *domain.PageTemplate
	templates.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { created = args.Get(1).(*domain.PageTemplate) }).Return(nil)
	uc := kb.NewCreateTemplateFromPageUseCase(kbRepo, templates)

	_, err := uc.Execute(context.Background(), kb.CreateTemplateFromPageInput{
		WorkspaceID: kbWS, PageID: kbPage, Name: "議事録", AuthorUserID: kbEditorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Contains(t, created.Doc, "雛形の本文")
	kbRepo.AssertNotCalled(t, "ListBlocksByPage", mock.Anything, mock.Anything, mock.Anything)
}

// Test_雛形として保存_snapshotが無ければブロックから組み立てる は currentPageDoc の
// フォールバック（ListBlocksByPage → treeFromBlocks/renderPageDoc）を固定する。
func Test_雛形として保存_snapshotが無ければブロックから組み立てる(t *testing.T) {
	inline := `[{"type":"text","text":"組み立てた本文"}]`
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	kbRepo.On("GetPageSnapshot", mock.Anything, kbWS, kbPage).Return(nil, repository.ErrPageSnapshotNotFound)
	kbRepo.On("ListBlocksByPage", mock.Anything, kbWS, kbPage).Return([]domain.Block{
		{ID: "b1", PageID: kbPage, Type: domain.BlockTypeParagraph, Position: "a0", Attrs: "{}", Inline: &inline},
	}, nil)
	templates := &mockPageTemplateRepo{}
	var created *domain.PageTemplate
	templates.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { created = args.Get(1).(*domain.PageTemplate) }).Return(nil)
	uc := kb.NewCreateTemplateFromPageUseCase(kbRepo, templates)

	_, err := uc.Execute(context.Background(), kb.CreateTemplateFromPageInput{
		WorkspaceID: kbWS, PageID: kbPage, Name: "議事録", AuthorUserID: kbEditorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Contains(t, created.Doc, "組み立てた本文")
}

// Test_雛形として保存_不正な名前を拒否しrepoを呼ばない は、ValidateTemplateName の検証を
// kbRepo / templates のどちらも呼ぶ前に行うことを固定する。
func Test_雛形として保存_不正な名前を拒否しrepoを呼ばない(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	templates := &mockPageTemplateRepo{}
	uc := kb.NewCreateTemplateFromPageUseCase(kbRepo, templates)

	_, err := uc.Execute(context.Background(), kb.CreateTemplateFromPageInput{
		WorkspaceID: kbWS, PageID: kbPage, Name: "   ", AuthorUserID: kbEditorUserID,
	})
	require.ErrorIs(t, err, domain.ErrInvalidTemplateName)
	kbRepo.AssertNotCalled(t, "FindPage", mock.Anything, mock.Anything, mock.Anything)
	templates.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// Test_雛形として保存_存在しないspaceIdを拒否する は、指定した spaceId が同じワークスペース内に
// 実在しないとき repository.ErrSpaceNotFound を返し、templates.Create を呼ばないことを固定する。
func Test_雛形として保存_存在しないspaceIdを拒否する(t *testing.T) {
	missingSpace := "0198a000-0000-7000-8000-0000000000ff"
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	kbRepo.On("GetPageSnapshot", mock.Anything, kbWS, kbPage).
		Return(&domain.PageSnapshot{PageID: kbPage, Doc: kbTemplateDoc}, nil)
	kbRepo.On("FindSpace", mock.Anything, kbWS, missingSpace).Return(nil, repository.ErrSpaceNotFound)
	templates := &mockPageTemplateRepo{}
	uc := kb.NewCreateTemplateFromPageUseCase(kbRepo, templates)

	_, err := uc.Execute(context.Background(), kb.CreateTemplateFromPageInput{
		WorkspaceID: kbWS, PageID: kbPage, SpaceID: &missingSpace, Name: "議事録", AuthorUserID: kbEditorUserID,
	})
	require.ErrorIs(t, err, repository.ErrSpaceNotFound)
	templates.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func Test_雛形一覧_repoをそのまま呼ぶ(t *testing.T) {
	templates := &mockPageTemplateRepo{}
	want := []domain.PageTemplate{{ID: "t1", WorkspaceID: kbWS, Name: "A"}, {ID: "t2", WorkspaceID: kbWS, Name: "B"}}
	templates.On("List", mock.Anything, kbWS, (*string)(nil)).Return(want, nil)
	uc := kb.NewListPageTemplatesUseCase(templates)

	got, err := uc.Execute(context.Background(), kb.ListPageTemplatesInput{WorkspaceID: kbWS})
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_雛形削除_repoをそのまま呼ぶ(t *testing.T) {
	templates := &mockPageTemplateRepo{}
	templates.On("Delete", mock.Anything, kbWS, "t1").Return(nil)
	uc := kb.NewDeletePageTemplateUseCase(templates)

	err := uc.Execute(context.Background(), kb.DeletePageTemplateInput{WorkspaceID: kbWS, TemplateID: "t1"})
	require.NoError(t, err)
	templates.AssertExpectations(t)
}

// Test_雛形から作成_CreatePageとReplaceBlocksをこの順で正しい引数で呼ぶ は
// CreatePageFromTemplateUseCase が CreatePageUseCase → ReplacePageBlocksUseCase の順で
// 薄いオーケストレーションを行うこと、雛形の本文のブロックidが剥がされてから
// 本文書き込みに渡ることを固定する。
func Test_雛形から作成_CreatePageとReplaceBlocksをこの順で正しい引数で呼ぶ(t *testing.T) {
	const tplDoc = `{"type":"doc","content":[{"type":"paragraph","attrs":{"id":"11111111-1111-1111-1111-111111111111"},"content":[{"type":"text","text":"雛形本文"}]}]}`
	templates := &mockPageTemplateRepo{}
	templates.On("Get", mock.Anything, kbWS, "tpl-1").Return(&domain.PageTemplate{
		ID: "tpl-1", WorkspaceID: kbWS, Name: "議事録", Doc: tplDoc,
	}, nil)

	var order []string
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	kbRepo.On("LastActiveSiblingPosition", mock.Anything, kbWS, kbSpace, (*string)(nil)).Return("", nil)
	kbRepo.On("CreatePage", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			order = append(order, "CreatePage")
			args.Get(1).(*domain.Page).ID = kbPage
		}).Return(nil)
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	kbRepo.On("TouchPageLastEditedBy", mock.Anything, kbWS, kbPage, kbEditorUserID).
		Run(func(mock.Arguments) { order = append(order, "TouchPageLastEditedBy") }).Return(nil)
	var replacedDoc string
	kbRepo.On("ReplacePageBlocks", mock.Anything, kbWS, kbPage, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			order = append(order, "ReplacePageBlocks")
			replacedDoc = args.String(4)
		}).Return(nil)
	kbRepo.On("GetPageSnapshot", mock.Anything, kbWS, kbPage).
		Return(&domain.PageSnapshot{PageID: kbPage, Doc: tplDoc}, nil)

	versionRepo := &mockPageVersionRepo{}
	versionRepo.On("CreateVersionIfDue", mock.Anything, kbWS, kbPage, mock.Anything, kbEditorUserID, (*string)(nil), false).
		Return(false, nil, nil)

	createPageUC := kb.NewCreatePageUseCase(kbRepo)
	replaceUC := kb.NewReplacePageBlocksUseCase(kbRepo, &fakeTxManager{}, versionRepo)
	deleteUC := kb.NewDeletePageUseCase(kbRepo)
	uc := kb.NewCreatePageFromTemplateUseCase(templates, createPageUC, replaceUC, deleteUC)

	out, err := uc.Execute(context.Background(), kb.CreatePageFromTemplateInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, TemplateID: "tpl-1", Title: "新しい議事録", AuthorUserID: kbEditorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, kbPage, out.Page.ID)
	assert.Equal(t, []string{"CreatePage", "TouchPageLastEditedBy", "ReplacePageBlocks"}, order)
	assert.NotContains(t, replacedDoc, "11111111-1111-1111-1111-111111111111",
		"雛形のブロックidは regenerateBlockIDs で剥がされてから本文書き込みに渡る")
	assert.Contains(t, replacedDoc, "雛形本文")
}

// Test_雛形から作成_本文書き込み失敗時に空ページの削除を試みる は、ReplacePageBlocksUseCase が
// 失敗したとき、作成済みの空ページを DeletePageUseCase で後始末しようとすることを固定する。
func Test_雛形から作成_本文書き込み失敗時に空ページの削除を試みる(t *testing.T) {
	const tplDoc = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"雛形本文"}]}]}`
	templates := &mockPageTemplateRepo{}
	templates.On("Get", mock.Anything, kbWS, "tpl-1").Return(&domain.PageTemplate{
		ID: "tpl-1", WorkspaceID: kbWS, Name: "議事録", Doc: tplDoc,
	}, nil)

	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	kbRepo.On("LastActiveSiblingPosition", mock.Anything, kbWS, kbSpace, (*string)(nil)).Return("", nil)
	kbRepo.On("CreatePage", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { args.Get(1).(*domain.Page).ID = kbPage }).Return(nil)
	// ReplacePageBlocksUseCase.Execute は最初に FindPage を呼ぶ。ここをアーカイブ済みにして
	// ErrPageArchived で失敗させ、本文書き込みそのものが失敗するケースを再現する。
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbArchivedPage(kbPage, kbSpace, nil), nil)
	kbRepo.On("DeletePageSubtree", mock.Anything, kbWS, kbPage).Return(nil)

	createPageUC := kb.NewCreatePageUseCase(kbRepo)
	replaceUC := kb.NewReplacePageBlocksUseCase(kbRepo, &fakeTxManager{}, &mockPageVersionRepo{})
	deleteUC := kb.NewDeletePageUseCase(kbRepo)
	uc := kb.NewCreatePageFromTemplateUseCase(templates, createPageUC, replaceUC, deleteUC)

	_, err := uc.Execute(context.Background(), kb.CreatePageFromTemplateInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, TemplateID: "tpl-1", Title: "新しい議事録", AuthorUserID: kbEditorUserID,
	})
	require.ErrorIs(t, err, kb.ErrPageArchived, "後始末に成功した場合は元のエラーをそのまま返す")
	kbRepo.AssertCalled(t, "DeletePageSubtree", mock.Anything, kbWS, kbPage)
}

// Test_雛形から作成_後始末の削除にも失敗したらエラーに残す は、空ページの削除自体が失敗した
// ケースで、返るエラーのメッセージから中途半端な空ページが残っていることが分かることを固定する。
func Test_雛形から作成_後始末の削除にも失敗したらエラーに残す(t *testing.T) {
	const tplDoc = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"雛形本文"}]}]}`
	templates := &mockPageTemplateRepo{}
	templates.On("Get", mock.Anything, kbWS, "tpl-1").Return(&domain.PageTemplate{
		ID: "tpl-1", WorkspaceID: kbWS, Name: "議事録", Doc: tplDoc,
	}, nil)

	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindSpace", mock.Anything, kbWS, kbSpace).Return(&domain.Space{ID: kbSpace, WorkspaceID: kbWS}, nil)
	kbRepo.On("LastActiveSiblingPosition", mock.Anything, kbWS, kbSpace, (*string)(nil)).Return("", nil)
	kbRepo.On("CreatePage", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { args.Get(1).(*domain.Page).ID = kbPage }).Return(nil)
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbArchivedPage(kbPage, kbSpace, nil), nil)
	kbRepo.On("DeletePageSubtree", mock.Anything, kbWS, kbPage).Return(repository.ErrPageNotFound)

	createPageUC := kb.NewCreatePageUseCase(kbRepo)
	replaceUC := kb.NewReplacePageBlocksUseCase(kbRepo, &fakeTxManager{}, &mockPageVersionRepo{})
	deleteUC := kb.NewDeletePageUseCase(kbRepo)
	uc := kb.NewCreatePageFromTemplateUseCase(templates, createPageUC, replaceUC, deleteUC)

	_, err := uc.Execute(context.Background(), kb.CreatePageFromTemplateInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, TemplateID: "tpl-1", Title: "新しい議事録", AuthorUserID: kbEditorUserID,
	})
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), kbPage), "残った空ページのidがエラーメッセージから分かる")
	require.ErrorIs(t, err, kb.ErrPageArchived, "後始末に失敗しても元のエラーはerrors.Isで判定できる")
}
