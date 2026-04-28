package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Notification, error)
	// MarkRead は所有者検証込みで is_read を立てる。
	// 自分以外の通知を既読化できないように、必ず WHERE で user_id を絞る。
	MarkRead(ctx context.Context, userID, id uint64) error
}

type notificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository { return &notificationRepository{db: db} }

func (r *notificationRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.Notification, error) {
	var rows []domain.Notification
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *notificationRepository) MarkRead(ctx context.Context, userID, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

// SnsPublisher は通知 push 用の interface（実装は AWS SDK 連携で別 PR）。
type SnsPublisher interface {
	Publish(ctx context.Context, userID uint64, title, body string) error
}

type stubSnsPublisher struct{}

func NewStubSnsPublisher() SnsPublisher { return &stubSnsPublisher{} }

func (p *stubSnsPublisher) Publish(_ context.Context, _ uint64, _, _ string) error { return nil }
