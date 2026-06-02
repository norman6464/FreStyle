package repository

import "context"

// HealthRepository は DB の到達性を確認する。
type HealthRepository interface {
	PingDB(ctx context.Context) error
}
