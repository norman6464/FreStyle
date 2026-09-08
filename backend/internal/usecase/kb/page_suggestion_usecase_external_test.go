package kb_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/kb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// kb.CreateSuggestionUseCase / kb.ListOpenPageSuggestionsUseCase /
// kb.AcceptPageSuggestionUseCase / kb.RejectPageSuggestionUseCase の単体テスト。
// page_template_usecase_external_test.go / page_version_usecase_external_test.go と同じ流儀
// （testify/mock、kbWS/kbSpace/kbPage/kbEditorUserID/kbActivePage/fakeTxManager/inTx を共有）。

const kbSuggestionDoc = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"提案の本文"}]}]}`

// Test_提案作成_AuthorUserIDが0なら拒否 は、著者未指定の入力を repo に触れる前に拒否することを
// 固定する（ReplacePageBlocksUseCase.Execute の EditorUserID 検証と同じ既存方針）。
func Test_提案作成_AuthorUserIDが0なら拒否(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	versionRepo := &mockPageVersionRepo{}
	suggestions := &mockPageSuggestionRepo{}
	uc := kb.NewCreateSuggestionUseCase(kbRepo, versionRepo, suggestions)

	_, err := uc.Execute(context.Background(), kb.CreateSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: kbSuggestionDoc, AuthorUserID: 0,
	})
	require.ErrorIs(t, err, kb.ErrPageEditorRequired)
	kbRepo.AssertNotCalled(t, "FindPage", mock.Anything, mock.Anything, mock.Anything)
	suggestions.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// Test_提案作成_アーカイブ済みページは拒否 は ReplacePageBlocksUseCase と同じ
// ErrPageArchived を再利用し、そこで止まって repo.Create を呼ばないことを固定する。
func Test_提案作成_アーカイブ済みページは拒否(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbArchivedPage(kbPage, kbSpace, nil), nil)
	versionRepo := &mockPageVersionRepo{}
	suggestions := &mockPageSuggestionRepo{}
	uc := kb.NewCreateSuggestionUseCase(kbRepo, versionRepo, suggestions)

	_, err := uc.Execute(context.Background(), kb.CreateSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: kbSuggestionDoc, AuthorUserID: kbEditorUserID,
	})
	require.ErrorIs(t, err, kb.ErrPageArchived)
	suggestions.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// Test_提案作成_不正なdocは保存せず拒否する は、壊れた doc が
// parsePageDoc/flattenPageDoc/renderPageDoc の検証パイプラインで弾かれ、repo.Create まで
// 到達しないことを固定する（あとで採用したときに初めて壊れて発覚するのを防ぐ、という設計）。
func Test_提案作成_不正なdocは保存せず拒否する(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	versionRepo := &mockPageVersionRepo{}
	suggestions := &mockPageSuggestionRepo{}
	uc := kb.NewCreateSuggestionUseCase(kbRepo, versionRepo, suggestions)

	_, err := uc.Execute(context.Background(), kb.CreateSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: `{invalid`, AuthorUserID: kbEditorUserID,
	})
	require.Error(t, err)
	suggestions.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	versionRepo.AssertNotCalled(t, "GetLatestVersion", mock.Anything, mock.Anything, mock.Anything)
}

// Test_提案作成_版が無ければBaseSeqはnil は、そのページにまだ 1 つも版が無いとき
// （versionRepo.GetLatestVersion が (nil, nil) を返すとき）、保存する提案の BaseSeq が nil の
// ままであることを固定する。
func Test_提案作成_版が無ければBaseSeqはnil(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	versionRepo := &mockPageVersionRepo{}
	versionRepo.On("GetLatestVersion", mock.Anything, kbWS, kbPage).Return(nil, nil)
	suggestions := &mockPageSuggestionRepo{}
	var created *domain.PageSuggestion
	suggestions.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { created = args.Get(1).(*domain.PageSuggestion) }).Return(nil)
	uc := kb.NewCreateSuggestionUseCase(kbRepo, versionRepo, suggestions)

	_, err := uc.Execute(context.Background(), kb.CreateSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: kbSuggestionDoc, AuthorUserID: kbEditorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Nil(t, created.BaseSeq)
}

// Test_提案作成_最新版のseqをBaseSeqにする は、既に版があるページへの提案は、その最新の
// seq を BaseSeq として保存することを固定する。
func Test_提案作成_最新版のseqをBaseSeqにする(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	versionRepo := &mockPageVersionRepo{}
	versionRepo.On("GetLatestVersion", mock.Anything, kbWS, kbPage).
		Return(&domain.PageVersion{PageID: kbPage, Seq: 5}, nil)
	suggestions := &mockPageSuggestionRepo{}
	var created *domain.PageSuggestion
	suggestions.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { created = args.Get(1).(*domain.PageSuggestion) }).Return(nil)
	uc := kb.NewCreateSuggestionUseCase(kbRepo, versionRepo, suggestions)

	_, err := uc.Execute(context.Background(), kb.CreateSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: kbSuggestionDoc, AuthorUserID: kbEditorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.BaseSeq)
	assert.Equal(t, int64(5), *created.BaseSeq)
}

// Test_提案作成_正規化後のdocを保存する は、提案者が送った生の文字列ではなく
// renderPageDoc が組み立てた正規形（ページ参照の title を剥がした後の doc）を保存することを
// 固定する（ReplacePageBlocksUseCase.Execute と同じ方針。StripPageRefTitles の doc 参照）。
func Test_提案作成_正規化後のdocを保存する(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	versionRepo := &mockPageVersionRepo{}
	versionRepo.On("GetLatestVersion", mock.Anything, kbWS, kbPage).Return(nil, nil)
	suggestions := &mockPageSuggestionRepo{}
	var created *domain.PageSuggestion
	suggestions.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { created = args.Get(1).(*domain.PageSuggestion) }).Return(nil)
	uc := kb.NewCreateSuggestionUseCase(kbRepo, versionRepo, suggestions)

	docWithTitle := `{"type":"doc","content":[{"type":"paragraph","content":[
		{"type":"pageRef","attrs":{"pageId":"` + kbPage + `","title":"読み手ごとの派生値"}}
	]}]}`
	_, err := uc.Execute(context.Background(), kb.CreateSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, Doc: docWithTitle, AuthorUserID: kbEditorUserID,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotContains(t, created.Doc, "読み手ごとの派生値", "pageRefのtitleは保存前に剥がされる")
	assert.Contains(t, created.Doc, "pageRef")
}

// Test_提案一覧_repoをそのまま呼ぶ は ListOpenPageSuggestionsUseCase が薄いラッパーである
// ことを固定する。
func Test_提案一覧_repoをそのまま呼ぶ(t *testing.T) {
	suggestions := &mockPageSuggestionRepo{}
	want := []domain.PageSuggestion{{ID: "s1", PageID: kbPage}, {ID: "s2", PageID: kbPage}}
	suggestions.On("ListOpen", mock.Anything, kbWS, kbPage).Return(want, nil)
	uc := kb.NewListOpenPageSuggestionsUseCase(suggestions)

	got, err := uc.Execute(context.Background(), kb.ListOpenPageSuggestionsInput{WorkspaceID: kbWS, PageID: kbPage})
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

// Test_提案採用_ForceVersionは必ずtrue は AcceptPageSuggestionUseCase が
// suggestions.Resolve の戻り値の Doc をそのまま ReplacePageBlocksUseCase.Execute の
// Input.Doc に渡し、ForceVersion を必ず true にして呼ぶことを固定する
// （RestorePageVersionUseCase の Test_復元_ForceVersionは必ずtrue と同じ検証パターン）。
func Test_提案採用_ForceVersionは必ずtrue(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(kbActivePage(kbPage, kbSpace, nil), nil)
	kbRepo.On("TouchPageLastEditedBy", mock.Anything, kbWS, kbPage, kbEditorUserID).Return(nil)
	kbRepo.On("ReplacePageBlocks", mock.Anything, kbWS, kbPage, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	kbRepo.On("GetPageSnapshot", mock.Anything, kbWS, kbPage).
		Return(&domain.PageSnapshot{PageID: kbPage, Doc: kbSuggestionDoc}, nil)

	versionRepo := &mockPageVersionRepo{}
	var gotForce bool
	versionRepo.On("CreateVersionIfDue", mock.Anything, kbWS, kbPage, mock.Anything, kbEditorUserID, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { gotForce = args.Bool(6) }).
		Return(true, &domain.PageVersion{PageID: kbPage, Seq: 2}, nil)

	suggestions := &mockPageSuggestionRepo{}
	resolved := &domain.PageSuggestion{ID: "s1", PageID: kbPage, Doc: kbSuggestionDoc, Status: domain.PageSuggestionStatusAccepted}
	suggestions.On("Resolve", mock.Anything, kbWS, kbPage, "s1", domain.PageSuggestionStatusAccepted, kbEditorUserID, mock.Anything).
		Return(resolved, nil)

	replaceUC := kb.NewReplacePageBlocksUseCase(kbRepo, &fakeTxManager{}, versionRepo)
	tx := &fakeTxManager{}
	uc := kb.NewAcceptPageSuggestionUseCase(suggestions, replaceUC, tx)

	got, err := uc.Execute(context.Background(), kb.AcceptSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, SuggestionID: "s1", ResolverUserID: kbEditorUserID,
	})
	require.NoError(t, err)
	assert.True(t, gotForce, "採用はForceVersionが必ずtrue（10分規則を無視して必ず版を切る）")
	assert.Same(t, resolved, got)
	assert.Equal(t, 1, tx.calls, "提案の解決と本文の書き換えは1つのトランザクションにまとめる")
}

// Test_提案採用_Resolveが失敗すれば本文を書き換えない は、suggestions.Resolve が失敗（既に
// 解決済み等）したら ReplacePageBlocksUseCase.Execute（＝ kbRepo）に一切触れないことを固定する。
func Test_提案採用_Resolveが失敗すれば本文を書き換えない(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	versionRepo := &mockPageVersionRepo{}
	suggestions := &mockPageSuggestionRepo{}
	suggestions.On("Resolve", mock.Anything, kbWS, kbPage, "s1", domain.PageSuggestionStatusAccepted, kbEditorUserID, mock.Anything).
		Return(nil, domain.ErrPageSuggestionAlreadyResolved)
	replaceUC := kb.NewReplacePageBlocksUseCase(kbRepo, &fakeTxManager{}, versionRepo)
	uc := kb.NewAcceptPageSuggestionUseCase(suggestions, replaceUC, &fakeTxManager{})

	_, err := uc.Execute(context.Background(), kb.AcceptSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, SuggestionID: "s1", ResolverUserID: kbEditorUserID,
	})
	require.ErrorIs(t, err, domain.ErrPageSuggestionAlreadyResolved)
	kbRepo.AssertNotCalled(t, "FindPage", mock.Anything, mock.Anything, mock.Anything)
}

// Test_提案採用_本文書き換えが失敗すればエラーを返す は、replaceBlocks.Execute の失敗が
// そのまま Execute のエラーとして返ることを固定する（同一トランザクションなので、実際の
// ロールバック自体は結合テスト側が確認する）。
func Test_提案採用_本文書き換えが失敗すればエラーを返す(t *testing.T) {
	kbRepo := &mockKnowledgeBaseRepo{}
	kbRepo.On("FindPage", mock.Anything, kbWS, kbPage).Return(nil, kb.ErrPageArchived)
	versionRepo := &mockPageVersionRepo{}
	suggestions := &mockPageSuggestionRepo{}
	resolved := &domain.PageSuggestion{ID: "s1", PageID: kbPage, Doc: kbSuggestionDoc}
	suggestions.On("Resolve", mock.Anything, kbWS, kbPage, "s1", domain.PageSuggestionStatusAccepted, kbEditorUserID, mock.Anything).
		Return(resolved, nil)
	replaceUC := kb.NewReplacePageBlocksUseCase(kbRepo, &fakeTxManager{}, versionRepo)
	uc := kb.NewAcceptPageSuggestionUseCase(suggestions, replaceUC, &fakeTxManager{})

	_, err := uc.Execute(context.Background(), kb.AcceptSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, SuggestionID: "s1", ResolverUserID: kbEditorUserID,
	})
	require.ErrorIs(t, err, kb.ErrPageArchived)
}

// Test_提案却下_repoをそのまま呼ぶ は RejectPageSuggestionUseCase が本文に一切触れず、
// suggestions.Resolve を rejected で呼ぶだけであることを固定する。
func Test_提案却下_repoをそのまま呼ぶ(t *testing.T) {
	suggestions := &mockPageSuggestionRepo{}
	resolved := &domain.PageSuggestion{ID: "s1", PageID: kbPage, Status: domain.PageSuggestionStatusRejected}
	var gotResolvedAt time.Time
	suggestions.On("Resolve", mock.Anything, kbWS, kbPage, "s1", domain.PageSuggestionStatusRejected, kbEditorUserID, mock.Anything).
		Run(func(args mock.Arguments) { gotResolvedAt = args.Get(6).(time.Time) }).
		Return(resolved, nil)
	uc := kb.NewRejectPageSuggestionUseCase(suggestions)

	got, err := uc.Execute(context.Background(), kb.RejectSuggestionInput{
		WorkspaceID: kbWS, PageID: kbPage, SuggestionID: "s1", ResolverUserID: kbEditorUserID,
	})
	require.NoError(t, err)
	assert.Same(t, resolved, got)
	assert.WithinDuration(t, time.Now(), gotResolvedAt, time.Second)
}
