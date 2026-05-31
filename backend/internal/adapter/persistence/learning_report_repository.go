package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// learningReportRepository は [repository.LearningReportRepository] の GORM 実装。
type learningReportRepository struct{ db *gorm.DB }

func NewLearningReportRepository(db *gorm.DB) repository.LearningReportRepository {
	return &learningReportRepository{db: db}
}

func (r *learningReportRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.LearningReport, error) {
	var rows []domain.LearningReport
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("period_to DESC").Find(&rows).Error
	return rows, err
}

func (r *learningReportRepository) Create(ctx context.Context, lr *domain.LearningReport) error {
	return r.db.WithContext(ctx).Create(lr).Error
}

// stubEnqueuer は [repository.SqsEnqueuer] の no-op 実装（本番の SQS 実装は別 PR）。
type stubEnqueuer struct{}

func NewStubSqsEnqueuer() repository.SqsEnqueuer { return &stubEnqueuer{} }

func (e *stubEnqueuer) Enqueue(_ context.Context, _ uint64) error { return nil }
