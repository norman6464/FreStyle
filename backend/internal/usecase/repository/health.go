package repository

import "context"

// HealthRepository は DB の到達性を確認する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// healthRepository (GORM の sql.DB を Ping)。
type HealthRepository interface {
	PingDB(ctx context.Context) error
}
