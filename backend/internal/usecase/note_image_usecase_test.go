package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubPresigner struct {
	url *domain.NoteImageUploadURL
	err error
}

func (s *stubPresigner) Generate(_ context.Context, _ uint64, _ string) (*domain.NoteImageUploadURL, error) {
	return s.url, s.err
}

func TestIssueNoteImageUploadURL_RequiresUserID(t *testing.T) {
	uc := NewIssueNoteImageUploadURLUseCase(&stubPresigner{})
	if _, err := uc.Execute(context.Background(), 0, "image/png"); err == nil {
		t.Fatal("expected error")
	}
}

func TestIssueNoteImageUploadURL_Returns(t *testing.T) {
	uc := NewIssueNoteImageUploadURLUseCase(&stubPresigner{
		url: &domain.NoteImageUploadURL{URL: "https://example", Key: "k", ExpiresIn: 60},
	})
	got, err := uc.Execute(context.Background(), 1, "image/png")
	if err != nil || got.URL == "" {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
