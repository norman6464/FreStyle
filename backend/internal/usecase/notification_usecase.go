package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListNotificationsUseCase は current user の 通知 一覧 を 返す。
// 依存 port: [repository.NotificationRepository]。
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

// MarkNotificationReadUseCase は 単一 通知 を 既読 化 する (所有者 検証 込み)。
// 依存 port: [repository.NotificationRepository]。
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

// MarkAllNotificationsReadUseCase は current user の 全 通知 を 一括 既読 化 する。
// 依存 port: [repository.NotificationRepository]。
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

// CountUnreadNotificationsUseCase は current user の 未読 通知 数 を 返す (バッジ 表示 用)。
// 依存 port: [repository.NotificationRepository]。
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
