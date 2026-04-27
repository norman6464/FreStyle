package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubScoreTrendRepo struct {
	pts []domain.ScoreTrendPoint
	err error
}

func (s *stubScoreTrendRepo) AggregateDaily(_ context.Context, _ uint64, _ int) ([]domain.ScoreTrendPoint, error) {
	return s.pts, s.err
}

func TestGetScoreTrend_RequiresUserID(t *testing.T) {
	uc := NewGetScoreTrendUseCase(&stubScoreTrendRepo{})
	if _, err := uc.Execute(context.Background(), 0, 30); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetScoreTrend_Wraps(t *testing.T) {
	uc := NewGetScoreTrendUseCase(&stubScoreTrendRepo{
		pts: []domain.ScoreTrendPoint{{Date: "2026-04-26", OverallScore: 7.5}},
	})
	got, err := uc.Execute(context.Background(), 1, 30)
	if err != nil || len(got.Points) != 1 || got.UserID != 1 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
