package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// teachingMaterialRepository は [repository.TeachingMaterialRepository] の GORM 実装。
type teachingMaterialRepository struct {
	db *gorm.DB
}

func NewTeachingMaterialRepository(db *gorm.DB) repository.TeachingMaterialRepository {
	return &teachingMaterialRepository{db: db}
}

// ListByCompany は backward-compat 用（コース対応完了後に削除予定）。
func (r *teachingMaterialRepository) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	var rows []domain.TeachingMaterial
	q := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	if !includeUnpublished {
		q = q.Where("is_published = ?", true)
	}
	if err := q.Order("updated_at desc, id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *teachingMaterialRepository) ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	var rows []domain.TeachingMaterial
	q := r.db.WithContext(ctx).Where("course_id = ?", courseID)
	if !includeUnpublished {
		q = q.Where("is_published = ?", true)
	}
	if err := q.Order("order_in_course asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *teachingMaterialRepository) GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error) {
	var m domain.TeachingMaterial
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
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
