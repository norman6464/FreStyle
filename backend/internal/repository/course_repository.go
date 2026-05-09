package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// CourseRepository はコースの永続化を担う。
//
// 教材と同様、 クエリは原則 company_id 指定で発行することで、 他社のコースが
// アプリ層のバグで漏れる事故を予防する。
type CourseRepository interface {
	ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.Course, error)
	GetByID(ctx context.Context, id uint64) (*domain.Course, error)
	Create(ctx context.Context, c *domain.Course) error
	Update(ctx context.Context, c *domain.Course) error
	Delete(ctx context.Context, id uint64) error
}

type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) CourseRepository {
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
	// 一部カラムのみ更新（CreatedBy / CompanyID は不変）。
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
