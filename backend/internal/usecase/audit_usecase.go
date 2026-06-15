package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// RecordAuditEventInput は監査イベント 1 件の記録入力。
type RecordAuditEventInput struct {
	ActorID    uint64
	ActorEmail string
	ActorRole  string
	Action     string
	TargetID   uint64
}

// RecordAuditEventUseCase は管理者の重要操作を監査ログに記録する。
type RecordAuditEventUseCase struct {
	repo repository.AuditRepository
}

func NewRecordAuditEventUseCase(r repository.AuditRepository) *RecordAuditEventUseCase {
	return &RecordAuditEventUseCase{repo: r}
}

func (u *RecordAuditEventUseCase) Execute(ctx context.Context, in RecordAuditEventInput) error {
	return u.repo.Record(ctx, &domain.AuditEvent{
		ActorID:    in.ActorID,
		ActorEmail: in.ActorEmail,
		ActorRole:  in.ActorRole,
		Action:     in.Action,
		TargetID:   in.TargetID,
	})
}

// ListAuditEventsUseCase は監査ログを新しい順で返す（super_admin 専用画面用）。
type ListAuditEventsUseCase struct {
	repo repository.AuditRepository
}

func NewListAuditEventsUseCase(r repository.AuditRepository) *ListAuditEventsUseCase {
	return &ListAuditEventsUseCase{repo: r}
}

func (u *ListAuditEventsUseCase) Execute(ctx context.Context) ([]domain.AuditEvent, error) {
	return u.repo.ListRecent(ctx, 200)
}
