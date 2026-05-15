package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// LearningReportRepository は learning_reports テーブル へ の アクセス を 提供 する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// learningReportRepository (GORM)。
type LearningReportRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.LearningReport, error)
	Create(ctx context.Context, r *domain.LearningReport) error
}

// SqsEnqueuer はレポート生成ジョブを SQS にキューする interface。
//
// 単一 メソッド の port な ので Effective Go 流 の -er 命名 を 採用 (multi-method
// の Repository とは ガイド ライン が 異なる)。 stub 実装 は
// [github.com/norman6464/FreStyle/backend/internal/adapter/persistence].NewStubSqsEnqueuer。
type SqsEnqueuer interface {
	Enqueue(ctx context.Context, reportID uint64) error
}
