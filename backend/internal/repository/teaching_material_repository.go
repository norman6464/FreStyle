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
	ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error)
	GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error)
	Create(ctx context.Context, m *domain.TeachingMaterial) error
	Update(ctx context.Context, m *domain.TeachingMaterial) error
	Delete(ctx context.Context, id uint64) error
}

type teachingMaterialRepository struct {
	db *gorm.DB
}

func NewTeachingMaterialRepository(db *gorm.DB) TeachingMaterialRepository {
	return &teachingMaterialRepository{db: db}
}

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
	// 一部カラムのみ更新（CreatedBy / CompanyID は不変）。
	return r.db.WithContext(ctx).Model(m).Updates(map[string]any{
		"title":        m.Title,
		"content":      m.Content,
		"is_published": m.IsPublished,
	}).Error
}

func (r *teachingMaterialRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.TeachingMaterial{}, id).Error
}
