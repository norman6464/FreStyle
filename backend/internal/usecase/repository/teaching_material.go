package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// TeachingMaterialRepository は教材の永続化を担う。
//
// クエリは原則 company_id 指定で発行することで、 他社の教材が
// アプリ層のバグで漏れる事故を予防する（最後の防衛線として）。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// teachingMaterialRepository (GORM)。
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
