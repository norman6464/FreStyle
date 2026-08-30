package usecase_test

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const validDoc = `{"type":"doc","content":[{"type":"paragraph"}]}`

func Test_GetRichDocument_認可(t *testing.T) {
	wsA := "0198a000-0000-7000-8000-0000000000d1"
	wsB := "0198a000-0000-7000-8000-0000000000d2"
	cases := []struct {
		name            string
		doc             *domain.RichDocument
		viewerID        uint64
		viewerWorkspace domain.WorkspaceRef
		wantErr         error
	}{
		{"所有者は自分の非公開を読める", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: false}, 7, domain.WorkspaceRefOf(wsA), nil},
		{"同一ワークスペースの他人は公開を読める", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: true, WorkspaceID: &wsA}, 99, domain.WorkspaceRefOf(wsA), nil},
		{"別ワークスペースの他人は公開を読めない(404)", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: true, WorkspaceID: &wsA}, 99, domain.WorkspaceRefOf(wsB), usecase.ErrRichDocumentNotFound},
		{"ワークスペース不明(NULL)の公開は他人から読めない(404)", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: true}, 99, domain.WorkspaceRefOf(wsA), usecase.ErrRichDocumentNotFound},
		{"所有者はワークスペースが別でも自分の文書を読める", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: true, WorkspaceID: &wsB}, 7, domain.WorkspaceRefOf(wsA), nil},
		{"他人は非公開を読めない(404)", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: false, WorkspaceID: &wsA}, 99, domain.WorkspaceRefOf(wsA), usecase.ErrRichDocumentNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRichDocRepo{}
			repo.On("FindByID", mock.Anything, "a").Return(tc.doc, nil).Once()
			uc := usecase.NewGetRichDocumentUseCase(repo)
			got, err := uc.Execute(context.Background(), "a", tc.viewerID, tc.viewerWorkspace)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "a", got.ID)
			repo.AssertExpectations(t)
		})
	}
}

func Test_GetRichDocument_存在しない(t *testing.T) {
	repo := &mockRichDocRepo{}
	repo.On("FindByID", mock.Anything, "x").Return((*domain.RichDocument)(nil), repository.ErrRichDocumentNotFound).Once()
	uc := usecase.NewGetRichDocumentUseCase(repo)
	_, err := uc.Execute(context.Background(), "x", 7, domain.NoWorkspace())
	assert.ErrorIs(t, err, usecase.ErrRichDocumentNotFound)
	repo.AssertExpectations(t)
}

func Test_CreateRichDocument_成功(t *testing.T) {
	repo := &mockRichDocRepo{}
	// 渡された doc の中身を型付きで確認し、ID 未設定なら採番して書き戻す契約を Run で再現する。
	repo.On("Create", mock.Anything, mock.MatchedBy(func(d *domain.RichDocument) bool {
		return d.OwnerID == 7 && d.Kind == domain.DocumentKindNote && d.Doc == validDoc
	})).
		Run(func(args mock.Arguments) {
			doc := args.Get(1).(*domain.RichDocument)
			if doc.ID == "" {
				doc.ID = "generated-uuid"
			}
		}).
		Return(nil).
		Once()

	uc := usecase.NewCreateRichDocumentUseCase(repo)
	got, err := uc.Execute(context.Background(), usecase.CreateRichDocumentInput{
		OwnerID: 7, Kind: domain.DocumentKindNote, Title: "メモ", Doc: validDoc,
	})
	require.NoError(t, err)
	assert.Equal(t, "generated-uuid", got.ID)
	assert.Equal(t, 1, got.Revision)
	assert.Equal(t, 1, got.SchemaVersion)
	repo.AssertExpectations(t)
}

func Test_CreateRichDocument_バリデーション(t *testing.T) {
	cases := []struct {
		name string
		in   usecase.CreateRichDocumentInput
	}{
		{"未知kind", usecase.CreateRichDocumentInput{OwnerID: 7, Kind: "weird", Title: "t", Doc: validDoc}},
		{"title空", usecase.CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "", Doc: validDoc}},
		{"doc空", usecase.CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: ""}},
		{"docがobjectでない", usecase.CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: `[1,2]`}},
		{"doc.typeがdocでない", usecase.CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: `{"type":"paragraph"}`}},
		{"ownerID0", usecase.CreateRichDocumentInput{OwnerID: 0, Kind: domain.DocumentKindNote, Title: "t", Doc: validDoc}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// バリデーション失敗時は repo を一切呼ばない（On を設定しないので呼べば panic する＝検証になる）。
			repo := &mockRichDocRepo{}
			uc := usecase.NewCreateRichDocumentUseCase(repo)
			_, err := uc.Execute(context.Background(), tc.in)
			assert.ErrorIs(t, err, usecase.ErrRichDocumentInvalid)
			repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}

func Test_RichDocument_NUL(t *testing.T) {
	t.Run("エスケープU+0000のdocは拒否", func(t *testing.T) {
		repo := &mockRichDocRepo{}
		uc := usecase.NewCreateRichDocumentUseCase(repo)
		_, err := uc.Execute(context.Background(), usecase.CreateRichDocumentInput{
			OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t",
			Doc: `{"type":"doc","content":[{"type":"text","text":"\u0000"}]}`,
		})
		assert.ErrorIs(t, err, usecase.ErrRichDocumentInvalid)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
	t.Run("リテラルNULバイトのdocは拒否", func(t *testing.T) {
		repo := &mockRichDocRepo{}
		uc := usecase.NewCreateRichDocumentUseCase(repo)
		_, err := uc.Execute(context.Background(), usecase.CreateRichDocumentInput{
			OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t",
			Doc: "{\"type\":\"doc\",\"content\":[{\"type\":\"text\",\"text\":\"a\x00b\"}]}",
		})
		assert.ErrorIs(t, err, usecase.ErrRichDocumentInvalid)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
	t.Run("バックスラッシュu0000リテラル文字列は受理", func(t *testing.T) {
		// JSON 値 "\\u0000" はデコードすると 6 文字の文字列（NUL ではない）。誤検知で弾いてはいけない。
		repo := &mockRichDocRepo{}
		repo.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				doc := args.Get(1).(*domain.RichDocument)
				if doc.ID == "" {
					doc.ID = "generated-uuid"
				}
			}).Return(nil).Once()
		uc := usecase.NewCreateRichDocumentUseCase(repo)
		_, err := uc.Execute(context.Background(), usecase.CreateRichDocumentInput{
			OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t",
			Doc: `{"type":"doc","content":[{"type":"text","text":"\\u0000"}]}`,
		})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func Test_UpdateRichDocument_成功(t *testing.T) {
	repo := &mockRichDocRepo{}
	repo.On("FindByID", mock.Anything, "a").
		Return(&domain.RichDocument{ID: "a", OwnerID: 7, SchemaVersion: 1, Revision: 3}, nil).Once()
	// expectedRevision=3 を第3引数で固定。成功時は repository が revision を +1 して書き戻す契約を Run で再現。
	repo.On("UpdateWithRevision", mock.Anything, mock.AnythingOfType("*domain.RichDocument"), 3).
		Run(func(args mock.Arguments) {
			doc := args.Get(1).(*domain.RichDocument)
			doc.Revision = 4
		}).Return(nil).Once()

	uc := usecase.NewUpdateRichDocumentUseCase(repo)
	got, err := uc.Execute(context.Background(), usecase.UpdateRichDocumentInput{
		ID: "a", ActorID: 7, Title: "new", Doc: validDoc, Revision: 3,
	})
	require.NoError(t, err)
	assert.Equal(t, 4, got.Revision)
	assert.Equal(t, "new", got.Title)
	repo.AssertExpectations(t)
}

func Test_UpdateRichDocument_他人は存在を漏らさず404(t *testing.T) {
	repo := &mockRichDocRepo{}
	repo.On("FindByID", mock.Anything, "a").
		Return(&domain.RichDocument{ID: "a", OwnerID: 7}, nil).Once()
	uc := usecase.NewUpdateRichDocumentUseCase(repo)
	_, err := uc.Execute(context.Background(), usecase.UpdateRichDocumentInput{
		ID: "a", ActorID: 99, Title: "x", Doc: validDoc, Revision: 1,
	})
	assert.ErrorIs(t, err, usecase.ErrRichDocumentNotFound)
	repo.AssertNotCalled(t, "UpdateWithRevision", mock.Anything, mock.Anything, mock.Anything)
}

func Test_UpdateRichDocument_負のrevisionは400(t *testing.T) {
	repo := &mockRichDocRepo{}
	uc := usecase.NewUpdateRichDocumentUseCase(repo)
	_, err := uc.Execute(context.Background(), usecase.UpdateRichDocumentInput{
		ID: "a", ActorID: 7, Title: "x", Doc: validDoc, Revision: -1,
	})
	assert.ErrorIs(t, err, usecase.ErrRichDocumentInvalid)
	// revision 検証は FindByID より前で弾くので repo は一切呼ばれない。
	repo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "UpdateWithRevision", mock.Anything, mock.Anything, mock.Anything)
}

func Test_UpdateRichDocument_版不一致は409(t *testing.T) {
	repo := &mockRichDocRepo{}
	repo.On("FindByID", mock.Anything, "a").
		Return(&domain.RichDocument{ID: "a", OwnerID: 7, Revision: 5}, nil).Once()
	repo.On("UpdateWithRevision", mock.Anything, mock.Anything, 3).
		Return(repository.ErrRichDocumentConflict).Once()
	uc := usecase.NewUpdateRichDocumentUseCase(repo)
	_, err := uc.Execute(context.Background(), usecase.UpdateRichDocumentInput{
		ID: "a", ActorID: 7, Title: "x", Doc: validDoc, Revision: 3,
	})
	assert.ErrorIs(t, err, usecase.ErrRichDocumentConflict)
	repo.AssertExpectations(t)
}

func Test_UpdateRichDocument_存在しない(t *testing.T) {
	repo := &mockRichDocRepo{}
	repo.On("FindByID", mock.Anything, "x").
		Return((*domain.RichDocument)(nil), repository.ErrRichDocumentNotFound).Once()
	uc := usecase.NewUpdateRichDocumentUseCase(repo)
	_, err := uc.Execute(context.Background(), usecase.UpdateRichDocumentInput{
		ID: "x", ActorID: 7, Title: "x", Doc: validDoc, Revision: 1,
	})
	assert.ErrorIs(t, err, usecase.ErrRichDocumentNotFound)
	repo.AssertExpectations(t)
}

func Test_DeleteRichDocument(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		repo := &mockRichDocRepo{}
		repo.On("SoftDelete", mock.Anything, "a", uint64(7)).Return(nil).Once()
		uc := usecase.NewDeleteRichDocumentUseCase(repo)
		require.NoError(t, uc.Execute(context.Background(), "a", 7))
		repo.AssertExpectations(t)
	})
	t.Run("未認証はforbidden", func(t *testing.T) {
		repo := &mockRichDocRepo{}
		uc := usecase.NewDeleteRichDocumentUseCase(repo)
		err := uc.Execute(context.Background(), "a", 0)
		assert.ErrorIs(t, err, usecase.ErrRichDocumentForbidden)
		repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("存在しない(他人)は404", func(t *testing.T) {
		repo := &mockRichDocRepo{}
		repo.On("SoftDelete", mock.Anything, "a", uint64(7)).Return(repository.ErrRichDocumentNotFound).Once()
		uc := usecase.NewDeleteRichDocumentUseCase(repo)
		err := uc.Execute(context.Background(), "a", 7)
		assert.ErrorIs(t, err, usecase.ErrRichDocumentNotFound)
		repo.AssertExpectations(t)
	})
}

func Test_ListRichDocuments_成功(t *testing.T) {
	repo := &mockRichDocRepo{}
	rows := []domain.RichDocument{
		{ID: "a", OwnerID: 7, Kind: domain.DocumentKindNote, Title: "A"},
		{ID: "b", OwnerID: 7, Kind: domain.DocumentKindNote, Title: "B"},
	}
	// kind 未指定は空文字で repo に渡す（全 kind）。
	repo.On("ListByOwner", mock.Anything, uint64(7), domain.DocumentKind("")).Return(rows, nil).Once()
	uc := usecase.NewListRichDocumentsUseCase(repo)
	got, err := uc.Execute(context.Background(), usecase.ListRichDocumentsInput{OwnerID: 7})
	require.NoError(t, err)
	assert.Len(t, got, 2)
	repo.AssertExpectations(t)
}

func Test_ListRichDocuments_kind指定で絞る(t *testing.T) {
	repo := &mockRichDocRepo{}
	repo.On("ListByOwner", mock.Anything, uint64(7), domain.DocumentKindNote).
		Return([]domain.RichDocument{{ID: "a", OwnerID: 7, Kind: domain.DocumentKindNote}}, nil).Once()
	uc := usecase.NewListRichDocumentsUseCase(repo)
	got, err := uc.Execute(context.Background(), usecase.ListRichDocumentsInput{OwnerID: 7, Kind: domain.DocumentKindNote})
	require.NoError(t, err)
	assert.Len(t, got, 1)
	repo.AssertExpectations(t)
}

func Test_ListRichDocuments_不正入力(t *testing.T) {
	t.Run("ownerID0は400", func(t *testing.T) {
		repo := &mockRichDocRepo{}
		uc := usecase.NewListRichDocumentsUseCase(repo)
		_, err := uc.Execute(context.Background(), usecase.ListRichDocumentsInput{OwnerID: 0})
		assert.ErrorIs(t, err, usecase.ErrRichDocumentInvalid)
		repo.AssertNotCalled(t, "ListByOwner", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("kind不正は400", func(t *testing.T) {
		repo := &mockRichDocRepo{}
		uc := usecase.NewListRichDocumentsUseCase(repo)
		_, err := uc.Execute(context.Background(), usecase.ListRichDocumentsInput{OwnerID: 7, Kind: "weird"})
		assert.ErrorIs(t, err, usecase.ErrRichDocumentInvalid)
		repo.AssertNotCalled(t, "ListByOwner", mock.Anything, mock.Anything, mock.Anything)
	})
}

// Test_RichDocumentRepository_読み取り口の一覧を固定する は、rich_documents を読む入口が
// この port の既知メソッドだけであることを固定する。usecase は必ずこの interface 越しに
// 読むので、新しい読み取り経路を足すときは必ずここが増えて落ちる。落ちたら
// 「その経路は domain.RichDocument.CanBeReadBy を通っているか」を確かめてから一覧を更新する。
func Test_RichDocumentRepository_読み取り口の一覧を固定する(t *testing.T) {
	want := []string{"Create", "FindByID", "ListByOwner", "SoftDelete", "UpdateWithRevision"}
	typ := reflect.TypeOf((*repository.RichDocumentRepository)(nil)).Elem()
	got := make([]string, 0, typ.NumMethod())
	for i := range typ.NumMethod() {
		got = append(got, typ.Method(i).Name)
	}
	sort.Strings(got)
	assert.Equal(t, want, got,
		"rich_documents の読み書き口が変わった。読み取りを足したなら CanBeReadBy を通してから一覧を更新すること")
}
