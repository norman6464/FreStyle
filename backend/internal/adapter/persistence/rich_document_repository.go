package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// mapPgDataError は PostgreSQL の data exception（SQLSTATE class 22。jsonb/text に格納できない
// U+0000 や不正サロゲート等）を repository.ErrRichDocumentInvalidData へ翻訳する。
// これにより「クライアント起因の不正データ」が 500 ではなく 400 として返る。
func mapPgDataError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, "22") {
		return repository.ErrRichDocumentInvalidData
	}
	return err
}

// richDocumentRepository は [repository.RichDocumentRepository] の実装。
// doc は jsonb 列に保存し、更新は revision 一致を条件にした楽観ロックで行う。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type richDocumentRepository struct{ db *sql.DB }

// NewRichDocumentRepository は rich_documents の repository を組み立てる。
func NewRichDocumentRepository(db *sql.DB) repository.RichDocumentRepository {
	return &richDocumentRepository{db: db}
}

// richDocumentRow は行全体（doc / deleted_at 含む）を返すクエリの共通の行形。GetRichDocumentByID /
// UpdateRichDocumentWithRevision は rich_documents の全列を返すため、sqlc は個別の Row 型を
// 生成せずテーブル型（sqlcgen.RichDocument）をそのまま返す。
type richDocumentRow = sqlcgen.RichDocument

// toDomainRichDocument は行全体（doc / deleted_at 含む）を domain へ写す。
func toDomainRichDocument(row richDocumentRow) domain.RichDocument {
	d := domain.RichDocument{
		ID:            row.ID.String(),
		OwnerID:       uint64(row.OwnerID),
		Kind:          domain.DocumentKind(row.Kind),
		Title:         row.Title,
		IsPublic:      row.IsPublic,
		SchemaVersion: int(row.SchemaVersion),
		Doc:           string(row.Doc),
		Revision:      int(row.Revision),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	if row.CompanyID.Valid {
		c := uint64(row.CompanyID.Int64)
		d.CompanyID = &c
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		d.DeletedAt = &t
	}
	if row.WorkspaceID.Valid {
		wid := row.WorkspaceID.UUID.String()
		d.WorkspaceID = &wid
	}
	return d
}

// toDomainRichDocumentSummary は一覧用の軽量行（doc / deleted_at を含まない）を domain へ写す。
func toDomainRichDocumentSummary(row sqlcgen.ListRichDocumentsByOwnerRow) domain.RichDocument {
	d := domain.RichDocument{
		ID:            row.ID.String(),
		OwnerID:       uint64(row.OwnerID),
		Kind:          domain.DocumentKind(row.Kind),
		Title:         row.Title,
		IsPublic:      row.IsPublic,
		SchemaVersion: int(row.SchemaVersion),
		Revision:      int(row.Revision),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	if row.CompanyID.Valid {
		c := uint64(row.CompanyID.Int64)
		d.CompanyID = &c
	}
	if row.WorkspaceID.Valid {
		wid := row.WorkspaceID.UUID.String()
		d.WorkspaceID = &wid
	}
	return d
}

// nullCompanyID は *uint64 の会社 ID を sql.NullInt64 へ変換する（NULL 可）。
func nullCompanyID(companyID *uint64) (sql.NullInt64, bool) {
	if companyID == nil {
		return sql.NullInt64{}, true
	}
	cid, ok := toInt64ID(*companyID)
	if !ok {
		return sql.NullInt64{}, false
	}
	return sql.NullInt64{Int64: cid, Valid: true}, true
}

func (r *richDocumentRepository) Create(ctx context.Context, doc *domain.RichDocument) error {
	if doc.ID == "" {
		// UUIDv7 を採番する。時系列で単調に増える（インデックス局所性が良く作成順ソート可能）うえ、
		// ランダム部 74bit により URL は推測困難のまま。v4 で作られた既存 ID とも同形式で互換。
		// 失敗は乱数源の故障（v4 でも同様に失敗する）なので、退避せずエラーで返す。
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("uuid v7 の採番に失敗: %w", err)
		}
		doc.ID = id.String()
	}
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return fmt.Errorf("rich document id が uuid ではない: %w", err)
	}
	ownerID, ok := toInt64ID(doc.OwnerID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("owner_id", doc.OwnerID)
	}
	companyID, ok := nullCompanyID(doc.CompanyID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		// nullCompanyID は nil のとき必ず ok=true なので、ここでは非 nil が保証される。
		return outOfRangeIDError("company_id", *doc.CompanyID)
	}
	workspaceID, ok := nullWorkspaceID(doc.WorkspaceID)
	if !ok {
		return fmt.Errorf("workspace_id が不正な形式です: %q", *doc.WorkspaceID)
	}
	now := time.Now()
	createdAt := doc.CreatedAt
	if createdAt.IsZero() {
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := doc.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(r.db).InsertRichDocument(ctx, sqlcgen.InsertRichDocumentParams{
		ID:            id,
		OwnerID:       ownerID,
		CompanyID:     companyID,
		WorkspaceID:   workspaceID,
		Kind:          string(doc.Kind),
		Title:         doc.Title,
		IsPublic:      doc.IsPublic,
		SchemaVersion: int64(doc.SchemaVersion), // 0 は SQL 側の COALESCE で既定 1 に倒す
		Doc:           json.RawMessage(doc.Doc),
		Revision:      int64(doc.Revision), // 0 は SQL 側の COALESCE で既定 1 に倒す
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	})
	if err != nil {
		return mapPgDataError(err)
	}
	doc.SchemaVersion = int(row.SchemaVersion) // 既定 1 が当たった場合を書き戻す
	doc.Revision = int(row.Revision)           // 既定 1 が当たった場合を書き戻す
	doc.CreatedAt = row.CreatedAt
	doc.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *richDocumentRepository) FindByID(ctx context.Context, id string) (*domain.RichDocument, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, repository.ErrRichDocumentNotFound // uuid でない ID は存在し得ない
	}
	row, err := sqlcgen.New(r.db).GetRichDocumentByID(ctx, uid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrRichDocumentNotFound
	}
	if err != nil {
		return nil, err
	}
	d := toDomainRichDocument(row)
	return &d, nil
}

func (r *richDocumentRepository) UpdateWithRevision(ctx context.Context, doc *domain.RichDocument, expectedRevision int) error {
	uid, err := uuid.Parse(doc.ID)
	if err != nil {
		return repository.ErrRichDocumentNotFound // uuid でない ID は存在し得ない
	}
	row, err := sqlcgen.New(r.db).UpdateRichDocumentWithRevision(ctx, sqlcgen.UpdateRichDocumentWithRevisionParams{
		Title:            doc.Title,
		IsPublic:         doc.IsPublic,
		SchemaVersion:    int64(doc.SchemaVersion),
		Doc:              json.RawMessage(doc.Doc),
		ID:               uid,
		ExpectedRevision: int64(expectedRevision),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 0 行 = 「存在しない/論理削除済み」か「版不一致」。存在確認で切り分ける。
			if _, ferr := r.FindByID(ctx, doc.ID); ferr != nil {
				return ferr // ErrRichDocumentNotFound
			}
			return repository.ErrRichDocumentConflict
		}
		return mapPgDataError(err)
	}
	// 更新後の正確な行（revision / updated_at など）を doc に反映する。
	*doc = toDomainRichDocument(row)
	return nil
}

func (r *richDocumentRepository) SoftDelete(ctx context.Context, id string, ownerID uint64) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return repository.ErrRichDocumentNotFound // uuid でない ID は存在し得ない
	}
	oid, ok := toInt64ID(ownerID)
	if !ok {
		return repository.ErrRichDocumentNotFound // 存在し得ない owner_id は対象なし
	}
	affected, err := sqlcgen.New(r.db).SoftDeleteRichDocument(ctx, sqlcgen.SoftDeleteRichDocumentParams{
		ID:      uid,
		OwnerID: oid,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return repository.ErrRichDocumentNotFound
	}
	return nil
}

func (r *richDocumentRepository) ListByOwner(ctx context.Context, ownerID uint64, kind domain.DocumentKind) ([]domain.RichDocument, error) {
	oid, ok := toInt64ID(ownerID)
	if !ok {
		return []domain.RichDocument{}, nil // 存在し得ない owner_id = 0 件
	}
	rows, err := sqlcgen.New(r.db).ListRichDocumentsByOwner(ctx, sqlcgen.ListRichDocumentsByOwnerParams{
		OwnerID: oid,
		Kind:    string(kind), // 空文字なら SQL 側で絞り込まない
	})
	if err != nil {
		return nil, err
	}
	// 0 件でも nil ではなく空スライスを返す（JSON が null にならずフロントの map/for-of が落ちない）。
	docs := make([]domain.RichDocument, 0, len(rows))
	for _, row := range rows {
		docs = append(docs, toDomainRichDocumentSummary(row))
	}
	return docs, nil
}
