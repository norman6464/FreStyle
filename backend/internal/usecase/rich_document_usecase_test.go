package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeRichDocRepo は RichDocumentRepository の手書きフェイク。
type fakeRichDocRepo struct {
	getDoc  *domain.RichDocument
	getErr  error
	created *domain.RichDocument

	createErr  error
	updateErr  error
	updateDoc  *domain.RichDocument // 成功時に doc へ反映する値
	updatedID  string
	updatedRev int

	deleteErr    error
	deletedID    string
	deletedOwner uint64
}

func (f *fakeRichDocRepo) Create(_ context.Context, doc *domain.RichDocument) error {
	if f.createErr != nil {
		return f.createErr
	}
	if doc.ID == "" {
		doc.ID = "generated-uuid"
	}
	f.created = doc
	return nil
}

func (f *fakeRichDocRepo) FindByID(_ context.Context, _ string) (*domain.RichDocument, error) {
	return f.getDoc, f.getErr
}

func (f *fakeRichDocRepo) UpdateWithRevision(_ context.Context, doc *domain.RichDocument, expected int) error {
	f.updatedID = doc.ID
	f.updatedRev = expected
	if f.updateErr != nil {
		return f.updateErr
	}
	if f.updateDoc != nil {
		*doc = *f.updateDoc
	}
	return nil
}

func (f *fakeRichDocRepo) SoftDelete(_ context.Context, id string, ownerID uint64) error {
	f.deletedID = id
	f.deletedOwner = ownerID
	return f.deleteErr
}

const validDoc = `{"type":"doc","content":[{"type":"paragraph"}]}`

func Test_GetRichDocument_認可(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name     string
		doc      *domain.RichDocument
		viewerID uint64
		wantErr  error
	}{
		{"所有者は自分の非公開を読める", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: false}, 7, nil},
		{"他人は公開を読める", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: true}, 99, nil},
		{"他人は非公開を読めない(404)", &domain.RichDocument{ID: "a", OwnerID: 7, IsPublic: false}, 99, ErrRichDocumentNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewGetRichDocumentUseCase(&fakeRichDocRepo{getDoc: tc.doc})
			got, err := uc.Execute(ctx, "a", tc.viewerID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && (got == nil || got.ID != "a") {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

func Test_GetRichDocument_存在しない(t *testing.T) {
	uc := NewGetRichDocumentUseCase(&fakeRichDocRepo{getErr: repository.ErrRichDocumentNotFound})
	_, err := uc.Execute(context.Background(), "x", 7)
	if !errors.Is(err, ErrRichDocumentNotFound) {
		t.Fatalf("err = %v, want ErrRichDocumentNotFound", err)
	}
}

func Test_CreateRichDocument_成功(t *testing.T) {
	repo := &fakeRichDocRepo{}
	uc := NewCreateRichDocumentUseCase(repo)
	got, err := uc.Execute(context.Background(), CreateRichDocumentInput{
		OwnerID: 7, Kind: domain.DocumentKindNote, Title: "メモ", Doc: validDoc,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.OwnerID != 7 || got.Kind != domain.DocumentKindNote || got.Revision != 1 || got.SchemaVersion != 1 {
		t.Fatalf("unexpected doc: %+v", got)
	}
	if repo.created == nil || repo.created.Doc != validDoc {
		t.Fatalf("repo.Create not called with doc")
	}
}

func Test_CreateRichDocument_バリデーション(t *testing.T) {
	uc := NewCreateRichDocumentUseCase(&fakeRichDocRepo{})
	cases := []struct {
		name string
		in   CreateRichDocumentInput
	}{
		{"未知kind", CreateRichDocumentInput{OwnerID: 7, Kind: "weird", Title: "t", Doc: validDoc}},
		{"title空", CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "", Doc: validDoc}},
		{"doc空", CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: ""}},
		{"docがobjectでない", CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: `[1,2]`}},
		{"doc.typeがdocでない", CreateRichDocumentInput{OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: `{"type":"paragraph"}`}},
		{"ownerID0", CreateRichDocumentInput{OwnerID: 0, Kind: domain.DocumentKindNote, Title: "t", Doc: validDoc}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := uc.Execute(context.Background(), tc.in); !errors.Is(err, ErrRichDocumentInvalid) {
				t.Fatalf("err = %v, want ErrRichDocumentInvalid", err)
			}
		})
	}
}

func Test_CreateRichDocument_サイズ上限(t *testing.T) {
	big := `{"type":"doc","x":"` + strings.Repeat("a", maxDocBytes) + `"}`
	uc := NewCreateRichDocumentUseCase(&fakeRichDocRepo{})
	_, err := uc.Execute(context.Background(), CreateRichDocumentInput{
		OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: big,
	})
	if !errors.Is(err, ErrRichDocumentInvalid) {
		t.Fatalf("err = %v, want ErrRichDocumentInvalid (size)", err)
	}
}

func Test_UpdateRichDocument_成功(t *testing.T) {
	repo := &fakeRichDocRepo{
		getDoc:    &domain.RichDocument{ID: "a", OwnerID: 7, SchemaVersion: 1, Revision: 3},
		updateDoc: &domain.RichDocument{ID: "a", OwnerID: 7, Title: "new", Revision: 4},
	}
	uc := NewUpdateRichDocumentUseCase(repo)
	got, err := uc.Execute(context.Background(), UpdateRichDocumentInput{
		ID: "a", ActorID: 7, Title: "new", Doc: validDoc, Revision: 3,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Revision != 4 || got.Title != "new" {
		t.Fatalf("unexpected: %+v", got)
	}
	if repo.updatedRev != 3 {
		t.Fatalf("expected revision passed = %d, want 3", repo.updatedRev)
	}
}

func Test_UpdateRichDocument_他人は存在を漏らさず404(t *testing.T) {
	repo := &fakeRichDocRepo{getDoc: &domain.RichDocument{ID: "a", OwnerID: 7}}
	uc := NewUpdateRichDocumentUseCase(repo)
	_, err := uc.Execute(context.Background(), UpdateRichDocumentInput{
		ID: "a", ActorID: 99, Title: "x", Doc: validDoc, Revision: 1,
	})
	if !errors.Is(err, ErrRichDocumentNotFound) {
		t.Fatalf("err = %v, want ErrRichDocumentNotFound (存在を漏らさない)", err)
	}
}

func Test_RichDocument_NULを拒否する(t *testing.T) {
	uc := NewCreateRichDocumentUseCase(&fakeRichDocRepo{})
	cases := map[string]string{
		"docにリテラルNUL":  "{\"type\":\"doc\",\"content\":[{\"type\":\"text\",\"text\":\"a\x00b\"}]}",
		"docにエスケープNUL": `{"type":"doc","content":[{"type":"text","text":"\u0000"}]}`,
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := uc.Execute(context.Background(), CreateRichDocumentInput{
				OwnerID: 7, Kind: domain.DocumentKindNote, Title: "t", Doc: doc,
			})
			if !errors.Is(err, ErrRichDocumentInvalid) {
				t.Fatalf("err = %v, want ErrRichDocumentInvalid (NUL)", err)
			}
		})
	}
}

func Test_UpdateRichDocument_版不一致は409(t *testing.T) {
	repo := &fakeRichDocRepo{
		getDoc:    &domain.RichDocument{ID: "a", OwnerID: 7, Revision: 5},
		updateErr: repository.ErrRichDocumentConflict,
	}
	uc := NewUpdateRichDocumentUseCase(repo)
	_, err := uc.Execute(context.Background(), UpdateRichDocumentInput{
		ID: "a", ActorID: 7, Title: "x", Doc: validDoc, Revision: 3,
	})
	if !errors.Is(err, ErrRichDocumentConflict) {
		t.Fatalf("err = %v, want ErrRichDocumentConflict", err)
	}
}

func Test_UpdateRichDocument_存在しない(t *testing.T) {
	uc := NewUpdateRichDocumentUseCase(&fakeRichDocRepo{getErr: repository.ErrRichDocumentNotFound})
	_, err := uc.Execute(context.Background(), UpdateRichDocumentInput{
		ID: "x", ActorID: 7, Title: "x", Doc: validDoc, Revision: 1,
	})
	if !errors.Is(err, ErrRichDocumentNotFound) {
		t.Fatalf("err = %v, want ErrRichDocumentNotFound", err)
	}
}

func Test_DeleteRichDocument(t *testing.T) {
	t.Run("成功", func(t *testing.T) {
		repo := &fakeRichDocRepo{}
		uc := NewDeleteRichDocumentUseCase(repo)
		if err := uc.Execute(context.Background(), "a", 7); err != nil {
			t.Fatalf("err: %v", err)
		}
		if repo.deletedID != "a" || repo.deletedOwner != 7 {
			t.Fatalf("deleted id=%q owner=%d", repo.deletedID, repo.deletedOwner)
		}
	})
	t.Run("未認証はforbidden", func(t *testing.T) {
		uc := NewDeleteRichDocumentUseCase(&fakeRichDocRepo{})
		if err := uc.Execute(context.Background(), "a", 0); !errors.Is(err, ErrRichDocumentForbidden) {
			t.Fatalf("err = %v, want forbidden", err)
		}
	})
	t.Run("存在しない(他人)は404", func(t *testing.T) {
		uc := NewDeleteRichDocumentUseCase(&fakeRichDocRepo{deleteErr: repository.ErrRichDocumentNotFound})
		if err := uc.Execute(context.Background(), "a", 7); !errors.Is(err, ErrRichDocumentNotFound) {
			t.Fatalf("err = %v, want not found", err)
		}
	})
}
