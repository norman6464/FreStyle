package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CourseRepository はコースの永続化を担う（クエリは company_id 指定で他社漏れを防ぐ）。
type CourseRepository interface {
	ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.Course, error)
	GetByID(ctx context.Context, id uint64) (*domain.Course, error)
	Create(ctx context.Context, c *domain.Course) error
	Update(ctx context.Context, c *domain.Course) error
	Delete(ctx context.Context, id uint64) error
}
