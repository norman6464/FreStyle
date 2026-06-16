package persistence

import (
	"context"

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
SELECT * FROM teaching_materials
WHERE company_id = ? AND (? OR is_published = TRUE)
ORDER BY updated_at DESC, id DESC`
	var rows []domain.TeachingMaterial
	if err := r.db.WithContext(ctx).Raw(q, companyID, includeUnpublished).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListByCourse はコース内の教材を order_in_course 昇順で返す。
func (r *teachingMaterialRepository) ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	// 一覧は content(本文 markdown)を返さない（章ごとに重く、 全章を先読みすると非効率）。
	// 本文は選択時に GetByID で都度取得する。Content は空文字のままになる。
	const q = `
SELECT id, company_id, course_id, created_by_user_id, title, order_in_course, is_published, created_at, updated_at
FROM teaching_materials
WHERE course_id = ? AND (? OR is_published = TRUE)
ORDER BY order_in_course ASC, id ASC`
	var rows []domain.TeachingMaterial
	if err := r.db.WithContext(ctx).Raw(q, courseID, includeUnpublished).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetByID は単一教材を返す。未存在は gorm.ErrRecordNotFound（handler が 404 に分岐）。
func (r *teachingMaterialRepository) GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error) {
	var m domain.TeachingMaterial
	res := r.db.WithContext(ctx).Raw(`SELECT * FROM teaching_materials WHERE id = ?`, id).Scan(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &m, nil
}

func (r *teachingMaterialRepository) Create(ctx context.Context, m *domain.TeachingMaterial) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *teachingMaterialRepository) Update(ctx context.Context, m *domain.TeachingMaterial) error {
	// CreatedBy / CompanyID / CourseID は不変なので更新対象から外す。
	return r.db.WithContext(ctx).Model(m).Updates(map[string]any{
		"title":           m.Title,
		"content":         m.Content,
		"order_in_course": m.OrderInCourse,
		"is_published":    m.IsPublished,
	}).Error
}

func (r *teachingMaterialRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.TeachingMaterial{}, id).Error
}

// DeleteByCourse はコース削除時の cascade 用に配下教材を全削除する（FK に頼らず明示削除）。
func (r *teachingMaterialRepository) DeleteByCourse(ctx context.Context, courseID uint64) error {
	return r.db.WithContext(ctx).Where("course_id = ?", courseID).Delete(&domain.TeachingMaterial{}).Error
}
