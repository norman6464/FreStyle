package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// courseRepository は [repository.CourseRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
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
	cid, ok := toInt64ID(c.CompanyID)
	if !ok {
		return nil // 存在し得ない company_id は書き込まない
	}
	createdBy, ok := toInt64ID(c.CreatedByUserID)
	if !ok {
		return nil // 存在し得ない created_by は書き込まない
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	now := time.Now()
	createdAt := c.CreatedAt
	if createdAt.IsZero() {
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := c.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(sqlDB).InsertCourse(ctx, sqlcgen.InsertCourseParams{
		CompanyID:       cid,
		CreatedByUserID: createdBy,
		Title:           c.Title,
		Description:     c.Description,
		Category:        c.Category,
		Language:        c.Language,
		SortOrder:       int32(c.SortOrder), // 0 は SQL 側の COALESCE で既定 100 に倒す
		IsPublished:     c.IsPublished,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	})
	if err != nil {
		return err
	}
	c.ID = uint64(row.ID)
	c.SortOrder = int(row.SortOrder) // 既定 100 が当たった場合を書き戻す（GORM の default タグ相当）
	c.CreatedAt = row.CreatedAt
	c.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *courseRepository) Update(ctx context.Context, c *domain.Course) error {
	id64, ok := toInt64ID(c.ID)
	if !ok {
		return nil // 存在し得ない id = 対象なし
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	// CreatedBy / CompanyID / Category / Language は更新対象外（GORM の Updates(map) と同じ 4 列のみ）。
	updatedAt, err := sqlcgen.New(sqlDB).UpdateCourse(ctx, sqlcgen.UpdateCourseParams{
		ID:          id64,
		Title:       c.Title,
		Description: c.Description,
		SortOrder:   int32(c.SortOrder),
		IsPublished: c.IsPublished,
	})
	if err != nil {
		return err
	}
	c.UpdatedAt = updatedAt // GORM Save 相当の書き戻し
	return nil
}

func (r *courseRepository) Delete(ctx context.Context, id uint64) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil // 存在し得ない id = 対象なし
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).DeleteCourse(ctx, id64)
}
