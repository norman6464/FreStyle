package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubSessionNoteRepo struct {
	n   *domain.SessionNote
	err error
}

func (s *stubSessionNoteRepo) FindBySessionID(_ context.Context, _ uint64) (*domain.SessionNote, error) {
	return s.n, s.err
}

func (s *stubSessionNoteRepo) Upsert(_ context.Context, n *domain.SessionNote) error {
	if s.err != nil {
		return s.err
	}
	s.n = n
	return nil
}

func Test_セッションノート取得_セッションIDが必須(t *testing.T) {
	uc := NewGetSessionNoteUseCase(&stubSessionNoteRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func Test_セッションノート保存_バリデーション(t *testing.T) {
	uc := NewUpsertSessionNoteUseCase(&stubSessionNoteRepo{})
	if _, err := uc.Execute(context.Background(), UpsertSessionNoteInput{}); err == nil {
		t.Fatal("expected error")
	}
}

func Test_セッションノート保存_永続化する(t *testing.T) {
	repo := &stubSessionNoteRepo{}
	uc := NewUpsertSessionNoteUseCase(repo)
	got, err := uc.Execute(context.Background(), UpsertSessionNoteInput{SessionID: 1, UserID: 2, Content: "hello"})
	if err != nil || got.Content != "hello" {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
