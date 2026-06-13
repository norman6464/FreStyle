package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubLearningReportRepo struct {
	rows []domain.LearningReport
	err  error
}

func (s *stubLearningReportRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.LearningReport, error) {
	return s.rows, s.err
}

func (s *stubLearningReportRepo) Create(_ context.Context, r *domain.LearningReport) error {
	if s.err != nil {
		return s.err
	}
	r.ID = 51
	return nil
}

type stubEnqueuer struct{ err error }

func (s *stubEnqueuer) Enqueue(_ context.Context, _ uint64) error { return s.err }

func Test_学習レポート一覧_ユーザーIDが必須(t *testing.T) {
	uc := NewListLearningReportsUseCase(&stubLearningReportRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func Test_学習レポート要求_バリデーション(t *testing.T) {
	uc := NewRequestLearningReportUseCase(&stubLearningReportRepo{}, &stubEnqueuer{})
	now := time.Now()
	if _, err := uc.Execute(context.Background(), RequestLearningReportInput{UserID: 1, PeriodFrom: now, PeriodTo: now}); err == nil {
		t.Fatal("expected error for equal periods")
	}
}

func Test_学習レポート要求_正常系(t *testing.T) {
	uc := NewRequestLearningReportUseCase(&stubLearningReportRepo{}, &stubEnqueuer{})
	now := time.Now()
	got, err := uc.Execute(context.Background(), RequestLearningReportInput{
		UserID: 1, PeriodFrom: now.Add(-24 * time.Hour), PeriodTo: now,
	})
	if err != nil || got.Status != domain.LearningReportStatusPending {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func Test_学習レポート要求_enqueueエラー(t *testing.T) {
	uc := NewRequestLearningReportUseCase(&stubLearningReportRepo{}, &stubEnqueuer{err: errors.New("sqs")})
	now := time.Now()
	if _, err := uc.Execute(context.Background(), RequestLearningReportInput{
		UserID: 1, PeriodFrom: now.Add(-24 * time.Hour), PeriodTo: now,
	}); err == nil {
		t.Fatal("expected error")
	}
}
