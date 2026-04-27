package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubScoreGoalRepo struct {
	g   *domain.ScoreGoal
	err error
}

func (s *stubScoreGoalRepo) FindByUserID(_ context.Context, _ uint64) (*domain.ScoreGoal, error) {
	return s.g, s.err
}
func (s *stubScoreGoalRepo) Upsert(_ context.Context, g *domain.ScoreGoal) error {
	if s.err != nil {
		return s.err
	}
	s.g = g
	return nil
}

func TestGetScoreGoal_RequiresUserID(t *testing.T) {
	uc := NewGetScoreGoalUseCase(&stubScoreGoalRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertScoreGoal_RangeValidation(t *testing.T) {
	uc := NewUpsertScoreGoalUseCase(&stubScoreGoalRepo{})
	if _, err := uc.Execute(context.Background(), 1, -0.1); err == nil {
		t.Fatal("expected error for negative")
	}
	if _, err := uc.Execute(context.Background(), 1, 10.5); err == nil {
		t.Fatal("expected error for >10")
	}
}

func TestUpsertScoreGoal_OK(t *testing.T) {
	uc := NewUpsertScoreGoalUseCase(&stubScoreGoalRepo{})
	got, err := uc.Execute(context.Background(), 1, 8.5)
	if err != nil || got.TargetScore != 8.5 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
