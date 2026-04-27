package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubScoreCardRepo struct {
	rows []domain.ScoreCard
	err  error
}

func (s *stubScoreCardRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.ScoreCard, error) {
	return s.rows, s.err
}
func (s *stubScoreCardRepo) Create(_ context.Context, c *domain.ScoreCard) error {
	if s.err != nil {
		return s.err
	}
	c.ID = 33
	return nil
}

func TestListScoreCards_RequiresUserID(t *testing.T) {
	uc := NewListScoreCardsByUserIDUseCase(&stubScoreCardRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateScoreCard_Validates(t *testing.T) {
	uc := NewCreateScoreCardUseCase(&stubScoreCardRepo{})
	if _, err := uc.Execute(context.Background(), &domain.ScoreCard{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateScoreCard_AssignsID(t *testing.T) {
	uc := NewCreateScoreCardUseCase(&stubScoreCardRepo{})
	got, err := uc.Execute(context.Background(), &domain.ScoreCard{UserID: 1, SessionID: 2})
	if err != nil || got.ID != 33 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
