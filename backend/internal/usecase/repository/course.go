package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CourseRepository はコースの永続化を担う。
//
// 教材と同様、 クエリは原則 company_id 指定で発行することで、 他社のコースが
// アプリ層のバグで漏れる事故を予防する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// courseRepository (GORM)。
type CourseRepository interface {
	ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.Course, error)
	GetByID(ctx context.Context, id uint64) (*domain.Course, error)
	Create(ctx context.Context, c *domain.Course) error
	Update(ctx context.Context, c *domain.Course) error
	Delete(ctx context.Context, id uint64) error
}
