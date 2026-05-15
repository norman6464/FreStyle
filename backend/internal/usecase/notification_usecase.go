package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/legacyrepository"
)

type ListNotificationsUseCase struct {
	repo legacyrepository.NotificationRepository
}

func NewListNotificationsUseCase(r legacyrepository.NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{repo: r}
}

func (u *ListNotificationsUseCase) Execute(ctx context.Context, userID uint64) ([]domain.Notification, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

type MarkNotificationReadUseCase struct {
	repo legacyrepository.NotificationRepository
}

func NewMarkNotificationReadUseCase(r legacyrepository.NotificationRepository) *MarkNotificationReadUseCase {
	return &MarkNotificationReadUseCase{repo: r}
}

func (u *MarkNotificationReadUseCase) Execute(ctx context.Context, userID, id uint64) error {
	if userID == 0 {
		return errors.New("userID is required")
	}
	if id == 0 {
		return errors.New("id is required")
	}
	return u.repo.MarkRead(ctx, userID, id)
}

type MarkAllNotificationsReadUseCase struct {
	repo legacyrepository.NotificationRepository
}

func NewMarkAllNotificationsReadUseCase(r legacyrepository.NotificationRepository) *MarkAllNotificationsReadUseCase {
	return &MarkAllNotificationsReadUseCase{repo: r}
}

func (u *MarkAllNotificationsReadUseCase) Execute(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return errors.New("userID is required")
	}
	return u.repo.MarkAllRead(ctx, userID)
}

type CountUnreadNotificationsUseCase struct {
	repo legacyrepository.NotificationRepository
}

func NewCountUnreadNotificationsUseCase(r legacyrepository.NotificationRepository) *CountUnreadNotificationsUseCase {
	return &CountUnreadNotificationsUseCase{repo: r}
}

func (u *CountUnreadNotificationsUseCase) Execute(ctx context.Context, userID uint64) (int64, error) {
	if userID == 0 {
		return 0, errors.New("userID is required")
	}
	return u.repo.CountUnread(ctx, userID)
}
