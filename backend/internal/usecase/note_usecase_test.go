package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubNoteRepo struct {
	rows []domain.Note
	one  *domain.Note
	err  error
}

func (s *stubNoteRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.Note, error) {
	return s.rows, s.err
}
func (s *stubNoteRepo) FindByID(_ context.Context, _ uint64) (*domain.Note, error) { return s.one, s.err }
func (s *stubNoteRepo) Create(_ context.Context, n *domain.Note) error {
	if s.err != nil {
		return s.err
	}
	n.ID = 21
	return nil
}
func (s *stubNoteRepo) Update(_ context.Context, _ *domain.Note) error { return s.err }
func (s *stubNoteRepo) Delete(_ context.Context, _ uint64) error       { return s.err }

func TestListNotes_RequiresUserID(t *testing.T) {
	uc := NewListNotesByUserIDUseCase(&stubNoteRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateNote_RequiresTitle(t *testing.T) {
	uc := NewCreateNoteUseCase(&stubNoteRepo{})
	if _, err := uc.Execute(context.Background(), CreateNoteInput{UserID: 1}); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateNote_AssignsID(t *testing.T) {
	uc := NewCreateNoteUseCase(&stubNoteRepo{})
	got, err := uc.Execute(context.Background(), CreateNoteInput{UserID: 1, Title: "t"})
	if err != nil || got.ID != 21 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestUpdateNote_RequiresID(t *testing.T) {
	uc := NewUpdateNoteUseCase(&stubNoteRepo{one: &domain.Note{ID: 1}})
	if _, err := uc.Execute(context.Background(), UpdateNoteInput{Title: "x"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteNote_RequiresID(t *testing.T) {
	uc := NewDeleteNoteUseCase(&stubNoteRepo{})
	if err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}
