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

// commentRepository は [repository.CommentRepository] の実装。comment_threads / comments は
// ナレッジ本体（knowledgeBaseRepository）と同じくスキーマの正本が schema.hcl で GORM を通さない
// 方針のため、クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
type commentRepository struct {
	baseRepository
}

// NewCommentRepository はコメントの repository を組み立てる。
func NewCommentRepository(db *sql.DB) repository.CommentRepository {
	return &commentRepository{baseRepository{db: db}}
}

// queries は ctx に乗っているトランザクション（あれば）に束縛した sqlc の Queries を作る。
// CreateCommentThreadUseCase が txManager.DoInTx の中で CreateCommentThread →
// CreateComment を呼ぶとき、両方がここ経由で同じトランザクションへ乗る。
func (r *commentRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

func toDomainCommentThread(row sqlcgen.CommentThread) domain.CommentThread {
	t := domain.CommentThread{
		ID:              row.ID.String(),
		WorkspaceID:     row.WorkspaceID.String(),
		PageID:          row.PageID.String(),
		CreatedByUserID: uint64(row.CreatedByUserID),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.BlockID.Valid {
		id := row.BlockID.UUID.String()
		t.BlockID = &id
	}
	if row.AnchorFrom.Valid {
		v := int(row.AnchorFrom.Int32)
		t.AnchorFrom = &v
	}
	if row.AnchorTo.Valid {
		v := int(row.AnchorTo.Int32)
		t.AnchorTo = &v
	}
	if row.Quote.Valid {
		t.Quote = &row.Quote.String
	}
	if row.ResolvedAt.Valid {
		v := row.ResolvedAt.Time
		t.ResolvedAt = &v
	}
	if row.ResolvedByUserID.Valid {
		v := uint64(row.ResolvedByUserID.Int64)
		t.ResolvedByUserID = &v
	}
	return t
}

func toDomainComment(row sqlcgen.Comment) domain.Comment {
	return domain.Comment{
		ID:           row.ID.String(),
		ThreadID:     row.ThreadID.String(),
		AuthorUserID: uint64(row.AuthorUserID),
		Body:         string(row.Body),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func (r *commentRepository) CreateCommentThread(
	ctx context.Context, workspaceID, pageID string, createdByUserID uint64,
) (*domain.CommentThread, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, repository.ErrCommentThreadNotFound
	}
	createdBy, ok3 := toInt64ID(createdByUserID)
	if !ok3 {
		return nil, outOfRangeIDError("created_by_user_id", createdByUserID)
	}
	id, err := kbNewID()
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).CreateCommentThread(ctx, sqlcgen.CreateCommentThreadParams{
		ID:              id,
		WorkspaceID:     wsID,
		PageID:          pgID,
		CreatedByUserID: createdBy,
	})
	if err != nil {
		return nil, err
	}
	t := toDomainCommentThread(row)
	return &t, nil
}

func (r *commentRepository) CreateComment(
	ctx context.Context, threadID string, authorUserID uint64, body string,
) (*domain.Comment, error) {
	thID, ok := kbParseID(threadID)
	if !ok {
		return nil, repository.ErrCommentThreadNotFound
	}
	authorID, ok2 := toInt64ID(authorUserID)
	if !ok2 {
		return nil, outOfRangeIDError("author_user_id", authorUserID)
	}
	id, err := kbNewID()
	if err != nil {
		return nil, err
	}
	row, err := r.queries(ctx).CreateComment(ctx, sqlcgen.CreateCommentParams{
		ID:           id,
		ThreadID:     thID,
		AuthorUserID: authorID,
		Body:         json.RawMessage(body),
	})
	if err != nil {
		return nil, err
	}
	c := toDomainComment(row)
	return &c, nil
}

func (r *commentRepository) GetCommentThread(ctx context.Context, workspaceID, pageID, threadID string) (*domain.CommentThread, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	thID, ok3 := kbParseID(threadID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrCommentThreadNotFound
	}
	row, err := r.queries(ctx).GetCommentThread(ctx, sqlcgen.GetCommentThreadParams{
		WorkspaceID: wsID, PageID: pgID, ID: thID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrCommentThreadNotFound
	}
	if err != nil {
		return nil, err
	}
	t := toDomainCommentThread(row)
	return &t, nil
}

func (r *commentRepository) ListCommentThreadsByPage(ctx context.Context, workspaceID, pageID string) ([]domain.CommentThread, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, repository.ErrCommentThreadNotFound
	}
	rows, err := r.queries(ctx).ListCommentThreadsByPage(ctx, sqlcgen.ListCommentThreadsByPageParams{
		WorkspaceID: wsID, PageID: pgID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.CommentThread, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainCommentThread(row))
	}
	return out, nil
}

func (r *commentRepository) ListCommentsByThreads(ctx context.Context, threadIDs []string) ([]domain.Comment, error) {
	if len(threadIDs) == 0 {
		return []domain.Comment{}, nil
	}
	idsJSON, err := json.Marshal(threadIDs)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries(ctx).ListCommentsByThreadIDs(ctx, idsJSON)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Comment, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainComment(row))
	}
	return out, nil
}

func (r *commentRepository) ResolveCommentThread(
	ctx context.Context, workspaceID, pageID, threadID string, resolvedByUserID uint64,
) (*domain.CommentThread, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	thID, ok3 := kbParseID(threadID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrCommentThreadNotFound
	}
	resolvedBy, ok4 := toInt64ID(resolvedByUserID)
	if !ok4 {
		return nil, outOfRangeIDError("resolved_by_user_id", resolvedByUserID)
	}
	row, err := r.queries(ctx).ResolveCommentThread(ctx, sqlcgen.ResolveCommentThreadParams{
		ResolvedByUserID: resolvedBy, WorkspaceID: wsID, PageID: pgID, ID: thID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrCommentThreadNotFound
	}
	if err != nil {
		return nil, err
	}
	t := toDomainCommentThread(row)
	return &t, nil
}

func (r *commentRepository) ReopenCommentThread(ctx context.Context, workspaceID, pageID, threadID string) (*domain.CommentThread, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	thID, ok3 := kbParseID(threadID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrCommentThreadNotFound
	}
	row, err := r.queries(ctx).ReopenCommentThread(ctx, sqlcgen.ReopenCommentThreadParams{
		WorkspaceID: wsID, PageID: pgID, ID: thID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrCommentThreadNotFound
	}
	if err != nil {
		return nil, err
	}
	t := toDomainCommentThread(row)
	return &t, nil
}
