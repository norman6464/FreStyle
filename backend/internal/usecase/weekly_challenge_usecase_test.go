package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubWeeklyRepo struct {
	c   *domain.WeeklyChallenge
	err error
}

func (s *stubWeeklyRepo) CurrentChallenge(_ context.Context) (*domain.WeeklyChallenge, error) {
	return s.c, s.err
}
func (s *stubWeeklyRepo) UpsertProgress(_ context.Context, _ *domain.WeeklyChallengeProgress) error {
	return s.err
}

func TestGetCurrentWeeklyChallenge_PropagatesError(t *testing.T) {
	uc := NewGetCurrentWeeklyChallengeUseCase(&stubWeeklyRepo{err: errors.New("none")})
	if _, err := uc.Execute(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestCompleteWeeklyChallenge_RequiresIDs(t *testing.T) {
	uc := NewCompleteWeeklyChallengeUseCase(&stubWeeklyRepo{})
	if err := uc.Execute(context.Background(), 0, 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestCompleteWeeklyChallenge_OK(t *testing.T) {
	uc := NewCompleteWeeklyChallengeUseCase(&stubWeeklyRepo{})
	if err := uc.Execute(context.Background(), 1, 2); err != nil {
		t.Fatalf("err: %v", err)
	}
}
