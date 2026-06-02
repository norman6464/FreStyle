package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListNotificationsUseCase は current user の通知一覧を返す。
type ListNotificationsUseCase struct {
	repo repository.NotificationRepository
}

func NewListNotificationsUseCase(r repository.NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{repo: r}
}

func (u *ListNotificationsUseCase) Execute(ctx context.Context, userID uint64) ([]domain.Notification, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

// MarkNotificationReadUseCase は単一通知を既読化する（所有者検証込み）。
type MarkNotificationReadUseCase struct {
	repo repository.NotificationRepository
}

func NewMarkNotificationReadUseCase(r repository.NotificationRepository) *MarkNotificationReadUseCase {
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

// MarkAllNotificationsReadUseCase は current user の全通知を一括既読化する。
type MarkAllNotificationsReadUseCase struct {
	repo repository.NotificationRepository
}

func NewMarkAllNotificationsReadUseCase(r repository.NotificationRepository) *MarkAllNotificationsReadUseCase {
	return &MarkAllNotificationsReadUseCase{repo: r}
}

func (u *MarkAllNotificationsReadUseCase) Execute(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return errors.New("userID is required")
	}
	return u.repo.MarkAllRead(ctx, userID)
}

// CountUnreadNotificationsUseCase は current user の未読通知数を返す（バッジ表示用）。
type CountUnreadNotificationsUseCase struct {
	repo repository.NotificationRepository
}

func NewCountUnreadNotificationsUseCase(r repository.NotificationRepository) *CountUnreadNotificationsUseCase {
	return &CountUnreadNotificationsUseCase{repo: r}
}

func (u *CountUnreadNotificationsUseCase) Execute(ctx context.Context, userID uint64) (int64, error) {
	if userID == 0 {
		return 0, errors.New("userID is required")
	}
	return u.repo.CountUnread(ctx, userID)
}
