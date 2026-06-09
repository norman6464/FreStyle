package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// LearningReportRepository は learning_reports テーブルへのアクセスを提供する。
type LearningReportRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.LearningReport, error)
	Create(ctx context.Context, r *domain.LearningReport) error
}

// SqsEnqueuer はレポート生成ジョブを SQS にキューする。
type SqsEnqueuer interface {
	Enqueue(ctx context.Context, reportID uint64) error
}
