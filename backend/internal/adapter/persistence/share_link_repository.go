package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// shareLinkRepository は [repository.ShareLinkRepository] の実装。
type shareLinkRepository struct {
	baseRepository
}

// NewShareLinkRepository は共有リンクの repository を組み立てる。
func NewShareLinkRepository(db *sql.DB) repository.ShareLinkRepository {
	return &shareLinkRepository{baseRepository{db: db}}
}

// queries は ctx に乗っているトランザクション（あれば）に束縛した sqlc の Queries を作る。
func (r *shareLinkRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

// runInTx は 1 つのトランザクションを開き、その中でだけ有効な Queries を fn に渡す。
// ctx に既に外側の DoInTx が開いたトランザクションがあれば、新規に開始せずそれへ相乗りする
// （二重に BeginTx するとデッドロックの原因になる。commit/rollback は外側だけが持つ）。
func (r *shareLinkRepository) runInTx(ctx context.Context, fn func(qtx *sqlcgen.Queries) error) error {
	if tx, ok := getTx(ctx); ok {
		return fn(sqlcgen.New(tx))
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	if err := fn(sqlcgen.New(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func toDomainShareLink(row sqlcgen.ShareLink) domain.ShareLink {
	l := domain.ShareLink{
		ID:              row.ID.String(),
		WorkspaceID:     row.WorkspaceID.String(),
		PageID:          row.PageID.String(),
		PrincipalID:     row.PrincipalID.String(),
		Capability:      domain.Capability(row.Capability),
		TokenHash:       row.TokenHash,
		CreatedByUserID: uint64(row.CreatedByUserID),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.PasswordHash.Valid {
		h := row.PasswordHash.String
		l.PasswordHash = &h
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		l.ExpiresAt = &t
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		l.RevokedAt = &t
	}
	return l
}

// nullString は *string を sql.NullString へ変換する。
func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// nullTime は *time.Time を sql.NullTime へ変換する。
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func (r *shareLinkRepository) Create(ctx context.Context, in repository.ShareLinkWrite) (*domain.ShareLink, error) {
	wsID, ok := kbParseID(in.WorkspaceID)
	pgID, ok2 := kbParseID(in.PageID)
	if !ok || !ok2 {
		return nil, repository.ErrPageNotFound
	}
	// share_links.created_by_user_id は bigint で、users への FK も張っている。
	// 範囲外の発行者 ID では 1 行も書けないので、書き込みに入る前にエラーで止める
	// （nil を返すと呼び出し側がリンクを発行できたと誤認する）。
	createdBy, cok := toInt64ID(in.CreatedByUserID)
	if !cok {
		return nil, outOfRangeIDError("created_by_user_id", in.CreatedByUserID)
	}
	principalID, err := kbNewID()
	if err != nil {
		return nil, err
	}
	linkID, err := kbNewID()
	if err != nil {
		return nil, err
	}

	// 主体とリンクは 1 対 1 で、どちらかだけが残ると失効も解決もできない行になる。
	// 同じトランザクションで両方作る。
	var created sqlcgen.ShareLink
	err = r.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
		if _, err := qtx.InsertPrincipal(ctx, sqlcgen.InsertPrincipalParams{
			ID:          principalID,
			WorkspaceID: wsID,
			Kind:        string(domain.PrincipalKindShareLink),
			PageID:      uuid.NullUUID{UUID: pgID, Valid: true},
		}); err != nil {
			return err
		}
		row, err := qtx.InsertShareLink(ctx, sqlcgen.InsertShareLinkParams{
			ID:              linkID,
			WorkspaceID:     wsID,
			PageID:          pgID,
			PrincipalID:     principalID,
			Capability:      string(in.Capability),
			TokenHash:       in.TokenHash,
			PasswordHash:    nullString(in.PasswordHash),
			ExpiresAt:       nullTime(in.ExpiresAt),
			CreatedByUserID: createdBy,
		})
		if err != nil {
			return err
		}
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	link := toDomainShareLink(created)
	return &link, nil
}

func (r *shareLinkRepository) Revoke(ctx context.Context, workspaceID, shareLinkID string) error {
	wsID, ok := kbParseID(workspaceID)
	lnID, ok2 := kbParseID(shareLinkID)
	if !ok || !ok2 {
		return repository.ErrShareLinkNotFound
	}
	n, err := r.queries(ctx).RevokeShareLink(ctx, sqlcgen.RevokeShareLinkParams{WorkspaceID: wsID, ID: lnID})
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	// 0 件は「既に失効済み」と「そもそも無い」の両方があり得る。区別して返す
	// （失効の再実行は冪等に成功させ、存在しない ID は not found にする）。
	if _, err := r.queries(ctx).GetShareLink(ctx, sqlcgen.GetShareLinkParams{WorkspaceID: wsID, ID: lnID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrShareLinkNotFound
		}
		return err
	}
	return nil
}

func (r *shareLinkRepository) FindByTokenHash(ctx context.Context, tokenHash []byte) (*domain.ShareLink, error) {
	row, err := r.queries(ctx).GetShareLinkByTokenHash(ctx, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrShareLinkNotFound
	}
	if err != nil {
		return nil, err
	}
	link := toDomainShareLink(row)
	return &link, nil
}

func (r *shareLinkRepository) ListByPage(ctx context.Context, workspaceID, pageID string) ([]domain.ShareLink, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []domain.ShareLink{}, nil
	}
	rows, err := r.queries(ctx).ListPageShareLinks(ctx, sqlcgen.ListPageShareLinksParams{WorkspaceID: wsID, PageID: pgID})
	if err != nil {
		return nil, err
	}
	links := make([]domain.ShareLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, toDomainShareLink(row))
	}
	return links, nil
}
