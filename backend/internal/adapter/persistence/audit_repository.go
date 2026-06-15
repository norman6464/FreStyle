package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// auditRepository は [repository.AuditRepository] の GORM 実装。
type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) repository.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Record(ctx context.Context, e *domain.AuditEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *auditRepository) ListRecent(ctx context.Context, limit int) ([]domain.AuditEvent, error) {
	if limit <= 0 {
		limit = 200
	}
	var rows []domain.AuditEvent
	err := r.db.WithContext(ctx).
		// created_at が同一秒でも順序が安定するよう id を tiebreaker にする。
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
