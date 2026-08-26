package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// learningReportRepository は [repository.LearningReportRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type learningReportRepository struct{ db *sql.DB }

func NewLearningReportRepository(db *sql.DB) repository.LearningReportRepository {
	return &learningReportRepository{db: db}
}

func toDomainLearningReport(row sqlcgen.LearningReport) domain.LearningReport {
	return domain.LearningReport{
		ID:         uint64(row.ID),
		UserID:     uint64(row.UserID),
		PeriodFrom: row.PeriodFrom,
		PeriodTo:   row.PeriodTo,
		Status:     row.Status,
		S3Key:      row.S3Key,
		CreatedAt:  row.CreatedAt,
	}
}

// ListByUserID は自分のレポートを期間末(period_to)降順で返す。
// period_to は同一期間のレポートが複数あれば同値になるため id をタイブレークに置く。
func (r *learningReportRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.LearningReport, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return []domain.LearningReport{}, nil // 存在し得ない user_id = 0 件
	}
	rows, err := sqlcgen.New(r.db).ListLearningReportsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.LearningReport, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainLearningReport(row))
	}
	return out, nil
}

func (r *learningReportRepository) Create(ctx context.Context, lr *domain.LearningReport) error {
	uid, ok := toInt64ID(lr.UserID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("user_id", lr.UserID)
	}
	createdAt := lr.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now() // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(r.db).InsertLearningReport(ctx, sqlcgen.InsertLearningReportParams{
		UserID:     uid,
		PeriodFrom: lr.PeriodFrom,
		PeriodTo:   lr.PeriodTo,
		Status:     lr.Status,
		S3Key:      lr.S3Key,
		CreatedAt:  createdAt,
	})
	if err != nil {
		return err
	}
	lr.ID = uint64(row.ID)
	lr.CreatedAt = row.CreatedAt
	return nil
}

// stubEnqueuer は [repository.SqsEnqueuer] の no-op 実装（本番の SQS 実装は別 PR）。
type stubEnqueuer struct{}

func NewStubSqsEnqueuer() repository.SqsEnqueuer { return &stubEnqueuer{} }

func (e *stubEnqueuer) Enqueue(_ context.Context, _ uint64) error { return nil }
