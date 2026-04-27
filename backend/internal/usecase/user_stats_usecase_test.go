package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubUserStatsRepo struct {
	stats *domain.UserStats
	err   error
}

func (s *stubUserStatsRepo) Compute(_ context.Context, _ uint64) (*domain.UserStats, error) {
	return s.stats, s.err
}

func TestGetUserStats_RequiresUserID(t *testing.T) {
	uc := NewGetUserStatsUseCase(&stubUserStatsRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetUserStats_PropagatesError(t *testing.T) {
	uc := NewGetUserStatsUseCase(&stubUserStatsRepo{err: errors.New("db")})
	if _, err := uc.Execute(context.Background(), 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetUserStats_Returns(t *testing.T) {
	uc := NewGetUserStatsUseCase(&stubUserStatsRepo{stats: &domain.UserStats{UserID: 1, TotalSessions: 3}})
	got, err := uc.Execute(context.Background(), 1)
	if err != nil || got.TotalSessions != 3 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
