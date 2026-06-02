package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// notificationRepository は [repository.NotificationRepository] の GORM 実装。
type notificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

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

func (r *notificationRepository) MarkAllRead(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uint64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&n).Error
	return n, err
}

// stubSnsPublisher は [repository.SnsPublisher] の no-op 実装（本番の SNS 実装は別 PR）。
type stubSnsPublisher struct{}

func NewStubSnsPublisher() repository.SnsPublisher { return &stubSnsPublisher{} }

func (p *stubSnsPublisher) Publish(_ context.Context, _ uint64, _, _ string) error { return nil }
