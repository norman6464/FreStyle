package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ticketCommentRepository は [repository.TicketCommentRepository] の実装（段 3）。
// ticketRepository とは別の struct にする（commentRepository と同じ判断 — 表の群が別なら
// repository も別にする）。
type ticketCommentRepository struct {
	baseRepository
}

// NewTicketCommentRepository はチケットの発言・編集履歴・反応の repository を組み立てる。
func NewTicketCommentRepository(db *sql.DB) repository.TicketCommentRepository {
	return &ticketCommentRepository{baseRepository{db: db}}
}

func (r *ticketCommentRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

func toDomainTicketComment(row sqlcgen.TicketComment) domain.TicketComment {
	c := domain.TicketComment{
		ID:           row.ID.String(),
		WorkspaceID:  row.WorkspaceID.String(),
		TicketID:     row.TicketID.String(),
		AuthorUserID: uint64(row.AuthorUserID),
		Body:         string(row.Body),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.ParentCommentID.Valid {
		id := row.ParentCommentID.UUID.String()
		c.ParentCommentID = &id
	}
	if row.EditedAt.Valid {
		v := row.EditedAt.Time
		c.EditedAt = &v
	}
	if row.DeletedAt.Valid {
		v := row.DeletedAt.Time
		c.DeletedAt = &v
	}
	return c
}

func toDomainTicketCommentEdit(row sqlcgen.TicketCommentEdit) domain.TicketCommentEdit {
	return domain.TicketCommentEdit{
		ID:           row.ID.String(),
		CommentID:    row.CommentID.String(),
		EditorUserID: uint64(row.EditorUserID),
		PreviousBody: string(row.PreviousBody),
		EditedAt:     row.EditedAt,
	}
}

func toDomainTicketCommentReaction(row sqlcgen.TicketCommentReaction) domain.TicketCommentReaction {
	return domain.TicketCommentReaction{
		CommentID: row.CommentID.String(),
		UserID:    uint64(row.UserID),
		Emoji:     row.Emoji,
		CreatedAt: row.CreatedAt,
	}
}

func (r *ticketCommentRepository) CreateTicketComment(ctx context.Context, c *domain.TicketComment) error {
	wsID, ok := kbParseID(c.WorkspaceID)
	tID, ok2 := kbParseID(c.TicketID)
	if !ok || !ok2 {
		return repository.ErrTicketNotFound
	}
	parentID, ok3 := kbNullID(c.ParentCommentID)
	if !ok3 {
		return repository.ErrTicketCommentNotFound
	}
	authorID, ok4 := toInt64ID(c.AuthorUserID)
	if !ok4 {
		return outOfRangeIDError("author_user_id", c.AuthorUserID)
	}
	id, err := ticketNewID()
	if err != nil {
		return err
	}
	row, err := r.queries(ctx).CreateTicketComment(ctx, sqlcgen.CreateTicketCommentParams{
		ID: id, WorkspaceID: wsID, TicketID: tID, ParentCommentID: parentID,
		AuthorUserID: authorID, Body: json.RawMessage(c.Body),
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return repository.ErrTicketNotFound
		}
		return err
	}
	*c = toDomainTicketComment(row)
	return nil
}

func (r *ticketCommentRepository) FindTicketComment(ctx context.Context, workspaceID, ticketID, commentID string) (*domain.TicketComment, error) {
	wsID, ok := kbParseID(workspaceID)
	tID, ok2 := kbParseID(ticketID)
	cID, ok3 := kbParseID(commentID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrTicketCommentNotFound
	}
	row, err := r.queries(ctx).GetTicketComment(ctx, sqlcgen.GetTicketCommentParams{
		WorkspaceID: wsID, TicketID: tID, ID: cID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrTicketCommentNotFound
	}
	if err != nil {
		return nil, err
	}
	c := toDomainTicketComment(row)
	return &c, nil
}

func (r *ticketCommentRepository) ListTicketComments(ctx context.Context, workspaceID, ticketID string) ([]domain.TicketComment, error) {
	wsID, ok := kbParseID(workspaceID)
	tID, ok2 := kbParseID(ticketID)
	if !ok || !ok2 {
		return nil, nil
	}
	rows, err := r.queries(ctx).ListTicketComments(ctx, sqlcgen.ListTicketCommentsParams{WorkspaceID: wsID, TicketID: tID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TicketComment, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainTicketComment(row))
	}
	return out, nil
}

func (r *ticketCommentRepository) UpdateTicketCommentBody(ctx context.Context, workspaceID, ticketID, commentID, body string) (*domain.TicketComment, error) {
	wsID, ok := kbParseID(workspaceID)
	tID, ok2 := kbParseID(ticketID)
	cID, ok3 := kbParseID(commentID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrTicketCommentNotFound
	}
	row, err := r.queries(ctx).UpdateTicketCommentBody(ctx, sqlcgen.UpdateTicketCommentBodyParams{
		WorkspaceID: wsID, TicketID: tID, ID: cID, Body: json.RawMessage(body),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrTicketCommentNotFound
	}
	if err != nil {
		return nil, err
	}
	c := toDomainTicketComment(row)
	return &c, nil
}

func (r *ticketCommentRepository) DeleteTicketComment(ctx context.Context, workspaceID, ticketID, commentID string) error {
	wsID, ok := kbParseID(workspaceID)
	tID, ok2 := kbParseID(ticketID)
	cID, ok3 := kbParseID(commentID)
	if !ok || !ok2 || !ok3 {
		return repository.ErrTicketCommentNotFound
	}
	n, err := r.queries(ctx).SoftDeleteTicketComment(ctx, sqlcgen.SoftDeleteTicketCommentParams{
		WorkspaceID: wsID, TicketID: tID, ID: cID,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrTicketCommentNotFound
	}
	return nil
}

func (r *ticketCommentRepository) InsertTicketCommentEdit(ctx context.Context, e *domain.TicketCommentEdit) error {
	wsID, ok := kbParseID(e.WorkspaceID)
	cID, ok2 := kbParseID(e.CommentID)
	if !ok || !ok2 {
		return repository.ErrTicketCommentNotFound
	}
	editorID, ok3 := toInt64ID(e.EditorUserID)
	if !ok3 {
		return outOfRangeIDError("editor_user_id", e.EditorUserID)
	}
	id, err := ticketNewID()
	if err != nil {
		return err
	}
	if err := r.queries(ctx).InsertTicketCommentEdit(ctx, sqlcgen.InsertTicketCommentEditParams{
		ID: id, WorkspaceID: wsID, CommentID: cID, EditorUserID: editorID, PreviousBody: json.RawMessage(e.PreviousBody),
	}); err != nil {
		if isForeignKeyViolation(err) {
			return repository.ErrTicketCommentNotFound
		}
		return err
	}
	return nil
}

func (r *ticketCommentRepository) ListTicketCommentEdits(ctx context.Context, workspaceID, commentID string) ([]domain.TicketCommentEdit, error) {
	wsID, ok := kbParseID(workspaceID)
	cID, ok2 := kbParseID(commentID)
	if !ok || !ok2 {
		return nil, nil
	}
	rows, err := r.queries(ctx).ListTicketCommentEdits(ctx, sqlcgen.ListTicketCommentEditsParams{WorkspaceID: wsID, CommentID: cID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TicketCommentEdit, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainTicketCommentEdit(row))
	}
	return out, nil
}

func (r *ticketCommentRepository) AddTicketCommentReaction(ctx context.Context, workspaceID, commentID string, userID uint64, emoji string) error {
	wsID, ok := kbParseID(workspaceID)
	cID, ok2 := kbParseID(commentID)
	if !ok || !ok2 {
		return repository.ErrTicketCommentNotFound
	}
	uID, ok3 := toInt64ID(userID)
	if !ok3 {
		return outOfRangeIDError("user_id", userID)
	}
	_, err := r.queries(ctx).InsertTicketCommentReaction(ctx, sqlcgen.InsertTicketCommentReactionParams{
		WorkspaceID: wsID, CommentID: cID, UserID: uID, Emoji: emoji,
	})
	if err != nil {
		if isForeignKeyViolation(err) {
			return repository.ErrTicketCommentNotFound
		}
		return err
	}
	return nil
}

func (r *ticketCommentRepository) RemoveTicketCommentReaction(ctx context.Context, workspaceID, commentID string, userID uint64, emoji string) error {
	wsID, ok := kbParseID(workspaceID)
	cID, ok2 := kbParseID(commentID)
	if !ok || !ok2 {
		return repository.ErrTicketCommentNotFound
	}
	uID, ok3 := toInt64ID(userID)
	if !ok3 {
		return outOfRangeIDError("user_id", userID)
	}
	// 付いていない反応を外そうとしても 0 行で成功扱い（冪等）。
	_, err := r.queries(ctx).DeleteTicketCommentReaction(ctx, sqlcgen.DeleteTicketCommentReactionParams{
		WorkspaceID: wsID, CommentID: cID, UserID: uID, Emoji: emoji,
	})
	return err
}

func (r *ticketCommentRepository) ListTicketCommentReactions(ctx context.Context, workspaceID string, commentIDs []string) ([]domain.TicketCommentReaction, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok || len(commentIDs) == 0 {
		return nil, nil
	}
	idsJSON, err := json.Marshal(commentIDs)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries(ctx).ListTicketCommentReactionsByComments(ctx, sqlcgen.ListTicketCommentReactionsByCommentsParams{
		WorkspaceID: wsID, CommentIds: idsJSON,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TicketCommentReaction, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainTicketCommentReaction(row))
	}
	return out, nil
}
