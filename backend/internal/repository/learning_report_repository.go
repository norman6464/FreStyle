package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type LearningReportRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.LearningReport, error)
	Create(ctx context.Context, r *domain.LearningReport) error
}

type learningReportRepository struct{ db *gorm.DB }

func NewLearningReportRepository(db *gorm.DB) LearningReportRepository {
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

// SqsEnqueuer はレポート生成ジョブを SQS にキューする interface。
// 実装は AWS SDK 連携で Phase 17.1 に分離。
type SqsEnqueuer interface {
	Enqueue(ctx context.Context, reportID uint64) error
}

type stubEnqueuer struct{}

func NewStubSqsEnqueuer() SqsEnqueuer { return &stubEnqueuer{} }

func (e *stubEnqueuer) Enqueue(_ context.Context, _ uint64) error { return nil }
