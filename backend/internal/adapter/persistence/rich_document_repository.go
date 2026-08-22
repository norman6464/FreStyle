package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
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

// richDocumentRepository は [repository.RichDocumentRepository] の GORM 実装。
// doc は jsonb 列に保存し、更新は revision 一致を条件にした楽観ロックで行う。
type richDocumentRepository struct{ db *gorm.DB }

// NewRichDocumentRepository は rich_documents の repository を組み立てる。
func NewRichDocumentRepository(db *gorm.DB) repository.RichDocumentRepository {
	return &richDocumentRepository{db: db}
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
	return mapPgDataError(r.db.WithContext(ctx).Create(doc).Error)
}

func (r *richDocumentRepository) FindByID(ctx context.Context, id string) (*domain.RichDocument, error) {
	var doc domain.RichDocument
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&doc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRichDocumentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *richDocumentRepository) UpdateWithRevision(ctx context.Context, doc *domain.RichDocument, expectedRevision int) error {
	res := r.db.WithContext(ctx).
		Model(&domain.RichDocument{}).
		Where("id = ? AND revision = ? AND deleted_at IS NULL", doc.ID, expectedRevision).
		Updates(map[string]any{
			"title":          doc.Title,
			"is_public":      doc.IsPublic,
			"schema_version": doc.SchemaVersion,
			"doc":            doc.Doc,
			"revision":       gorm.Expr("revision + 1"),
			"updated_at":     gorm.Expr("now()"),
		})
	if res.Error != nil {
		return mapPgDataError(res.Error)
	}
	if res.RowsAffected == 0 {
		// 0 行 = 「存在しない/論理削除済み」か「版不一致」。存在確認で切り分ける。
		if _, err := r.FindByID(ctx, doc.ID); err != nil {
			return err // ErrRichDocumentNotFound
		}
		return repository.ErrRichDocumentConflict
	}
	// 更新後の正確な行（revision / updated_at など）を読み直して doc に反映する。
	fresh, err := r.FindByID(ctx, doc.ID)
	if err != nil {
		return err
	}
	*doc = *fresh
	return nil
}

func (r *richDocumentRepository) SoftDelete(ctx context.Context, id string, ownerID uint64) error {
	res := r.db.WithContext(ctx).
		Model(&domain.RichDocument{}).
		Where("id = ? AND owner_id = ? AND deleted_at IS NULL", id, ownerID).
		Update("deleted_at", gorm.Expr("now()"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return repository.ErrRichDocumentNotFound
	}
	return nil
}

func (r *richDocumentRepository) ListByOwner(ctx context.Context, ownerID uint64, kind domain.DocumentKind) ([]domain.RichDocument, error) {
	q := r.db.WithContext(ctx).
		Model(&domain.RichDocument{}).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		// doc(jsonb) 本体は一覧では要らないので読み込まない（転送量とメモリを抑える）。
		Select("id, owner_id, company_id, kind, title, is_public, schema_version, revision, created_at, updated_at").
		Order("updated_at DESC")
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	// 0 件でも nil ではなく空スライスを返す（JSON が null にならずフロントの map/for-of が落ちない）。
	rows := make([]domain.RichDocument, 0)
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
