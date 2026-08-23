package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// teachingMaterialRepository は [repository.TeachingMaterialRepository] の実装。
// 読み取りは生 SQL 直書き(db.Raw)、書き込みは GORM(採番 ID・autoTime の利便)のハイブリッド。
type teachingMaterialRepository struct {
	db *gorm.DB
}

func NewTeachingMaterialRepository(db *gorm.DB) repository.TeachingMaterialRepository {
	return &teachingMaterialRepository{db: db}
}

// ListByCompany は backward-compat 用（コース対応完了後に削除予定）。
func (r *teachingMaterialRepository) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	const q = `
SELECT * FROM course_chapters
WHERE company_id = ? AND (? OR is_published = TRUE)
ORDER BY updated_at DESC, id DESC`
	rows := make([]domain.TeachingMaterial, 0)
	if err := r.db.WithContext(ctx).Raw(q, companyID, includeUnpublished).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListByCourse はコース内の章を sort_order 昇順で返す。
func (r *teachingMaterialRepository) ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	// 一覧は本文（doc・jsonb）を返さない（章ごとに重く、全章を先読みすると非効率）。
	// 本文は選択時に GetByID で都度取得する。Doc は nil のままになる。
	const q = `
SELECT id, company_id, course_id, created_by_user_id, title, sort_order, is_published, created_at, updated_at
FROM course_chapters
WHERE course_id = ? AND (? OR is_published = TRUE)
ORDER BY sort_order ASC, id ASC`
	rows := make([]domain.TeachingMaterial, 0)
	if err := r.db.WithContext(ctx).Raw(q, courseID, includeUnpublished).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetByID は単一教材を返す。未存在は gorm.ErrRecordNotFound（handler が 404 に分岐）。
func (r *teachingMaterialRepository) GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error) {
	var m domain.TeachingMaterial
	res := r.db.WithContext(ctx).Raw(`SELECT * FROM course_chapters WHERE id = ?`, id).Scan(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &m, nil
}

// CountByCourseForCompany は course_id ごとの教材件数を 1 クエリで集計する。
// trainee 向け(includeUnpublished=false)は published のみ数え、コース詳細の進捗分母と一致させる。
func (r *teachingMaterialRepository) CountByCourseForCompany(ctx context.Context, companyID uint64, includeUnpublished bool) (map[uint64]int, error) {
	const q = `
SELECT course_id, COUNT(*) AS cnt FROM course_chapters
WHERE company_id = ? AND (? OR is_published = TRUE)
GROUP BY course_id`
	var rows []struct {
		CourseID uint64
		Cnt      int
	}
	if err := r.db.WithContext(ctx).Raw(q, companyID, includeUnpublished).Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint64]int, len(rows))
	for _, row := range rows {
		counts[row.CourseID] = row.Cnt
	}
	return counts, nil
}

func (r *teachingMaterialRepository) Create(ctx context.Context, m *domain.TeachingMaterial) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *teachingMaterialRepository) Update(ctx context.Context, m *domain.TeachingMaterial) error {
	// CreatedBy / CompanyID / CourseID は不変なので更新対象から外す。
	return r.db.WithContext(ctx).Model(m).Updates(map[string]any{
		"title":        m.Title,
		"sort_order":   m.OrderInCourse,
		"is_published": m.IsPublished,
	}).Error
}

// UpdateDocWithRevision はリッチ本文（tiptap JSON）を revision 一致条件の楽観ロックで更新する。
// rich_documents の UpdateWithRevision と同じパターン（0 行更新は存在確認で 404/409 を切り分け）。
func (r *teachingMaterialRepository) UpdateDocWithRevision(ctx context.Context, id uint64, doc string, expectedRevision int) (*domain.TeachingMaterial, error) {
	res := r.db.WithContext(ctx).
		Model(&domain.TeachingMaterial{}).
		Where("id = ? AND revision = ?", id, expectedRevision).
		Updates(map[string]any{
			"doc":        doc,
			"revision":   gorm.Expr("revision + 1"),
			"updated_at": gorm.Expr("now()"),
		})
	if res.Error != nil {
		return nil, mapChapterDocError(res.Error)
	}
	if res.RowsAffected == 0 {
		// 0 行 = 「存在しない」か「版不一致」。存在確認で切り分ける。
		if _, err := r.GetByID(ctx, id); err != nil {
			return nil, err // gorm.ErrRecordNotFound
		}
		return nil, repository.ErrChapterDocConflict
	}
	return r.GetByID(ctx, id)
}

// mapChapterDocError は PostgreSQL の data exception（SQLSTATE class 22。jsonb に格納できない
// U+0000 等）を repository.ErrChapterDocInvalidData へ翻訳する（400 として返すため）。
func mapChapterDocError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, "22") {
		return repository.ErrChapterDocInvalidData
	}
	return err
}

func (r *teachingMaterialRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.TeachingMaterial{}, id).Error
}

// DeleteByCourse はコース削除時の cascade 用に配下教材を全削除する（FK に頼らず明示削除）。
func (r *teachingMaterialRepository) DeleteByCourse(ctx context.Context, courseID uint64) error {
	return r.db.WithContext(ctx).Where("course_id = ?", courseID).Delete(&domain.TeachingMaterial{}).Error
}
