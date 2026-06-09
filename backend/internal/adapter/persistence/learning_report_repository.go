package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// learningReportRepository は [repository.LearningReportRepository] の実装。
// 読み取りは生 SQL 直書き(db.Raw)、書き込み(Create)は採番 ID・autoTime の利便のため GORM。
type learningReportRepository struct{ db *gorm.DB }

func NewLearningReportRepository(db *gorm.DB) repository.LearningReportRepository {
	return &learningReportRepository{db: db}
}

// ListByUserID は自分のレポートを期間末(period_to)降順で返す。
func (r *learningReportRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.LearningReport, error) {
	const q = `SELECT * FROM learning_reports WHERE user_id = ? ORDER BY period_to DESC`
	var rows []domain.LearningReport
	if err := r.db.WithContext(ctx).Raw(q, userID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *learningReportRepository) Create(ctx context.Context, lr *domain.LearningReport) error {
	return r.db.WithContext(ctx).Create(lr).Error
}

// stubEnqueuer は [repository.SqsEnqueuer] の no-op 実装（本番の SQS 実装は別 PR）。
type stubEnqueuer struct{}

func NewStubSqsEnqueuer() repository.SqsEnqueuer { return &stubEnqueuer{} }

func (e *stubEnqueuer) Enqueue(_ context.Context, _ uint64) error { return nil }
