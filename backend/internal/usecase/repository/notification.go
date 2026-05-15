package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NotificationRepository は notifications テーブル へ の アクセス を 提供 する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// notificationRepository (GORM)。
type NotificationRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Notification, error)
	// MarkRead は所有者検証込みで is_read を立てる。
	// 自分以外の通知を既読化できないように、必ず WHERE で user_id を絞る。
	MarkRead(ctx context.Context, userID, id uint64) error
	// MarkAllRead は current user の全通知を既読化する。
	MarkAllRead(ctx context.Context, userID uint64) error
	// CountUnread は current user の未読通知数を返す。
	CountUnread(ctx context.Context, userID uint64) (int64, error)
}

// SnsPublisher は通知 push 用の interface（実装は AWS SDK 連携で別 PR）。
//
// 単一 メソッド の port な ので Effective Go 流 の -er 命名 を 採用。 stub 実装 は
// [github.com/norman6464/FreStyle/backend/internal/adapter/persistence].NewStubSnsPublisher。
type SnsPublisher interface {
	Publish(ctx context.Context, userID uint64, title, body string) error
}
