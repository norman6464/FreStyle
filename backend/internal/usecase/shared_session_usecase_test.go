package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubSharedRepo struct {
	rows []domain.SharedSession
	err  error
}

func (s *stubSharedRepo) ListPublic(_ context.Context, _ int) ([]domain.SharedSession, error) {
	return s.rows, s.err
}
func (s *stubSharedRepo) Create(_ context.Context, x *domain.SharedSession) error {
	if s.err != nil {
		return s.err
	}
	x.ID = 11
	return nil
}

func TestListSharedSessions(t *testing.T) {
	uc := NewListSharedSessionsUseCase(&stubSharedRepo{rows: []domain.SharedSession{{ID: 1}}})
	got, err := uc.Execute(context.Background(), 10)
	if err != nil || len(got) != 1 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestCreateSharedSession_Validates(t *testing.T) {
	uc := NewCreateSharedSessionUseCase(&stubSharedRepo{})
	if _, err := uc.Execute(context.Background(), CreateSharedSessionInput{OwnerID: 1, SessionID: 2}); err == nil {
		t.Fatal("expected error for empty title")
	}
	if _, err := uc.Execute(context.Background(), CreateSharedSessionInput{Title: "x"}); err == nil {
		t.Fatal("expected error for missing IDs")
	}
}

func TestCreateSharedSession_AssignsID(t *testing.T) {
	uc := NewCreateSharedSessionUseCase(&stubSharedRepo{})
	got, err := uc.Execute(context.Background(), CreateSharedSessionInput{OwnerID: 1, SessionID: 2, Title: "x"})
	if err != nil || got.ID != 11 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
