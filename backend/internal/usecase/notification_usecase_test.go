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
func (s *stubNotificationRepo) MarkRead(_ context.Context, _, _ uint64) error { return s.err }
func (s *stubNotificationRepo) MarkAllRead(_ context.Context, _ uint64) error  { return s.err }
func (s *stubNotificationRepo) CountUnread(_ context.Context, _ uint64) (int64, error) {
	return int64(len(s.rows)), s.err
}

func TestListNotifications_RequiresUserID(t *testing.T) {
	uc := NewListNotificationsUseCase(&stubNotificationRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestMarkRead_RequiresUserID(t *testing.T) {
	uc := NewMarkNotificationReadUseCase(&stubNotificationRepo{})
	if err := uc.Execute(context.Background(), 0, 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestMarkRead_RequiresID(t *testing.T) {
	uc := NewMarkNotificationReadUseCase(&stubNotificationRepo{})
	if err := uc.Execute(context.Background(), 1, 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestMarkAllNotificationsRead_RequiresUserID(t *testing.T) {
	uc := NewMarkAllNotificationsReadUseCase(&stubNotificationRepo{})
	if err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestCountUnreadNotifications_RequiresUserID(t *testing.T) {
	uc := NewCountUnreadNotificationsUseCase(&stubNotificationRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}
