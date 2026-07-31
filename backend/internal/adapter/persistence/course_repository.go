package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// courseRepository は [repository.CourseRepository] の実装。
// 読み取りは sqlc 生成コード（生 SQL）、書き込み(Create/Update/Delete)は採番 ID・autoTime の
// 利便のため GORM を使う(ハイブリッド方針)。
type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) repository.CourseRepository {
	return &courseRepository{db: db}
}

func toDomainCourse(row sqlcgen.Course) domain.Course {
	return domain.Course{
		ID:              uint64(row.ID),
		CompanyID:       uint64(row.CompanyID),
		CreatedByUserID: uint64(row.CreatedByUserID),
		Title:           row.Title,
		Description:     row.Description,
		Category:        row.Category,
		Language:        row.Language,
		SortOrder:       int(row.SortOrder),
		IsPublished:     row.IsPublished,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

// ListByCompany は自社のコースを sort_order 昇順で返す。includeUnpublished=false なら公開のみ。
func (r *courseRepository) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.Course, error) {
	cid, ok := toInt64ID(companyID)
	if !ok {
		return []domain.Course{}, nil // 存在し得ない company_id = 0 件
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListCoursesByCompany(ctx, sqlcgen.ListCoursesByCompanyParams{
		CompanyID:          cid,
		IncludeUnpublished: includeUnpublished,
	})
	if err != nil {
		return nil, err
	}
	courses := make([]domain.Course, 0, len(rows))
	for _, row := range rows {
		courses = append(courses, toDomainCourse(row))
	}
	return courses, nil
}

// GetByID は単一コースを返す。未存在は gorm.ErrRecordNotFound（handler が 404 に分岐）。
func (r *courseRepository) GetByID(ctx context.Context, id uint64) (*domain.Course, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, gorm.ErrRecordNotFound // 存在し得ない id = not found
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetCourseByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorm.ErrRecordNotFound // 404 シグナルを維持
	}
	if err != nil {
		return nil, err
	}
	c := toDomainCourse(row)
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
