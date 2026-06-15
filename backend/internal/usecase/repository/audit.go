package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// AuditRepository は監査イベント（audit_events）の記録と一覧取得を提供する。
type AuditRepository interface {
	// Record は監査イベントを 1 件保存する。
	Record(ctx context.Context, e *domain.AuditEvent) error
	// ListRecent は新しい順で最大 limit 件の監査イベントを返す。
	ListRecent(ctx context.Context, limit int) ([]domain.AuditEvent, error)
}
