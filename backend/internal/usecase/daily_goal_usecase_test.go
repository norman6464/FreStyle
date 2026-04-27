package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubDailyGoalRepo struct {
	g   *domain.DailyGoal
	err error
}

func (s *stubDailyGoalRepo) FindByUserAndDate(_ context.Context, _ uint64, _ time.Time) (*domain.DailyGoal, error) {
	return s.g, s.err
}
func (s *stubDailyGoalRepo) Upsert(_ context.Context, g *domain.DailyGoal) error {
	if s.err != nil {
		return s.err
	}
	s.g = g
	return nil
}

func TestGetDailyGoal_RequiresUserID(t *testing.T) {
	uc := NewGetDailyGoalUseCase(&stubDailyGoalRepo{})
	if _, err := uc.Execute(context.Background(), 0, time.Now()); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertDailyGoal_NegativeRejected(t *testing.T) {
	uc := NewUpsertDailyGoalUseCase(&stubDailyGoalRepo{})
	if _, err := uc.Execute(context.Background(), UpsertDailyGoalInput{UserID: 1, TargetMin: -1}); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertDailyGoal_AchievedFlag(t *testing.T) {
	uc := NewUpsertDailyGoalUseCase(&stubDailyGoalRepo{})
	got, err := uc.Execute(context.Background(), UpsertDailyGoalInput{UserID: 1, TargetMin: 30, ActualMin: 30})
	if err != nil || !got.IsAchieved {
		t.Fatalf("expected achieved, got %+v err=%v", got, err)
	}
}
