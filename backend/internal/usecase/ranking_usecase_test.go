package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubRankingRepo struct {
	rows []domain.RankingEntry
	err  error
}

func (s *stubRankingRepo) TopByAverageScore(_ context.Context, _ int) ([]domain.RankingEntry, error) {
	return s.rows, s.err
}

func TestGetRanking_Returns(t *testing.T) {
	uc := NewGetRankingUseCase(&stubRankingRepo{rows: []domain.RankingEntry{{UserID: 1, Rank: 1}}})
	got, err := uc.Execute(context.Background(), 20)
	if err != nil || len(got) != 1 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestGetRanking_PropagatesError(t *testing.T) {
	uc := NewGetRankingUseCase(&stubRankingRepo{err: errors.New("db")})
	if _, err := uc.Execute(context.Background(), 20); err == nil {
		t.Fatal("expected error")
	}
}
