package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// courseRepository は [repository.CourseRepository] の実装。
// 読み取りは生 SQL 直書き(db.Raw)、書き込み(Create/Update/Delete)は採番 ID・autoTime の
// 利便のため GORM を使う(ハイブリッド方針)。
type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) repository.CourseRepository {
	return &courseRepository{db: db}
}

// ListByCompany は自社のコースを sort_order 昇順で返す。includeUnpublished=false なら公開のみ。
func (r *courseRepository) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.Course, error) {
	const q = `
SELECT * FROM courses
WHERE company_id = ? AND (? OR is_published = TRUE)
ORDER BY sort_order ASC, id ASC`
	var rows []domain.Course
	if err := r.db.WithContext(ctx).Raw(q, companyID, includeUnpublished).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetByID は単一コースを返す。未存在は gorm.ErrRecordNotFound（handler が 404 に分岐）。
func (r *courseRepository) GetByID(ctx context.Context, id uint64) (*domain.Course, error) {
	var c domain.Course
	res := r.db.WithContext(ctx).Raw(`SELECT * FROM courses WHERE id = ?`, id).Scan(&c)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
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
