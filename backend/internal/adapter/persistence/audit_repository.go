package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// auditRepository は [repository.AuditRepository] の実装。
// 読み取り（ListRecent）は sqlc 生成コード、書き込み（Record）は GORM。
type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) repository.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Record(ctx context.Context, e *domain.AuditEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func toDomainAuditEvent(row sqlcgen.AuditEvent) domain.AuditEvent {
	return domain.AuditEvent{
		ID:         uint64(row.ID),
		ActorID:    uint64(row.ActorID),
		ActorEmail: row.ActorEmail,
		ActorRole:  row.ActorRole,
		Action:     row.Action,
		TargetID:   uint64(row.TargetID),
		CreatedAt:  row.CreatedAt,
	}
}

func (r *auditRepository) ListRecent(ctx context.Context, limit int) ([]domain.AuditEvent, error) {
	if limit <= 0 {
		limit = 200
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListRecentAuditEvents(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	out := make([]domain.AuditEvent, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainAuditEvent(row))
	}
	return out, nil
}
