package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubNoteRepo struct {
	rows          []domain.Note
	one           *domain.Note
	err           error
	deletedUserID uint64
	deletedID     uint64
	updatedNote   *domain.Note
}

func (s *stubNoteRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.Note, error) {
	return s.rows, s.err
}

func (s *stubNoteRepo) FindByID(_ context.Context, _ uint64) (*domain.Note, error) {
	return s.one, s.err
}

func (s *stubNoteRepo) Create(_ context.Context, n *domain.Note) error {
	if s.err != nil {
		return s.err
	}
	n.ID = 21
	return nil
}

func (s *stubNoteRepo) Update(_ context.Context, n *domain.Note) error {
	if s.err != nil {
		return s.err
	}
	s.updatedNote = n
	return nil
}

func (s *stubNoteRepo) Delete(_ context.Context, userID, id uint64) error {
	s.deletedUserID = userID
	s.deletedID = id
	return s.err
}

func Test_ノート一覧_ユーザーIDが必須(t *testing.T) {
	uc := NewListNotesByUserIDUseCase(&stubNoteRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func Test_ノート作成_タイトルが必須(t *testing.T) {
	uc := NewCreateNoteUseCase(&stubNoteRepo{}, &nopActivityRepo{})
	if _, err := uc.Execute(context.Background(), CreateNoteInput{UserID: 1}); err == nil {
		t.Fatal("expected error")
	}
}

func Test_ノート作成_IDを採番(t *testing.T) {
	uc := NewCreateNoteUseCase(&stubNoteRepo{}, &nopActivityRepo{})
	got, err := uc.Execute(context.Background(), CreateNoteInput{UserID: 1, Title: "t"})
	if err != nil || got.ID != 21 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func Test_ノート更新_ユーザーIDが必須(t *testing.T) {
	uc := NewUpdateNoteUseCase(&stubNoteRepo{one: &domain.Note{ID: 1, UserID: 1}})
	if _, err := uc.Execute(context.Background(), UpdateNoteInput{ID: 1, Title: "x"}); err == nil {
		t.Fatal("expected error when userID is 0")
	}
}

func Test_ノート更新_IDが必須(t *testing.T) {
	uc := NewUpdateNoteUseCase(&stubNoteRepo{one: &domain.Note{ID: 1, UserID: 1}})
	if _, err := uc.Execute(context.Background(), UpdateNoteInput{UserID: 1, Title: "x"}); err == nil {
		t.Fatal("expected error when id is 0")
	}
}

func Test_ノート更新_他人の所有を拒否(t *testing.T) {
	repo := &stubNoteRepo{one: &domain.Note{ID: 1, UserID: 99}}
	uc := NewUpdateNoteUseCase(repo)
	_, err := uc.Execute(context.Background(), UpdateNoteInput{UserID: 1, ID: 1, Title: "x"})
	if !errors.Is(err, ErrNoteForbidden) {
		t.Fatalf("expected ErrNoteForbidden, got %v", err)
	}
	if repo.updatedNote != nil {
		t.Fatal("must not call repo.Update when foreign owner")
	}
}

func Test_ノート更新_所有者は許可(t *testing.T) {
	repo := &stubNoteRepo{one: &domain.Note{ID: 1, UserID: 1, Title: "old"}}
	uc := NewUpdateNoteUseCase(repo)
	got, err := uc.Execute(context.Background(), UpdateNoteInput{UserID: 1, ID: 1, Title: "new"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Title != "new" {
		t.Fatalf("title should be updated, got %q", got.Title)
	}
	if repo.updatedNote == nil {
		t.Fatal("repo.Update should be called")
	}
}

func Test_ノート削除_ユーザーIDが必須(t *testing.T) {
	uc := NewDeleteNoteUseCase(&stubNoteRepo{})
	if err := uc.Execute(context.Background(), 0, 1); err == nil {
		t.Fatal("expected error when userID is 0")
	}
}

func Test_ノート削除_IDが必須(t *testing.T) {
	uc := NewDeleteNoteUseCase(&stubNoteRepo{})
	if err := uc.Execute(context.Background(), 1, 0); err == nil {
		t.Fatal("expected error when id is 0")
	}
}

func Test_ノート削除_ユーザーIDをリポジトリへ渡す(t *testing.T) {
	repo := &stubNoteRepo{}
	uc := NewDeleteNoteUseCase(repo)
	if err := uc.Execute(context.Background(), 7, 11); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.deletedUserID != 7 || repo.deletedID != 11 {
		t.Fatalf("repo.Delete should be called with (7,11), got (%d,%d)", repo.deletedUserID, repo.deletedID)
	}
}
