package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// ticketAttachmentRepository は [repository.TicketAttachmentRepository] の実装（段 4）。
type ticketAttachmentRepository struct {
	baseRepository
}

// NewTicketAttachmentRepository はチケット添付メタデータの repository を組み立てる。
func NewTicketAttachmentRepository(db *sql.DB) repository.TicketAttachmentRepository {
	return &ticketAttachmentRepository{baseRepository{db: db}}
}

func (r *ticketAttachmentRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

func toDomainTicketAttachment(row sqlcgen.TicketAttachment) domain.TicketAttachment {
	return domain.TicketAttachment{
		ID:               row.ID.String(),
		WorkspaceID:      row.WorkspaceID.String(),
		TicketID:         row.TicketID.String(),
		Key:              row.Key,
		Filename:         row.Filename,
		ContentType:      row.ContentType,
		SizeBytes:        row.SizeBytes,
		UploadedByUserID: uint64(row.UploadedByUserID),
		CreatedAt:        row.CreatedAt,
	}
}

func (r *ticketAttachmentRepository) CreateTicketAttachment(ctx context.Context, a *domain.TicketAttachment) error {
	wsID, ok := kbParseID(a.WorkspaceID)
	tID, ok2 := kbParseID(a.TicketID)
	if !ok || !ok2 {
		return repository.ErrTicketNotFound
	}
	uploaderID, ok3 := toInt64ID(a.UploadedByUserID)
	if !ok3 {
		return outOfRangeIDError("uploaded_by_user_id", a.UploadedByUserID)
	}
	id, err := ticketNewID()
	if err != nil {
		return err
	}
	row, err := r.queries(ctx).CreateTicketAttachment(ctx, sqlcgen.CreateTicketAttachmentParams{
		ID: id, WorkspaceID: wsID, TicketID: tID, Key: a.Key, Filename: a.Filename,
		ContentType: a.ContentType, SizeBytes: a.SizeBytes, UploadedByUserID: uploaderID,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return repository.ErrTicketNotFound
		}
		return err
	}
	*a = toDomainTicketAttachment(row)
	return nil
}

func (r *ticketAttachmentRepository) FindTicketAttachment(ctx context.Context, workspaceID, ticketID, attachmentID string) (*domain.TicketAttachment, error) {
	wsID, ok := kbParseID(workspaceID)
	tID, ok2 := kbParseID(ticketID)
	aID, ok3 := kbParseID(attachmentID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrTicketAttachmentNotFound
	}
	row, err := r.queries(ctx).FindTicketAttachment(ctx, sqlcgen.FindTicketAttachmentParams{
		WorkspaceID: wsID, TicketID: tID, ID: aID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrTicketAttachmentNotFound
	}
	if err != nil {
		return nil, err
	}
	a := toDomainTicketAttachment(row)
	return &a, nil
}

func (r *ticketAttachmentRepository) ListTicketAttachments(ctx context.Context, workspaceID, ticketID string) ([]domain.TicketAttachment, error) {
	wsID, ok := kbParseID(workspaceID)
	tID, ok2 := kbParseID(ticketID)
	if !ok || !ok2 {
		return nil, nil
	}
	rows, err := r.queries(ctx).ListTicketAttachments(ctx, sqlcgen.ListTicketAttachmentsParams{WorkspaceID: wsID, TicketID: tID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TicketAttachment, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainTicketAttachment(row))
	}
	return out, nil
}

func (r *ticketAttachmentRepository) DeleteTicketAttachment(ctx context.Context, workspaceID, ticketID, attachmentID string) error {
	wsID, ok := kbParseID(workspaceID)
	tID, ok2 := kbParseID(ticketID)
	aID, ok3 := kbParseID(attachmentID)
	if !ok || !ok2 || !ok3 {
		return repository.ErrTicketAttachmentNotFound
	}
	n, err := r.queries(ctx).DeleteTicketAttachment(ctx, sqlcgen.DeleteTicketAttachmentParams{
		WorkspaceID: wsID, TicketID: tID, ID: aID,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrTicketAttachmentNotFound
	}
	return nil
}
