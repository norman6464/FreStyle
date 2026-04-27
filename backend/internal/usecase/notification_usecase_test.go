package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubNotificationRepo struct {
	rows []domain.Notification
	err  error
}

func (s *stubNotificationRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.Notification, error) {
	return s.rows, s.err
}
func (s *stubNotificationRepo) MarkRead(_ context.Context, _ uint64) error { return s.err }

func TestListNotifications_RequiresUserID(t *testing.T) {
	uc := NewListNotificationsUseCase(&stubNotificationRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestMarkRead_RequiresID(t *testing.T) {
	uc := NewMarkNotificationReadUseCase(&stubNotificationRepo{})
	if err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}
