package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// notificationRepository は [repository.NotificationRepository] の実装。
// 読み取り（ListByUserID / CountUnread）は sqlc 生成コード（生 SQL）、書き込みは GORM。
type notificationRepository struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.Notification, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return []domain.Notification{}, nil // 存在し得ない user_id = 0 件
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListNotificationsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Notification{
			ID:        uint64(row.ID),
			UserID:    uint64(row.UserID),
			Type:      row.Type,
			Title:     row.Title,
			Body:      row.Body,
			IsRead:    row.IsRead,
			CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
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
	uid, ok := toInt64ID(userID)
	if !ok {
		return 0, nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return 0, err
	}
	return sqlcgen.New(sqlDB).CountUnreadNotifications(ctx, uid)
}

// stubSnsPublisher は [repository.SnsPublisher] の no-op 実装（本番の SNS 実装は別 PR）。
type stubSnsPublisher struct{}

func NewStubSnsPublisher() repository.SnsPublisher { return &stubSnsPublisher{} }

func (p *stubSnsPublisher) Publish(_ context.Context, _ uint64, _, _ string) error { return nil }
