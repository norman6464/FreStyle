package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// TeachingMaterialRepository は教材の永続化を担う（クエリは company_id 指定で他社漏れを防ぐ）。
type TeachingMaterialRepository interface {
	// ListByCompany は company 内全教材を返す backward-compat 用（コース対応への移行後に削除予定）。
	ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error)
	ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error)
	GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error)
	// CountByCourseForCompany は company 内の教材件数を course_id ごとに一括集計して返す
	// (コース一覧に章数を出すための集計。コースごとの個別クエリだと N+1 になるため)。
	CountByCourseForCompany(ctx context.Context, companyID uint64, includeUnpublished bool) (map[uint64]int, error)
	Create(ctx context.Context, m *domain.TeachingMaterial) error
	Update(ctx context.Context, m *domain.TeachingMaterial) error
	Delete(ctx context.Context, id uint64) error
	DeleteByCourse(ctx context.Context, courseID uint64) error
}
