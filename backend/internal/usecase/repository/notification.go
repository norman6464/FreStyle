package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NotificationRepository は notifications テーブルへのアクセスを提供する。
type NotificationRepository interface {
	// Create は通知を 1 件作成する（システムが super_admin 等に通知を出す用途）。
	Create(ctx context.Context, n *domain.Notification) error
	// CreateMany は通知をまとめて作成する。宛先が増えるほど往復が比例して増えるのを避ける。
	// 空スライスは何もせず nil を返す（呼び出し側で件数を気にしなくてよいように）。
	CreateMany(ctx context.Context, ns []domain.Notification) error
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Notification, error)
	// MarkRead は WHERE で user_id を絞り、自分以外の通知を既読化できないようにする。
	MarkRead(ctx context.Context, userID, id uint64) error
	MarkAllRead(ctx context.Context, userID uint64) error
	CountUnread(ctx context.Context, userID uint64) (int64, error)
}

// SnsPublisher は通知 push 用（実装は AWS SDK 連携で別 PR）。
type SnsPublisher interface {
	Publish(ctx context.Context, userID uint64, title, body string) error
}
