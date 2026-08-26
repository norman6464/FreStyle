package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// auditRepository は [repository.AuditRepository] の実装。
// 読み書きとも sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type auditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) repository.AuditRepository {
	return &auditRepository{db: db}
}

// Record は監査ログを 1 件記録し、採番 id と created_at を引数の構造体へ書き戻す。
func (r *auditRepository) Record(ctx context.Context, e *domain.AuditEvent) error {
	actorID, ok := toInt64ID(e.ActorID)
	if !ok {
		// 何も書かずに nil を返すと「記録できた」と誤認される。書けないことをエラーで伝える。
		return fmt.Errorf("actor id %d が int64 の範囲外です", e.ActorID)
	}
	targetID, ok := toInt64ID(e.TargetID)
	if !ok {
		return fmt.Errorf("target id %d が int64 の範囲外です", e.TargetID)
	}
	createdAt := e.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now() // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(r.db).InsertAuditEvent(ctx, sqlcgen.InsertAuditEventParams{
		ActorID:    actorID,
		ActorEmail: e.ActorEmail,
		ActorRole:  e.ActorRole,
		Action:     e.Action,
		TargetID:   targetID,
		CreatedAt:  createdAt,
	})
	if err != nil {
		return err
	}
	e.ID = uint64(row.ID)
	e.CreatedAt = row.CreatedAt
	return nil
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
	rows, err := sqlcgen.New(r.db).ListRecentAuditEvents(ctx, int64(limit))
	if err != nil {
		return nil, err
	}
	out := make([]domain.AuditEvent, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainAuditEvent(row))
	}
	return out, nil
}
