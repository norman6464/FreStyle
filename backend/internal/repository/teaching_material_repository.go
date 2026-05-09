package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// TeachingMaterialRepository は教材の永続化を担う。
//
// クエリは原則 company_id 指定で発行することで、 他社の教材が
// アプリ層のバグで漏れる事故を予防する（最後の防衛線として）。
type TeachingMaterialRepository interface {
	// ListByCompany は backward-compat: 旧 GET /teaching-materials エンドポイント
	// 用に company 内全教材を返す。 frontend がコース対応に切り替わったら削除予定。
	ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error)
	ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error)
	GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error)
	Create(ctx context.Context, m *domain.TeachingMaterial) error
	Update(ctx context.Context, m *domain.TeachingMaterial) error
	Delete(ctx context.Context, id uint64) error
	DeleteByCourse(ctx context.Context, courseID uint64) error
}

type teachingMaterialRepository struct {
	db *gorm.DB
}

func NewTeachingMaterialRepository(db *gorm.DB) TeachingMaterialRepository {
	return &teachingMaterialRepository{db: db}
}

// ListByCompany は backward-compat 用。 frontend のコース対応完了後に削除予定。
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
	// コース内の並び順は order_in_course → id 昇順。
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
	// 一部カラムのみ更新（CreatedBy / CompanyID / CourseID は不変）。
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

// DeleteByCourse は指定 course に属する教材を全削除する。 コース削除時の
// cascade 用（FK の ON DELETE は GORM AutoMigrate で安定しないので明示的に消す）。
func (r *teachingMaterialRepository) DeleteByCourse(ctx context.Context, courseID uint64) error {
	return r.db.WithContext(ctx).Where("course_id = ?", courseID).Delete(&domain.TeachingMaterial{}).Error
}
