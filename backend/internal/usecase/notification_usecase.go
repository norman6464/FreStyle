package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListNotificationsUseCase struct{ repo repository.NotificationRepository }

func NewListNotificationsUseCase(r repository.NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{repo: r}
}

func (u *ListNotificationsUseCase) Execute(ctx context.Context, userID uint64) ([]domain.Notification, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

type MarkNotificationReadUseCase struct{ repo repository.NotificationRepository }

func NewMarkNotificationReadUseCase(r repository.NotificationRepository) *MarkNotificationReadUseCase {
	return &MarkNotificationReadUseCase{repo: r}
}

func (u *MarkNotificationReadUseCase) Execute(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("id is required")
	}
	return u.repo.MarkRead(ctx, id)
}
