package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// courseRepository は [repository.CourseRepository] の GORM 実装。
type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) repository.CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.Course, error) {
	var rows []domain.Course
	q := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	if !includeUnpublished {
		q = q.Where("is_published = ?", true)
	}
	if err := q.Order("sort_order asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *courseRepository) GetByID(ctx context.Context, id uint64) (*domain.Course, error) {
	var c domain.Course
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *courseRepository) Create(ctx context.Context, c *domain.Course) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *courseRepository) Update(ctx context.Context, c *domain.Course) error {
	// CreatedBy / CompanyID は不変なので更新対象から外す。
	return r.db.WithContext(ctx).Model(c).Updates(map[string]any{
		"title":        c.Title,
		"description":  c.Description,
		"sort_order":   c.SortOrder,
		"is_published": c.IsPublished,
	}).Error
}

func (r *courseRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Course{}, id).Error
}
