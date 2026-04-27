package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubConversationTemplateRepo struct {
	rows []domain.ConversationTemplate
	err  error
}

func (s *stubConversationTemplateRepo) ListActive(_ context.Context) ([]domain.ConversationTemplate, error) {
	return s.rows, s.err
}

func TestListConversationTemplates_Returns(t *testing.T) {
	uc := NewListConversationTemplatesUseCase(&stubConversationTemplateRepo{rows: []domain.ConversationTemplate{{ID: 1}}})
	got, err := uc.Execute(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestListConversationTemplates_Error(t *testing.T) {
	uc := NewListConversationTemplatesUseCase(&stubConversationTemplateRepo{err: errors.New("db")})
	if _, err := uc.Execute(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
