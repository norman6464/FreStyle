package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CourseRepository はコースの永続化を担う。
type CourseRepository interface {
	// ListByWorkspaceID はワークスペース単位のコース一覧を返す。
	ListByWorkspaceID(ctx context.Context, workspaceID string, includeUnpublished bool) ([]domain.Course, error)
	GetByID(ctx context.Context, id uint64) (*domain.Course, error)
	Create(ctx context.Context, c *domain.Course) error
	Update(ctx context.Context, c *domain.Course) error
	Delete(ctx context.Context, id uint64) error
}
